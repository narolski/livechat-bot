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

![alt text](/readme/1.png "Logo Title Text 1")

You will also need to add `http://localhost:8000` if you will be running the app locally to the *Redirect URL whitelist* and `agents--all:rw` and `chats--all:rw` to *App scopes and API access*. 

Make sure that all the changes have been saved successfully before proceeding.

## Getting an access token needed to validate API calls

We will now obtain an access token which will be used to set up a websocket-based connection with the LiveChat RTM API using the [Go OAuth2](https://github.com/golang/oauth2) library.

Clone the repo and open the `oauth/livechat.go` file in your favourite editor. You will have to make sure that in the `LiveChatOauthConfig` variable there is a correct `ClientID` and `ClientSecret` provided, as well as appropriate `RedirectURL` (leave `http://localhost:8000/callback` if running locally).

```Go
var LiveChatOauthConfig = &oauth2.Config{
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

A handy method called `GetLiveChatAPIToken` will make sure that we always have a valid access token by obtaining a new, valid pair when the old one expires.

## Creating a Bot Agent using the Configuration API

After obtaining the *access token* we will be able to make a request to create a new, custom *Bot Agent* to the LiveChat Configuration API.

To do that, send a `POST` request to `api.livechatinc.com/configuration/agents/create_bot_agent`. The authorization method should be set to a `Bearer Token` with obtained *access token* as its' value, and the `Content-Type` should be `application/json`.

