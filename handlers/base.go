package handlers

import (
	"integration/bot"
	"integration/oauth"
	"net/http"
)

func New() http.Handler {
	mux := http.NewServeMux()

	// Static Home
	mux.Handle("/", http.FileServer(http.Dir("templates/")))

	// OAuth
	mux.HandleFunc("/login", oauth.OAuthLiveChatLogin)
	mux.HandleFunc("/callback", oauth.OAuthLiveChatCallback)

	// Bot
	mux.HandleFunc("/bot", bot.StartBotAgent)

	return mux
}
