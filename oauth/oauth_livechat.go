package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

var LiveChatToken *oauth2.Token

// Scopes: OAuth 2.0 scopes provide a way to limit the amount of access that is granted to an access token.
var LiveChatOauthConfig = &oauth2.Config{
	RedirectURL:  "http://e5fc0b6d.ngrok.io/callback",
	ClientID:     "e65988fbff37b8cf03d661d4976fd213",
	ClientSecret: "99e5d80080b1fc823fe5a852ed006ce8",
	// Scopes:       []string{"https://www.LiveChatapis.com/auth/userinfo.email"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://accounts.livechatinc.com/",
		TokenURL: "https://accounts.livechatinc.com/token",
	},
	Scopes: []string{"chats--all:rw"},
}

const oauthLiveChatUrlAPI = "https://accounts.livechatinc.com/token"

func OauthLiveChatLogin(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Handling the login...")

	// Create oauthState cookie (random state)
	oauthState := generateStateOauthCookie(w)

	/*
		AuthCodeURL receive state that is a token to protect the user from CSRF attacks. You must always provide a non-empty string and
		validate that it matches the the state query parameter on your redirect callback.
	*/
	url := LiveChatOauthConfig.AuthCodeURL(oauthState, oauth2.AccessTypeOnline)

	fmt.Println("The AuthCodeURL is: ", url)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func OauthLiveChatCallback(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Handling the callback...")

	// Read oauthState from Cookie
	oauthState, err := r.Cookie("oauthstate")

	// Read oauthstate not from cookie
	// oauthState := r.FormValue("oauthstate")

	// fmt.Println("oauthstate is: ", oauthState)

	if err != nil {
		fmt.Println("Error when reading oauth state: ", err)
	}

	// Verify that the state value is correct
	if r.FormValue("state") != oauthState.Value {
		log.Println("invalid oauth LiveChat state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Swap the code for the access token
	tokenData, err := getCredentialsFromCode(r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Update the token
	LiveChatToken = tokenData

	// GetOrCreate User in your db.
	// Redirect or response with a token.
	// More code .....
	fmt.Fprintf(w, "UserInfo: %s\n", tokenData)
}

func getCredentialsFromCode(code string) (*oauth2.Token, error) {
	// Use the code to get the access_token and refresh_token from LiveChat.

	// Converts an authorization code into a access_token
	token, err := LiveChatOauthConfig.Exchange(context.Background(), code)

	if err != nil {
		return nil, fmt.Errorf("Code exchange error: %s", err.Error())
	}

	return token, nil
}

// GetHTTPClient refreshes an access token if needed and returns a http.Client used for requests
func LiveChatAPIClient() *http.Client {
	tokenSource := LiveChatOauthConfig.TokenSource(oauth2.NoContext, LiveChatToken)
	newToken, err := tokenSource.Token()

	if err != nil {
		log.Fatalln(err)
	}

	if newToken.AccessToken != LiveChatToken.AccessToken {
		// Update the token
		LiveChatToken = newToken
	}

	return oauth2.NewClient(oauth2.NoContext, tokenSource)
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(20 * time.Minute)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}
