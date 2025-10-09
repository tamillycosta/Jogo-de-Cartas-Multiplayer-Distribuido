package topics

import (
	
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler/authHandler"
	websocket "Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub/webSocket"
)

// seta os handlersTopics da aplicação
func SetUpTopics(pubsub websocket.PubSubHandler, authHandler *handlers.AuthTopicHandler){
	pubsub.RegisterHandler("auth", authHandler)
}