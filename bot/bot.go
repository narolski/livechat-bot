package bot

import (
	"encoding/json"
	"integration/oauth"
	"log"
	"net/http"
	"path"
	"strings"
	"text/template"

	"github.com/gorilla/websocket"
)

var (
	botAgentID  string = "d5b2377faebc9c5aef9a2bdc40bf7510"
	triggerWord string = "pizza"
	botResponse string = "The pizza is on its' way!"
)

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

// botHandleIncomingMessage checks the incoming chat message for a trigger word presence and delegates appropriate actions
func botHandleIncomingMessage(c *websocket.Conn, message string, chatID string, authorID string) error {

	if strings.Contains(message, triggerWord) && authorID != botAgentID {
		// If the message contains bot's trigger word, send a defined response
		log.Printf("Found messagee containing the trigger word '%s': '%s'", triggerWord, message)
		botSendChatMessage(c, chatID, botResponse)
	}

	return nil
}

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
