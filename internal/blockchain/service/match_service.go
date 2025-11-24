package service

import (
	c "Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/loader"
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type MatchChainService struct {
	client    *c.BlockchainClient
	contracts *loader.Contracts
}

func NewMatchChainService(client *c.BlockchainClient, contracts *loader.Contracts) *MatchChainService {
	return &MatchChainService{
		client:    client,
		contracts: contracts,
	}
}

// ===== TIPOS =====

type MatchType uint8

const (
	MatchTypeLocal  MatchType = 0
	MatchTypeRemote MatchType = 1
)

type MatchInfo struct {
	MatchID    string
	MatchType  string
	Status     string
	Player1ID  string
	Player2ID  string
	WinnerID   string
	StartedAt  uint64
	FinishedAt uint64
	TotalTurns uint64
	ServerHost string
}

type PlayerStats struct {
	TotalMatches uint64
	Wins         uint64
	Losses       uint64
	WinRate      uint64
}

// ===== REGISTRAR INÍCIO DE PARTIDA =====

func (ms *MatchChainService) RegisterMatchStart(
	ctx context.Context,
	matchID string,
	isRemote bool,
	player1ID, player2ID string,
	serverID string,
) error {
	auth, err := ms.client.NewTransactor(ctx)
	if err != nil {
		return fmt.Errorf("erro ao criar transactor: %w", err)
	}

	matchType := uint8(MatchTypeLocal)
	if isRemote {
		matchType = uint8(MatchTypeRemote)
	}

	tx, err := ms.contracts.Match.StartMatch(
		auth,
		matchID,
		matchType,
		player1ID,
		player2ID,
		serverID,
	)
	if err != nil {
		return fmt.Errorf("erro ao registrar partida: %w", err)
	}

	log.Printf("[Blockchain] Partida %s iniciada. TX: %s", matchID, tx.Hash().Hex())

	// Aguarda confirmação assíncrona
	go func() {
		receipt, err := bind.WaitMined(context.Background(), ms.client.Client, tx)
		if err != nil {
			log.Printf("⚠️ [Blockchain] Erro ao confirmar TX: %v", err)
			return
		}
		log.Printf(" [Blockchain] Partida %s confirmada no bloco %d",
			matchID, receipt.BlockNumber.Uint64())
	}()

	return nil
}

// ===================== REGISTRAR FIM DE PARTIDA ===================

func (ms *MatchChainService) RegisterMatchFinish(ctx context.Context,matchID string,winnerID string,totalTurns uint64,)error {
	auth, err := ms.client.NewTransactor(ctx)
	if err != nil {
		return fmt.Errorf("erro ao criar transactor: %w", err)
	}

	tx, err := ms.contracts.Match.FinishMatch(
		auth,
		matchID,
		winnerID,
		big.NewInt(int64(totalTurns)),
	)
	if err != nil {
		return fmt.Errorf("erro ao finalizar partida: %w", err)
	}

	log.Printf(" [Blockchain] Partida %s finalizada. Vencedor: %s. TX: %s",
		matchID, winnerID, tx.Hash().Hex())

	// Aguarda confirmação
	receipt, err := bind.WaitMined(ctx, ms.client.Client, tx)
	if err != nil {
		return fmt.Errorf("erro ao confirmar finalização: %w", err)
	}

	log.Printf(" [Blockchain] Finalização confirmada no bloco %d", receipt.BlockNumber.Uint64())
	return nil
}

// =============== REGISTRAR ABANDONO ===============

func (ms *MatchChainService) RegisterMatchAbandon(ctx context.Context,matchID string,playerID string,) error {
	auth, err := ms.client.NewTransactor(ctx)
	if err != nil {
		return fmt.Errorf("erro ao criar transactor: %w", err)
	}

	tx, err := ms.contracts.Match.AbandonMatch(
		auth,
		matchID,
		playerID,
	)
	if err != nil {
		return fmt.Errorf("erro ao registrar abandono: %w", err)
	}

	log.Printf(" [Blockchain] Jogador %s abandonou partida %s. TX: %s",
		playerID, matchID, tx.Hash().Hex())

	receipt, err := bind.WaitMined(ctx, ms.client.Client, tx)
	if err != nil {
		return fmt.Errorf("erro ao confirmar abandono: %w", err)
	}

	log.Printf(" [Blockchain] Abandono confirmado no bloco %d", receipt.BlockNumber.Uint64())
	return nil
}

// =================== CONSULTAS ============================

func (ms *MatchChainService) GetMatchInfo(ctx context.Context, matchID string) (*MatchInfo, error) {
	match, err := ms.contracts.Match.GetMatch(ms.client.CallOpts(), matchID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar partida: %w", err)
	}

	matchType := "LOCAL"
	if match.MatchType == 1 {
		matchType = "REMOTE"
	}

	status := "IN_PROGRESS"
	switch match.Status {
	case 1:
		status = "FINISHED"
	case 2:
		status = "ABANDONED"
	}

	return &MatchInfo{
		MatchID:    match.MatchId,
		MatchType:  matchType,
		Status:     status,
		Player1ID:  match.Player1Id,
		Player2ID:  match.Player2Id,
		WinnerID:   match.WinnerId,
		StartedAt:  match.StartedAt.Uint64(),
		FinishedAt: match.FinishedAt.Uint64(),
		TotalTurns: match.TotalTurns.Uint64(),
		ServerHost: match.ServerHost,
	}, nil
}

func (ms *MatchChainService) GetPlayerStats(ctx context.Context, playerID string) (*PlayerStats, error) {
	stats, err := ms.contracts.Match.GetPlayerStats(ms.client.CallOpts(), playerID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar estatísticas: %w", err)
	}

	return &PlayerStats{
		TotalMatches: stats.Total.Uint64(),
		Wins:         stats.Wins.Uint64(),
		Losses:       stats.Losses.Uint64(),
		WinRate:      stats.WinRate.Uint64(),
	}, nil
}


func (ms *MatchChainService) GetSystemStats(ctx context.Context) (map[string]uint64, error) {
	stats, err := ms.contracts.Match.GetSystemStats(ms.client.CallOpts())
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar estatísticas do sistema: %w", err)
	}
	
	return map[string]uint64{
		"total_matches":  stats.TotalMatches.Uint64(),
		"local_matches":  stats.TotalLocal.Uint64(),
		"remote_matches": stats.TotalRemote.Uint64(),
		"active":         stats.Active.Uint64(),
		"finished":       stats.Finished.Uint64(),
	}, nil
}

func (ms *MatchChainService) MatchExists(ctx context.Context, matchID string) (bool, error) {
	exists, err := ms.contracts.Match.MatchExists(ms.client.CallOpts(), matchID)
	if err != nil {
		return false, err
	}
	return exists, nil
}