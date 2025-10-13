package raft

import (
	disco "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/discovery"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"fmt"
	"net/http"
	"log"
)

// separação das chamdas as rotas da api rest para gestão interna (metodos n rpc)


func (rs *RaftService) TryJoinCluster(discovery *disco.Discovery, myServerInfo *entities.ServerInfo) error {
	if rs.config.Bootstrap {
		log.Println("Bootstrap ativado — iniciando cluster como líder.")
		return nil
	}

	if len(discovery.KnownServers) == 0 {
		log.Println("Nenhum servidor conhecido — aguardando descoberta...")
		return nil
	}

	log.Println("Tentando ingressar no cluster existente...")

	// Descobrir líder
	leaderInfo, err := rs.GetLeader(discovery)
	if err != nil {
		return fmt.Errorf("falha ao localizar líder: %w", err)
	}

	// Solicitar join
	if err := rs.joinLeader(leaderInfo, myServerInfo); err != nil {
		return fmt.Errorf("falha ao solicitar join: %w", err)
	}

	log.Println("Join aceito — servidor adicionado ao cluster.")
	return nil
}


// faz broadcast em servidores conhecidos para descobrir o leader
func (rs *RaftService) GetLeader(discovery *disco.Discovery) (*entities.ServerInfo, error) {
	if !rs.config.Bootstrap && len(discovery.KnownServers) > 0 {

		for id, server := range discovery.KnownServers {
			if id == discovery.MyInfo.ID {
				continue
			}
			// Faz request GET /api/v1/raft/status
			respMap, err := rs.ApiClient.RaftInterface.GetLeader(server.Address, server.Port)
			if err != nil {
				continue 
			}
			// Verifica se é líder
			isLeader, ok := respMap["is_leader"].(bool)
			if ok && isLeader {
				return server, nil
			}
		}
	}

	return nil, fmt.Errorf("nenhum líder encontrado no cluster")
}


// manda solicitaão http post para fazer join no cluster 
func (rs *RaftService) joinLeader(leaderInfo *entities.ServerInfo, myServerInfo *entities.ServerInfo) error {
	httpAddr := fmt.Sprintf("http://%s:%d", myServerInfo.Address, myServerInfo.Port)
	req := map[string]string{
		"server_id": myServerInfo.ID,
		"http_addr": httpAddr,
	}

	resp, err := rs.ApiClient.RaftInterface.RequestJoin(leaderInfo, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("join rejeitado (%d)", resp.StatusCode)
	}
	return nil
}
