# LiveChat Bot Integration Example in Go

Have you ever wanted to write your own LiveChat integration, but didn't know where to start? Well, you've found the right place!

![sample](/readme/5.png "sample")

By following this tutorial you will create your very own LiveChat integration - a basic bot that will respond to a predefined trigger by sending its' own custom message. How cool is that?

## What you will learn from this tutorial

You will learn how to:
1. Obtain the access token used to authenticate API calls from OAuth2-based *LiveChat SSO Server*
2. Create a custom Bot Agent using the *LiveChat Configuration API*
3. Set-up a websocket-based connection with the *LiveChat Real-Time Messaging API *to receive and send events to the chats which you have the access to

It should take you no more than **30 minutes** to go through this entire tutorial.

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

Make sure that no one unauthorized has access to this token!

To create a *Bot Agent*, send a `POST` request to `api.livechatinc.com/configuration/agents/create_bot_agent`. An example request made via `cURL`:

```JSON
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

```json
{
    "name": "<YOUR BOT NAME>",
    "status": "not accepting chats"
}
```
Setting the bot's `status` to `not accepting chats` means that the LiveChat's chat router will not treat is as one of the agents logged in to the account and will not pass the messages from the customers to the bot to respond. It will, however, allow us to join chats and write new messages to them.

If everything goes well, you will get a response with the created *Bot Agent's* `ID`:
```json
{
    "bot_agent_id": "<YOUR BOT AGENT ID>"
}
```
Set the *Bot Agent* `ID` returned by the *Configuration API* as a value of a `botAgentID` variable defined in `bot/bot.go`.

Don't click that shiny *Run the Bot* button just yet, though!

## Creating a Bot

It's high time we coded our bot! To make it work smoothly, we will use the LiveChat [Real-Time Messaging API](https://developers.livechatinc.com/beta-docs/agent-chat-api/#real-time-messaging-api) which is a part of the *Agent Chat API*. 

The *Real-Time Messaging API* is based on a websocket connection. We will use it to read all messages arriving to our account. 

The `bot` package is divided into two files: `api.go` and `bot.go`. In the first one you will find methods responsible for running and maintaining the connection to the *RTM API*, as well as sending requests. The second one defines the bot's logic.

### Connection to the API

A websocket connection will be handled through the `gorilla/websocket` library. 

We start by defining the `agentChatAPIURL` and a `pingInterval`, which the *RTM API* requires to be no longer than once every 15 seconds:

```Go
const (
	agentChatAPIURL string = "wss://api.livechatinc.com/v3.0/agent/rtm/ws"
	pingInterval time.Duration = 15 * time.Second
)
```

Then an `apiConnect` method is defined, which is responsible for creating a connection and handling the incoming events from the *RTM API*:

```Go
// connect creates and maintains the websocket connection with LiveChat's RTM API
func apiConnect() {
	// Set up the connection
	c, _, err := websocket.DefaultDialer.Dial(agentChatAPIURL, nil)

	if err != nil {
		log.Fatalln("Error when dialing websocket: ", err)
	}
	defer c.Close()

	// Start the pinger, which will ping the service at a set interval in order to keep the connection open
	go apiPinger(c)

	// Login to the API through access token
	if err := apiLogin(c); err != nil {
		log.Fatalln("Error when authenticating: ", err)
	}

	for {
		_, raw, err := c.ReadMessage()
		if err != nil {
			log.Fatalf("API read message error: %s", err)
		}

		if err := apiHandleMessage(c, raw); err != nil {
			log.Fatalf("API handle message error: %s", err)
		}
	}
}
```

It also starts a `apiPinger` method concurrently, which will make sure the connection is not closed prematurely:

```Go
// pinger pings the server at a given time interval in order to maintain the websocket connection
func apiPinger(c *websocket.Conn) {
	t := time.NewTimer(pingInterval)

	for {
		<-t.C
		c.WriteMessage(websocket.PingMessage, []byte{})
		t.Reset(pingInterval)
	}
}
```

The `apiLogin` method handles the authentication using the previously obtained *access token*:

```Go
// apiLogin handles the user authentication using the generated token k
func apiLogin(c *websocket.Conn) error {
	type loginRequest struct {
		Token string `json:"token"`
	}

	// Get the access token
	token := "Bearer " + oauth.GetLiveChatAPIToken().AccessToken

	payload := &loginRequest{
		Token: token,
	}
	return apiSendRequest(c, "login", false, payload)
}
```
After a successful authentication the `apiConnect` enters a loop, reading incoming messages and passing them to the `apiHandleMessageMethod`:

```Go
// apiHandleMessage reads the details of the protocol responses and performs defined actions
func apiHandleMessage(c *websocket.Conn, raw []byte) error {
	type protocolResponse struct {
		RequestID string          `json:"request_id,omitempty"`
		Action    string          `json:"action"`
		Type      string          `json:"type"`
		Payload   json.RawMessage `json:"payload"`
		Success   *bool           `json:"success"`
	}

	log.Printf("API Message received: %s", raw)

	msg := &protocolResponse{}
	if err := json.Unmarshal(raw, msg); err != nil {
		return err
	}

	// log.Printf("API message understood as: %s", msg)

	if msg.Success != nil && !*msg.Success {
		return errors.New(fmt.Sprintf("Message %s failed", msg.Action))
	}

	// Handle different API message types through the response's action
	switch msg.Action {
	// Add more cases here
	case "incoming_event":
		return botHandleIncomingEvent(c, msg.Payload)
	}

	return nil
}
```
`apiHandleMessage` reads the basic details about the message from *RTM API*, such as its' `action` parameter and calls bot to perform actions appropriately. 

Here, upon the arrival of `incoming_event` message the bot's `botHandleIncomingEvent` method is invoked with the message's payload passed.

Finally, `apiSendRequest` method receives the `payload` from the bot and makes a request to the *RTM API*:

```Go
// apiSendRequest makes the request to the API
func apiSendRequest(c *websocket.Conn, action string, asBot bool, payload interface{}) error {
	type protocolRequest struct {
		Action    string      `json:"action"`
		RequestID string      `json:"request_id"`
		Payload   interface{} `json:"payload"`
		AuthorID  string      `json:"author_id"`
	}

	msg := protocolRequest{
		Action:    action,
		RequestID: strconv.Itoa(rand.Int()),
		Payload:   payload,
	}

	if asBot {
		// Send the message as a bot agent instead of the authenticated user's agent
		msg.AuthorID = botAgentID
	}

	raw, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if err := c.WriteMessage(websocket.TextMessage, raw); err != nil {
		return err
	}

	log.Printf("API message sent: %s", raw)
	return nil
}

