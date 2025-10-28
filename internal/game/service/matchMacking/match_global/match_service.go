package matchglobal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/aplication/usecases"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/comands"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/fms"

	"github.com/google/uuid"
)

type GlobalMatchmakingService struct {
	raft      *raft.RaftService
	fsm       *fms.GameFSM
	apiClient *client.Client
	serverID  string

	stopMatchmaking chan struct{}
}

func New(
	raft *raft.RaftService,
	fsm *fms.GameFSM,
	apiClient *client.Client,
	serverID string,
) *GlobalMatchmakingService {
	gms := &GlobalMatchmakingService{
		raft:            raft,
		fsm:             fsm,
		apiClient:       apiClient,
		serverID:        serverID,
		stopMatchmaking: make(chan struct{}),
	}

	// Loop de matchmaking global (apenas líder processa)
	go gms.globalMatchmakingLoop()

	return gms
}

//  Jogador entra na fila global (via Raft)
func (gms *GlobalMatchmakingService) JoinGlobalQueue(clientID, playerID, username, serverID string) error {
	log.Printf("[GlobalMatchmaking] Player %s entrando na fila GLOBAL...", username)
	
	if !gms.raft.IsLeader() {
		log.Printf("[GlobalMatchmaking] Não sou líder, encaminhando para líder...")
		return gms.forwardToLeader(clientID, playerID, username, serverID)
	}

	if (len(gms.raft.GetServers()) <= 1){
		return errors.New("não é possivel fazer partida remota, apenas 1 server no cluter")
	}

	return gms.joinGlobalQueueAsLeader(clientID, playerID, username, serverID)

}

func (gms *GlobalMatchmakingService) forwardToLeader(clientID, playerID, username, serverID string) error {
	leaderAddr := gms.raft.GetLeaderHTTPAddr()

	if leaderAddr == "" {
		return fmt.Errorf("nenhum líder disponível")
	}

	log.Printf("[GlobalMatchmaking] Encaminhando para líder: %s", leaderAddr)

	err := gms.apiClient.MatchInterface.JoinGlobalQueue(
		leaderAddr,
		playerID,
		username,
		serverID,
		clientID,
	)

	if err != nil {
		return fmt.Errorf("erro ao encaminhar para líder: %v", err)
	}

	log.Printf("[GlobalMatchmaking] Requisição encaminhada com sucesso")
	return nil
}

func (gms *GlobalMatchmakingService) joinGlobalQueueAsLeader(clientID, playerID, username, serverID string) error {
	log.Printf("[GlobalMatchmaking] SOU LÍDER! Adicionando %s à fila global via Raft...", username)

	cmd := comands.JoinGlobalQueueCommand{
		PlayerID: playerID,
		Username: username,
		ServerID: serverID,
		ClientID: clientID,
		JoinedAt: time.Now().Unix(),
	}

	data, _ := json.Marshal(cmd)
	response, err := gms.raft.ApplyCommand(comands.Command{
		Type: comands.CommandJoinGlobalQueue,
		Data: data,
	})

	if err != nil {
		return fmt.Errorf("erro ao aplicar comando Raft: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("falha ao entrar na fila global: %s", response.Error)
	}

	log.Printf("[GlobalMatchmaking] Player %s adicionado à fila global", username)
	return nil
}

// LeaveGlobalQueue - Jogador sai da fila global (via Raft)
func (gms *GlobalMatchmakingService) LeaveGlobalQueue(playerID string) error {
	log.Printf("[GlobalMatchmaking] Player %s saindo da fila global...", playerID)

	cmd := comands.LeaveGlobalQueueCommand{
		PlayerID: playerID,
	}

	data, _ := json.Marshal(cmd)
	response, err := gms.raft.ApplyCommand(comands.Command{
		Type: comands.CommandLeaveGlobalQueue,
		Data: data,
	})

	if err != nil {
		return fmt.Errorf("erro ao sair da fila global: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("falha ao sair da fila global: %s", response.Error)
	}

	log.Printf("[GlobalMatchmaking] Player %s removido da fila global", playerID)
	return nil
}

// Apenas LÍDER cria partidas remotas
func (gms *GlobalMatchmakingService) globalMatchmakingLoop() {
	ticker := time.NewTicker(3 * time.Second) // Verifica a cada 3 segundos
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !gms.raft.IsLeader() {
				continue // Apenas líder cria matches
			}

			gms.processGlobalQueue()

		case <-gms.stopMatchmaking:
			log.Println("[GlobalMatchmaking] Loop encerrado")
			return
		}
	}
}

