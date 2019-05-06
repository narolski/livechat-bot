package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"path"
	"text/template"
	"time"

	"golang.org/x/oauth2"
)

// Stores the OAuth2 configuration
var LiveChatOAuthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8000/callback",
	ClientID:     "<YOUR CLIENT ID>",
	ClientSecret: "<YOUR CLIENT SECRET>",
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://accounts.livechatinc.com/",
		TokenURL: "https://accounts.livechatinc.com/token",
	},
	Scopes: []string{"chats--all:rw", "agents-bot--all:rw"},
}

// Stores the valid authentication token
var LiveChatToken *oauth2.Token

// OAuthLiveChatLogin handles redirection to the LiveChat's login page where an authorization code is generated
func OAuthLiveChatLogin(w http.ResponseWriter, r *http.Request) {

	// Creates OAuth2 state cookie which is used to protect against the CSRF attacks
	oauthState := generateStateOAuthCookie(w)

	// Creates an URL to which the redirection will be performed
	url := LiveChatOAuthConfig.AuthCodeURL(oauthState, oauth2.AccessTypeOnline)

	log.Printf("Handling login. The AuthCodeURL is: %s", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// OAuthLiveChatCallback handles the response from the LiveChat's SSO containing the access token
// It converts the received authorization code to an access token, which will be used to authenticate the API calls
func OAuthLiveChatCallback(w http.ResponseWriter, r *http.Request) {

	// Get the OAuth2 state value
	oauthState, err := r.Cookie("oauthstate")

	if err != nil {
		log.Fatalln("Error when reading oauth state: ", err.Error())
	}

	// Verify the state value
	if r.FormValue("state") != oauthState.Value {
		log.Println("Invalid OAuth2 state value. Redirecting to login page...")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Use the authorization code to get the access token
	tokenData, err := getAccessTokenFromAuthorizatonCode(r.FormValue("code"))
	if err != nil {
		log.Printf("Error when getting the access token: %s. Redirecting to login page...", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Set the authentication token
	setLiveChatToken(tokenData)

	// Redirect to the "bot" page
	http.Redirect(w, r, "/token", http.StatusTemporaryRedirect)
	return
}

// getAccessTokenFromAuthorizatonCode converts the received authentication code into a access token and a refresh token used to obtain a new, valid pair should an old one pass its' expiration period
func getAccessTokenFromAuthorizatonCode(code string) (*oauth2.Token, error) {
	// Converts an authorization code into a access_token
	token, err := LiveChatOAuthConfig.Exchange(context.Background(), code)

	if err != nil {
		log.Fatalln("Code exchange error:", err.Error())
	}

	return token, nil
}

// getLiveChatToken returns the currently held access token
func getLiveChatToken() *oauth2.Token {
	return LiveChatToken
}

// setLiveChatToken updates the currently held access token to a new value
func setLiveChatToken(token *oauth2.Token) {
	LiveChatToken = token
}

// HasLiveChatToken returns whether the access token was obtained or not
func HasLiveChatToken() bool {
	if getLiveChatToken() != nil {
		return true
	}
	return false
}

// GetLiveChatAPIToken returns a valid access token, performing an access token refresh if necessary
func GetLiveChatAPIToken() *oauth2.Token {
	tokenSource := LiveChatOAuthConfig.TokenSource(oauth2.NoContext, getLiveChatToken())
	newToken, err := tokenSource.Token()

	if err != nil {
		log.Fatalln("Unable to validate/refresh the token:", err.Error())
	}

	if newToken.AccessToken != LiveChatToken.AccessToken {
		// Update the token value stored
		setLiveChatToken(newToken)
	}

	return LiveChatToken
}

// generateStateOauthCookie generates a random state OAuth cookie
// Note that this might not be secure and should be used only for the testing purpouses
func generateStateOAuthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(20 * time.Minute)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

// ShowOAuthToken renders a page with a current valid oauth token
func ShowOAuthToken(w http.ResponseWriter, r *http.Request) {
	type pageData struct {
		AccessToken string
	}

	if !HasLiveChatToken() {
		http.Error(w, "No token available", http.StatusInternalServerError)
		return
	}

	data := pageData{
		AccessToken: getLiveChatToken().AccessToken,
	}

	lp := path.Join("templates", "token.html")
	tmpl, err := template.ParseFiles(lp)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