```

### Defining Bot Actions

The bot actions are defined by methods found in `bot/bot.go` file.

To start with, check whether the `botAgentID` contains a correct *Bot Agent ID*. Then, define your custom bot `triggerWord` and `botResponse`:

```Go
const (
	botAgentID  string = "<BOT AGENT ID>"
	triggerWord string = "pizza" // Sample trigger, response
	botResponse string = "The pizza is on its' way!"
)
```

The bot works like this:
- upon the arrival of the `incoming_event` (an object containing the message) check if the message has the `triggerWord` in it
- if `triggerWord` is found in a message, respond to the current thread by sending a message with predefined `botResponse`

Bot is started through the `StartBotAgent` method which checks for the presence of the *access token*, runs the `apiConnect` concurrently and renders the status page below:

![status](/readme/4.png "status page")

```Go
// StartBotAgent runs the bot agent by creating a websocket-based connection with LiveChat's Real Time Messaging API, listens for incoming messages and sends responses based on their content
func StartBotAgent(w http.ResponseWriter, r *http.Request) {

	// Verify that the access token is present
	if !oauth.HasLiveChatToken() {
		http.Error(w, "No token present", http.StatusInternalServerError)
		return
	}

	lp := path.Join("templates", "bot.html")
	tmpl, err := template.ParseFiles(lp)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Run the websocket-based connection to the API
	go apiConnect()

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
```

Upon an incoming event arrival the method `botHandleIncomingEvent` checks whether the event is of type message and if so, passes it to `botHandleIncomingMessage` method:

```Go
// handleIncomingEvent performs appropriate actions according to the received event's details
func botHandleIncomingEvent(c *websocket.Conn, raw []byte) error {
	type incomingEventResponse struct {
		ChatID   string `json:"chat_id"`
		ThreadID string `json:"thread_id"`
		Event    struct {
			ID         string `json:"id"`
			Order      int    `json:"order"`
			Timestamp  int    `json:"timestamp"`
			Recipients string `json:"recipients"`
			Type       string `json:"type"`
			Text       string `json:"text"`
			AuthorID   string `json:"author_id"`
		} `json:"event"`
	}

	// Parse the event details
	response := &incomingEventResponse{}
	json.Unmarshal([]byte(raw), response)

	switch response.Event.Type {
	// Add more cases here
	case "message":
		log.Println("Handling the incoming message:", response.Event.Text)
		return botHandleIncomingMessage(c, response.Event.Text, response.ChatID, response.Event.AuthorID)
	}

	return nil
}
```

`botHandleIncomingMessage` performs a `triggerWord` check through use of a `strings.Contains` method. 

When a trigger word is present in the message, `botSendChatMessage` method is called with the `botResponse` passed as a message contents, and `chatID` as the ID of chat where the message should be sent:

```Go
// botSendChatMessage sends the message string to a chat with a given
func botSendChatMessage(c *websocket.Conn, chatID string, message string) error {
	type event struct {
		Type       string `json:"type"`
		Text       string `json:"text"`
		Recipients string `json:"recipients"`
		AuthorID   string `json:"author_id"`
	}

	type sendEventRequest struct {
		ChatID             string `json:"chat_id"`
		AttachToLastThread bool   `json:"attach_to_last_thread"`
		Event              *event `json:"event"`
	}

	payload := &sendEventRequest{
		ChatID:             chatID,
		AttachToLastThread: true,
		Event: &event{
			Type:       "message",
			Text:       message,
			Recipients: "all",
		},
	}
	return apiSendRequest(c, "send_event", true, payload)
}
```

The sent message is automatically attached to the last thread in a chat by setting the `AttachToLastThread` flag in the request's `payload` to `true`.

## Summary

By now you should be familiar with:

- how does the API calls authentication work using the OAuth2 *access tokens*
- how to make authenticated calls to the external API, such as *LiveChat Configuration API*
- how to set-up a websocket-based connection with the external service, such as *LiveChat RTM API*
- how to receive and send events through a websocket-based connection

We hope that you've found this lesson enjoyable and can't wait to see what you come up with thanks to our extensive API!

Be sure to check out the [full documentation](https://developers.livechatinc.com/beta-docs/) to see what can be done next.