package topics

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/handler"
	websocket "Jogo-de-Cartas-Multiplayer-Distribuido/internal/pubsub/webSocket"
)

// seta os handlersTopics da aplicação
func SetUpTopics(pubsub websocket.PubSubHandler, handler *handler.Handler ){
	pubsub.RegisterHandler("auth", handler.AuthHandler)
	pubsub.RegisterHandler("package", handler.PackageHandler)
	pubsub.RegisterHandler("match", handler.MatchHandler)
	pubsub.RegisterHandler("trade", handler.TradeHandler)
}