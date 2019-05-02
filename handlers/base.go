package handlers

import (
	"integration/bot"
	"integration/oauth"
	"net/http"
)

func New() http.Handler {
	mux := http.NewServeMux()
	// Root
	mux.Handle("/", http.FileServer(http.Dir("templates/")))

	// OauthLiveChat
	mux.HandleFunc("/login", oauth.OauthLiveChatLogin)
	mux.HandleFunc("/callback", oauth.OauthLiveChatCallback)
	mux.HandleFunc("/bot", bot.StartBotAgent)

	return mux
}