// Processa fila global e cria partidas remotas
func (gms *GlobalMatchmakingService) processGlobalQueue() {
	state := gms.fsm.GetGlobalMatchmakingState()

	p1, p2 := state.GetNextGlobalPair()

	if p1 == nil || p2 == nil {
		return // Fila vazia ou com apenas 1 jogador
	}

	// Cria partida remota
	if err := gms.createRemoteMatch(p1, p2); err != nil {
		log.Printf("[GlobalMatchmaking] Erro ao criar match remoto: %v", err)
		// Remove jogadores da fila em caso de erro
		gms.LeaveGlobalQueue(p1.PlayerID)
		gms.LeaveGlobalQueue(p2.PlayerID)
		// adiciona de novo para tentar outro match 
		gms.JoinGlobalQueue(p1.ClientID,p1.PlayerID, p1.Username, p1.ServerID)
		gms.JoinGlobalQueue(p2.ClientID,p2.PlayerID, p2.Username, p2.ServerID)
	}
}

// createRemoteMatch - Cria partida remota e notifica servidores
func (gms *GlobalMatchmakingService) createRemoteMatch(p1, p2 *entities.GlobalQueueEntry) error {
	if(gms.raft.IsLeader()){
		matchID := uuid.New().String()

	// Escolhe servidor host (simplificado: servidor do player1)

	hostServer := p1.ServerID

	if(p1.ServerID == p2.ServerID){
		return errors.New("não é possivel fazer partida remota com players de mesmo servidor")
		
	}

	
	// Cria match remoto via Raft (replica em todos os servidores)
	cmd := comands.CreateRemoteMatchCommand{
		MatchID:         matchID,
		Player1ID:       p1.PlayerID,
		Player1Server:   p1.ServerID,
		Player1ClientID: p1.ClientID,
		Player2ID:       p2.PlayerID,
		Player2Server:   p2.ServerID,
		Player2ClientID: p2.ClientID,
		HostServer:      hostServer,
	}

	data, _ := json.Marshal(cmd)
	response, err := gms.raft.ApplyCommand(comands.Command{
		Type: comands.CommandCreateRemoteMatch,
		Data: data,
	})

	if err != nil {
		return fmt.Errorf("erro ao criar match via Raft: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("falha ao criar match: %s", response.Error)
	}

	// Notifica servidores envolvidos via API REST
	go gms.notifyServersAboutRemoteMatch(matchID, p1, p2, hostServer)

	log.Printf("[GlobalMatchmaking] Match remoto %s criado | Host=%s", matchID, hostServer)
	return nil
	}
	return errors.New("não sou lider ")
}

// Notifica servidores sobre nova partida remota
func (gms *GlobalMatchmakingService) notifyServersAboutRemoteMatch(
	matchID string,
	p1, p2 *entities.GlobalQueueEntry,
	hostServer string,
) {
	log.Printf("  [GlobalMatchmaking] Notificando servidores sobre match %s", matchID)
	log.Printf("   P1: %s (srv=%s, client=%s)", p1.Username, p1.ServerID, p1.ClientID)
	log.Printf("   P2: %s (srv=%s, client=%s)", p2.Username, p2.ServerID, p2.ClientID)
	log.Printf("   Host: %s", hostServer)
	
	//  Notifica servidor do Player1
	notification1 := map[string]interface{}{
		"match_id":                matchID,
		"local_player_id":         p1.PlayerID,
		"local_player_client_id":  p1.ClientID,
		"remote_player_id":        p2.PlayerID,
		"remote_player_username":  p2.Username,
		"remote_server_id":        p2.ServerID,
		"is_host":                 (p1.ServerID == hostServer),
	}


	if err := gms.notifyServer(p1.ServerID, notification1); err != nil {
		log.Printf("[GlobalMatchmaking] Erro ao notificar %s: %v", p1.ServerID, err)
	} else {
		log.Printf("[GlobalMatchmaking] %s notificado (isHost=%v)", p1.ServerID, p1.ServerID == hostServer)
	}

	//  Notifica servidor do Player2
	notification2 := map[string]interface{}{
		"match_id":                matchID,
		"local_player_id":         p2.PlayerID,
		"local_player_client_id":  p2.ClientID,
		"remote_player_id":        p1.PlayerID,
		"remote_player_username":  p1.Username,
		"remote_server_id":        p1.ServerID,
		"is_host":                 (p2.ServerID == hostServer),
	}


	if err := gms.notifyServer(p2.ServerID, notification2); err != nil {
		log.Printf("[GlobalMatchmaking] Erro ao notificar %s: %v", p2.ServerID, err)
	} else {
		log.Printf("[GlobalMatchmaking] %s notificado (isHost=%v)", p2.ServerID, p2.ServerID == hostServer)
	}

	log.Printf("[GlobalMatchmaking] Notificações completas para match %s", matchID)
}


func (gms *GlobalMatchmakingService) notifyServer(serverID string, notification map[string]interface{}) error {
		
	serverAddr := gms.getServerHTTPAddr(serverID)
	if serverAddr == "" {
		return fmt.Errorf("endereço HTTP do servidor %s não encontrado", serverID)
	}

	log.Printf("   Endereço: %s", serverAddr)
	

	if err := gms.apiClient.MatchInterface.NotifyRemoteMatchCreated(serverAddr, notification); err != nil {
		return fmt.Errorf("erro ao notificar %s: %w", serverID, err)
	}

	log.Printf("[GlobalMatchmaking] Servidor %s notificado com sucesso", serverID)
	return nil
}


//  Busca endereço HTTP do servidor no Raft
func (gms *GlobalMatchmakingService) getServerHTTPAddr(serverID string) string {
	// Busca na configuração do Raft
	serverAddr := usecases.GetServer(serverID, gms.raft.GetServers())
	if(serverAddr == ""){
		return ""
	}
	return  serverAddr
	
}


// Finaliza partida remota via Raft
func (gms *GlobalMatchmakingService) EndRemoteMatch(matchID, winnerID, reason string) error {
	log.Printf("[GlobalMatchmaking] Finalizando match remoto %s | Vencedor=%s", matchID, winnerID)

	cmd := comands.EndRemoteMatchCommand{
		MatchID:  matchID,
		WinnerID: winnerID,
		Reason:   reason,
		EndedAt:  time.Now().Unix(),
	}

	data, _ := json.Marshal(cmd)
	response, err := gms.raft.ApplyCommand(comands.Command{
		Type: comands.CommandEndRemoteMatch,
		Data: data,
	})

	if err != nil {
		return fmt.Errorf("erro ao finalizar match: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("falha ao finalizar match: %s", response.Error)
	}

	log.Printf("[GlobalMatchmaking] Match remoto %s finalizado", matchID)
	return nil
}

// GetGlobalQueueSize - Retorna tamanho da fila global
func (gms *GlobalMatchmakingService) GetGlobalQueueSize() int {
	return gms.fsm.GetGlobalMatchmakingState().GetGlobalQueueSize()
}

// GetRemoteMatch - Retorna metadados da partida remota
func (gms *GlobalMatchmakingService) GetRemoteMatch(matchID string) (*entities.RemoteMatch, error) {
	match, exists := gms.fsm.GetGlobalMatchmakingState().GetRemoteMatch(matchID)
	if !exists {
		return nil, fmt.Errorf("match remoto não encontrado")
	}
	return match, nil
}

// Shutdown - Para o loop de matchmaking
func (gms *GlobalMatchmakingService) Shutdown() {
	close(gms.stopMatchmaking)
}
