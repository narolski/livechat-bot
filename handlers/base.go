package handlers

import (
	"net/http"
	"oauth2-example/bot"
	"oauth2-example/oauth"
)

func New() http.Handler {
	mux := http.NewServeMux()
	// Root
	mux.Handle("/", http.FileServer(http.Dir("templates/")))

	// OauthLiveChat
	mux.HandleFunc("/login", oauth.OauthLiveChatLogin)
	mux.HandleFunc("/callback", oauth.OauthLiveChatCallback)
	mux.HandleFunc("/chats", bot.WriteResponse)

	return mux
}
