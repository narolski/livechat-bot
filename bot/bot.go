package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"oauth2-example/oauth"
)

type Archives struct {
	Filters    string `json:"filters,omitempty"`
	Pagination string `json:"pagination,omitempty"`
}

// getActiveThreads returns the active threads the authorized agent has access to
func getActiveThreads() string {

	service := "https://api.livechatinc.com/v3.0/agent/action/get_archives"
	content_type := "application/json"
	client := oauth.LiveChatAPIClient()

	type protocolRequest struct {
		Action    string      `json:"action"`
		RequestID string      `json:"request_id,omitempty"`
		Payload   interface{} `json:"payload"`
	}

	pr := &protocolRequest{"get_archives", "", Archives{}}

	requestBody, err := json.Marshal(pr)

	if err != nil {
		log.Fatalln(err)
	}

	request, err := http.NewRequest("POST", service, bytes.NewBuffer(requestBody))

	if err != nil {
		log.Fatalln(err)
	}

	request.Header.Add("Content-Type", content_type)
	request.Header.Add("Authorization", oauth.LiveChatToken.AccessToken)

	response, err := client.Do(request)

	if err != nil {
		log.Fatalln(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatalln(err)
	}

	return string(body)
}

func WriteResponse(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Response: %s\n", getActiveThreads())
}
