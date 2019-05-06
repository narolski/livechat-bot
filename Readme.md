# LiveChat Bot Integration Example in Go

Have you ever wanted to write your own LiveChat integration, but didn't know where to start? Well, you've found the right place!

By following this tutorial you will create your very own LiveChat Integration - a basic bot that will respond to a predefined trigger by sending its' own custom message. How cool is that?

## What you will learn from this tutorial

You will learn how to:
1. Obtain the access token used to authenticate API calls from OAuth2-based LiveChat SSO Server
2. Create a custom Bot Agent using the LiveChat Configuration API
3. Set-up a websocket-based connection with the LiveChat Real-Time Messaging API to receive and send events to the chats which you have the access to

## Getting started

To start, [create a LiveChat account](https://www.livechatinc.com/signup/?source_id=header_cta&source_url=https://www.livechatinc.com/&source_type=website) if you don't already have one.

You will have to set up a new app in the [LiveChat Developer Console](https://developers.livechatinc.com/console/apps). Click a *New App* button, enter a custom name and select the *Server-side Webhook App* template.

You will be redirected to a page where a *Client ID* and *Client Secret* are shown. Write them down - we will need them later.

![app authorization](/readme/1.png "App authorization")

You will also need to add `http://localhost:8000` if you will be running the app locally to the *Redirect URL whitelist* and `agents--all:rw` and `chats--all:rw` to *App scopes and API access*. 

Make sure that all the changes have been saved successfully before proceeding.

## Getting an access token needed to validate API calls

We will now obtain an access token which will be used to set up a websocket-based connection with the LiveChat RTM API using the [Go OAuth2](https://github.com/golang/oauth2) library.

Clone the repo and open the `oauth/livechat.go` file in your favourite editor. You will have to make sure that in the `LiveChatOauthConfig` variable there is a correct `ClientID` and `ClientSecret` provided, as well as appropriate `RedirectURL` (leave `http://localhost:8000/callback` if running locally).

```Go
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
```

The code from `oauth/livechat.go` will, upon execution:
1. Redirect you to the login URL, where LiveChat login credentials should be passed
2. Obtain the *authentication code* from the LiveChat's SSO
3. Exchange the *authentication code* into the *access token* and *refresh token*

The first part is handled by the `OAuthLiveChatLogin` method, using the configuration values defined previously in `LiveChatOAuthConfig`:

```Go
// OAuthLiveChatLogin handles redirection to the LiveChat's login page where an authorization code is generated
func OAuthLiveChatLogin(w http.ResponseWriter, r *http.Request) {

	// Creates OAuth2 state cookie which is used to protect against the CSRF attacks
	oauthState := generateStateOAuthCookie(w)

	// Creates an URL to which the redirection will be performed
	url := LiveChatOAuthConfig.AuthCodeURL(oauthState, oauth2.AccessTypeOnline)

	log.Printf("Handling login. The AuthCodeURL is: %s", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
```
`OAuthLiveChatLogin` will perform a redirection to the LiveChat's login page, where you might be prompted to provide your login credentials to the LiveChat account. It generates an `oauthState` cookie through the `generateStateOAuthCookie` method, containing the `state` variable used to protect against *CSRF attacks*.

The callback from the login page is then handled by the `OAuthLiveChatCallback` method:

```Go
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
```
`OAuthLiveChatCallback` first verifies whether the OAuth `state` value returned by the server is the same as the one we've previously generated. If correct, an *access token* is generated from the *authorization code* returned by the LiveChat's login page.

The *access token* is stored internally and will be used later to make requests.

A handy method called `GetLiveChatAPIToken` will make sure that we always have a valid *access token* by obtaining a new, valid one should the currently held token expire.

You may start the sample app's webserver by entering 
```
go run main.go
``` 
in the terminal. You will see a short introduction:

![App first run](/readme/3.png "App first run")

Click the *Login via LiveChat and Run* button and proceed to the next part of this tutorial.


## Creating a Bot Agent using the Configuration API

After obtaining the *access token* we will be able to make a request to create a new, custom *Bot Agent*. This request will be made to the LiveChat's [Configuration API](https://developers.livechatinc.com/beta-docs/configuration-api/).

The previously generated *access token* needed to create a *Bot Agent* will be shown by the running web app:

![authorization token](/readme/2.png "Authorization token")

To create a *Bot Agent*, send a `POST` request to `api.livechatinc.com/configuration/agents/create_bot_agent`. An example request made via `cURL`:

```cURL
curl -X POST \
  https://api.livechatinc.com/configuration/agents/create_bot_agent \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <PUT THE TOKEN HERE>' \
  -H 'cache-control: no-cache' \
  -d '{
    "name": "<YOUR BOT NAME>",
    "status": "not accepting chats"
}'
```

Make sure that the `content-type` is set to `application/json` and the correct *access token* is set in a request's header.

You can see the detailed documentation and payload examples at the [ Configuration API](https://developers.livechatinc.com/beta-docs/configuration-api/#bot-agent) documentation page. For our purpouses, a request body can be set to:

```
{
    "name": "<YOUR BOT NAME>",
    "status": "not accepting chats"
}
```
Setting the bot's `status` to `not accepting chats` means that the LiveChat's chat router will not treat is as one of the agents logged in to the account and will not pass the messages from the customers to the bot to respond. It will, however, allow us to join chats and write new messages to them.

If everything goes well, you will get a response with the created *Bot Agent's* `ID`:
```
{
    "bot_agent_id": "<YOUR BOT AGENT ID>"
}
```
Set the *Bot Agent* `ID` returned by the *Configuration API* as a value of a `botAgentID` variable defined in `bot/bot.go`.

Don't click that shiny *Run the Bot* button just yet, though!

## Creating a Bot

It's high time we coded our bot! To make it work smoothly, we will use the LiveChat [Real Time Messaging API](https://developers.livechatinc.com/beta-docs/agent-chat-api/#real-time-messaging-api) which is a part of the Agent Chat API. 