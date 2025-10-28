package matchhandler

import (
	"log"
	"net/http"
	"time"

	gamesession "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/gameSession"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/gameSession/remota"
	matchglobal "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/matchMacking/match_global"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"

	"github.com/gin-gonic/gin"
)

type MatchHandler struct {
	sessionManager    *gamesession.GameSessionManager
	globalMatchmaking *matchglobal.GlobalMatchmakingService
}

func New(sessionManager *gamesession.GameSessionManager, globalMatchmaking *matchglobal.GlobalMatchmakingService) *MatchHandler {
	return &MatchHandler{
		sessionManager:    sessionManager,
		globalMatchmaking: globalMatchmaking,
	}
}
// POST /api/v1//match/global/join
// è rota para servidores que não lideres requsitarem lider 
// para colocar player na lista global 
func (h *MatchHandler) HandleJoinGlobalQueue(ctx *gin.Context) {
	var req struct {
		PlayerID string `json:"player_id"`
		Username string `json:"username"`
		ServerID string `json:"server_id"`
		ClientID string `json:"client_id"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	log.Printf("[MatchHandler] Requisição para adicionar à fila global: %s", req.Username)

	err := h.globalMatchmaking.JoinGlobalQueue(
		req.ClientID,
		req.PlayerID,
		req.Username,
		req.ServerID,
	)

	if err != nil {
		log.Printf("[MatchHandler] Erro ao adicionar à fila global: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}


	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "player adicionado à fila global",
	})
}

// POST /api/v1//match/created
// rota para servidor lider do cluester mandar servidres da partida remota
// criarem suas partidas 
func (h *MatchHandler) HandleRemoteMatchNotification(ctx *gin.Context) {
	var payload struct {
		MatchID              string `json:"match_id"`
		LocalPlayerID        string `json:"local_player_id"`
		LocalClientID        string `json:"local_player_client_id"`
		RemotePlayerID       string `json:"remote_player_id"`
		RemotePlayerUsername string `json:"remote_player_username,omitempty"`
		RemoteServerID       string `json:"remote_server_id"`
		IsHost               bool   `json:"is_host"`
	}

	if err := ctx.BindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	if payload.MatchID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid matchID"})
		return
	}

	log.Printf("  [MatchHandler] Notificação de partida remota recebida")
	log.Printf("   MatchID: %s", payload.MatchID)
	log.Printf("   LocalPlayer: %s", payload.LocalPlayerID)
	log.Printf("   RemotePlayer: %s (%s)", payload.RemotePlayerID, payload.RemotePlayerUsername)
	log.Printf("   IsHost: %v", payload.IsHost)

	err := h.sessionManager.CreateRemoteMatch(
		payload.MatchID,
		payload.LocalPlayerID,
		payload.LocalClientID,
		payload.RemotePlayerID,
		payload.RemotePlayerUsername,
		payload.RemoteServerID,
		payload.IsHost,
	)

	if err != nil {
		log.Printf(" [MatchHandler] Erro ao criar partida remota: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf(" [MatchHandler] Partida remota criada com sucesso")
	ctx.JSON(http.StatusOK, gin.H{"status": "remote match created"})
}

// rota para servidor não host receber sincronização da partida remta 
//  POST /api/v1//match/sync
func (h *MatchHandler) HandleMatchSync(ctx *gin.Context) {
	var update remota.GameStateUpdate

	if err := ctx.BindJSON(&update); err != nil {
		log.Printf("[MatchHandler] Erro ao parsear sync: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	//  LOGS DETALHADOS de debuh
	log.Printf("   [MatchHandler] SYNC RECEBIDO:")
	log.Printf("   MatchID: %s", update.MatchID)
	log.Printf("   Turn: %d | CurrentTurn: %s", update.TurnNumber, update.CurrentTurnPlayerID)
	log.Printf("   Status: %s", update.Status)
	log.Printf("   LocalPlayerLife: %d | HasCard: %v", 
		update.LocalPlayerLife, update.LocalPlayerCurrentCard != nil)
	log.Printf("   RemotePlayerLife: %d | HasCard: %v", 
		update.RemotePlayerLife, update.RemotePlayerCurrentCard != nil)


	
	ctx.JSON(http.StatusAccepted, gin.H{"status": "sync received"})

	go func() {
		if err := h.sessionManager.ReceiveRemoteSync(update.MatchID, update); err != nil {
			log.Printf("[MatchHandler] Erro ao processar sync: %v", err)
		} else {
			log.Printf("[MatchHandler] Sincronização processada com sucesso")
		}
	}()
}


// rota para srvidor host receber ação do jogador remoto 
//  POST /api/v1//match/action
func (h *MatchHandler) HandleMatchAction(ctx *gin.Context) {
	var req struct {
		MatchID  string              `json:"match_id"`
		PlayerID string              `json:"player_id"`
		Action   entities.GameAction `json:"action"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("[MatchHandler] Erro ao parsear JSON: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	log.Printf(" [MatchHandler] Ação recebida:")
	log.Printf("   MatchID: %s", req.MatchID)
	log.Printf("   PlayerID: %s", req.PlayerID)
	log.Printf("   Action Type: %s", req.Action.Type)

	
	if req.Action.Type == "leave_match" {
		log.Printf("[MatchHandler] Leave recebido, respondendo e processando...")
		ctx.JSON(http.StatusOK, gin.H{"status": "leave received"})
	} else {
		ctx.JSON(http.StatusAccepted, gin.H{"status": "action received"})
	}

	
	go func() {
		if err := h.sessionManager.ProcessAction(req.MatchID, req.PlayerID, req.Action); err != nil {
			log.Printf("[MatchHandler] Erro ao processar ação: %v", err)
		} else {
			log.Printf("[MatchHandler] Ação %s processada com sucesso", req.Action.Type)
		}
	}()
}


// rota para servidor receber heartbeat periodico 
//  /api/v1//match/heartbeat
func (h *MatchHandler) HandleHeartbeat(ctx *gin.Context) {
	var req struct {
		MatchID string `json:"match_id"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	h.sessionManager.UpdateHeartbeat(req.MatchID, time.Now())

	ctx.JSON(http.StatusOK, gin.H{"status": "heartbeat updated"})
}