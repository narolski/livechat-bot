package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var (
	agentChatAPIURL string        = "wss://api.livechatinc.com/v3.0/agent/rtm/ws"
	pingInterval    time.Duration = 15 * time.Second
	botAgentID      string        = "d5b2377faebc9c5aef9a2bdc40bf7510"
)

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

// pinger pings the server at a given time interval in order to maintain the websocket connection
func apiPinger(c *websocket.Conn) {
	t := time.NewTimer(pingInterval)

	for {
		<-t.C
		c.WriteMessage(websocket.PingMessage, []byte{})
		t.Reset(pingInterval)
	}
}

// apiLogin handles the user authentication using the generated token k
func apiLogin(c *websocket.Conn) error {
	type loginRequest struct {
		Token string `json:"token"`
	}

	// Get the access token
	// token := "Bearer " + oauth.GetLiveChatAPIToken().AccessToken
	token := "Bearer dal:gaJD246dSRCJ-LgZDPOtRg"

	payload := &loginRequest{
		Token: token,
	}
	return apiSendRequest(c, "login", false, payload)
}

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
