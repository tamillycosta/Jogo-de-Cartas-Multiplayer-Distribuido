package tradehandler

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/comands"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TradeHandler struct {
	gameServer *service.GameServer
}

func New(gameServer *service.GameServer) *TradeHandler {
	return &TradeHandler{
		gameServer: gameServer,
	}
}

// POST /api/v1/trade/execute
// Rota para servidores que nao sao lideres
// requisitarem ao lider a execucao de uma troca.
func (th *TradeHandler) ExecuteTrade(ctx *gin.Context) {
	var cmd comands.TradeCardsCommand

	if err := ctx.BindJSON(&cmd); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// O GameServer tem a instância do TradeService
	// O serviço já é o `ExecuteTradeAsLeader`
	err := th.gameServer.Trade.ExecuteTradeAsLeader(cmd)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "trade command applied",
		"request_id": cmd.RequestID,
	})
}