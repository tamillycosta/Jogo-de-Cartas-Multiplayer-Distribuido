package gamesession

import (
	"context"
	"log"
)


// Registrar início da partida na blockchain
func (gsm *GameSessionManager) registerMatchStart(
	matchID string,
	isRemote bool,
	player1ID, player2ID string,
	serverID string,
) {
	// Verifica se blockchain está disponível
	if gsm.chainService == nil || gsm.chainService.MatchChainService == nil {
		log.Printf("⚠️ [Blockchain] Serviço não disponível para registrar partida %s", matchID)
		return
	}

	ctx := context.Background()
	err := gsm.chainService.MatchChainService.RegisterMatchStart(
		ctx,
		matchID,
		isRemote,
		player1ID,
		player2ID,
		serverID,
	)

	if err != nil {
		log.Printf("⚠️ [Blockchain] Erro ao registrar início da partida %s: %v", matchID, err)
	} else {
		log.Printf(" [Blockchain] Partida %s registrada com sucesso", matchID)
	}
}

// Registrar fim da partida na blockchain
func (gsm *GameSessionManager) registerMatchFinish(matchID string,	winnerID string,totalTurns uint64,wasAbandoned bool,abandonerID string,
) {
	if gsm.chainService == nil || gsm.chainService.MatchChainService == nil {
		log.Printf(" [Blockchain] Serviço não disponível para finalizar partida %s", matchID)
		return
	}

	ctx := context.Background()
	var err error

	if wasAbandoned {
		// Registra abandono
		err = gsm.chainService.MatchChainService.RegisterMatchAbandon(
			ctx,
			matchID,
			abandonerID,
		)
	} else {
		// Registra finalização normal
		err = gsm.chainService.MatchChainService.RegisterMatchFinish(
			ctx,
			matchID,
			winnerID,
			totalTurns,
		)
	}

	if err != nil {
		log.Printf("[Blockchain] Erro ao finalizar partida %s: %v", matchID, err)
	} else {
		log.Printf(" [Blockchain] Partida %s finalizada na blockchain", matchID)
	}
}