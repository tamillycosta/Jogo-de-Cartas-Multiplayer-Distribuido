package raft

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/client"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/comands"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/fms"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/trasport"

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/util"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

// gerencia o cluster Raft usando HTTP para comunicação
type RaftService struct {
	raft      *raft.Raft
	fsm       *fms.GameFSM
	transport *trasport.HTTPTransport
	config    *RaftConfig
	ApiClient    *client.Client
}

type RaftConfig struct {
	ServerID    string
	HTTPAddr    string
	RaftDir     string 
	Bootstrap   bool   // Se true, inicia como líder
}

func New(config *RaftConfig, fsm *fms.GameFSM,  client *client.Client) (*RaftService, error) {
	rs := &RaftService{
		fsm:    fsm,
		config: config,
		ApiClient: client,
	}

	if err := rs.setupRaft(); err != nil {
		return nil, err
	}

	return rs, nil
}

// chamado no SetUpGameServer
func InitRaft(repository *repository.PlayerRepository, myServerInfo *entities.ServerInfo, client *client.Client)(*RaftService, error) {
	httpAddr := fmt.Sprintf("http://%s:%d", myServerInfo.Address, myServerInfo.Port)
	raftDir := filepath.Join("./data", myServerInfo.ID, "raft")
	bootstrap := util.GetEnvBool("RAFT_BOOTSTRAP",false)

	// Inicializa sistema de raft 
	raftConfig := &RaftConfig{
		ServerID:  myServerInfo.ID,
		HTTPAddr:  httpAddr, 
		RaftDir:   raftDir,
		Bootstrap: bootstrap,
		
	}
	fsm := fms.New(repository)
	return New(raftConfig, fsm, client)

}

// Seta configurações basicas do raft
func (rs *RaftService) setupRaft() error {
	
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(rs.config.ServerID)

	// Timeouts
	raftConfig.HeartbeatTimeout = 1000 * time.Millisecond
	raftConfig.ElectionTimeout = 1000 * time.Millisecond
	raftConfig.CommitTimeout = 500 * time.Millisecond
	raftConfig.LeaderLeaseTimeout = 500 * time.Millisecond

	// Diretório para dados
	if err := os.MkdirAll(rs.config.RaftDir, 0755); err != nil {
		return fmt.Errorf("failed to create raft directory: %v", err)
	}

	// LogStore e StableStore (BoltDB)
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(rs.config.RaftDir, "raft-log.db"))
	if err != nil {
		return fmt.Errorf("failed to create log store: %v", err)
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(rs.config.RaftDir, "raft-stable.db"))
	if err != nil {
		return fmt.Errorf("failed to create stable store: %v", err)
	}

	// SnapshotStore
	snapshotStore, err := raft.NewFileSnapshotStore(rs.config.RaftDir, 2, os.Stderr)
	if err != nil {
		return fmt.Errorf("failed to create snapshot store: %v", err)
	}

	// Transport HTTP 
	transport := trasport.New(rs.config.HTTPAddr, 10*time.Second)
	rs.transport = transport

	// Cria instância do Raft
	ra, err := raft.NewRaft(raftConfig, rs.fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return fmt.Errorf("failed to create raft: %v", err)
	}
	rs.raft = ra

	// Bootstrap (se for true , significa que criou o cluster)
	if rs.config.Bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raft.ServerID(rs.config.ServerID),
					Address: raft.ServerAddress(rs.config.HTTPAddr),
				},
			},
		}
		future := ra.BootstrapCluster(configuration)
		if err := future.Error(); err != nil {
			log.Printf("⚠️ Bootstrap error: %v", err)
		} else {
			log.Printf("Cluster bootstrapped: %s (HTTP: %s)", rs.config.ServerID, rs.config.HTTPAddr)
		}
	}

	log.Printf("Raft configurado: %s usando HTTP transport", rs.config.ServerID)
	return nil
}

// retorna o transport HTTP (para handlers)
func (rs *RaftService) GetTransport() *trasport.HTTPTransport {
	return rs.transport
}

//aplica comando no cluster
func (rs *RaftService) ApplyCommand(cmd comands.Command) (*comands.ApplyResponse, error) {
	if cmd.RequestID == "" {
		cmd.RequestID = uuid.New().String()
	}
	cmd.Timestamp = time.Now()

	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %v", err)
	}

	future := rs.raft.Apply(data, 10*time.Second) // aplica comandos na mfs
	if err := future.Error(); err != nil {
		return nil, fmt.Errorf("failed to apply command: %v", err)
	}

	response := future.Response().(*comands.ApplyResponse)
	return response, nil
}

// verifica se é líder
func (rs *RaftService) IsLeader() bool {
	return rs.raft.State() == raft.Leader
}


func (rs *RaftService) GetLeaderHTTPAddr() string {
	addr, _ := rs.raft.LeaderWithID()
	return string(addr)
}

func (rs *RaftService) GetLeaderID() string {
	_, id := rs.raft.LeaderWithID()
	return string(id)
}

// aguarda eleição
func (rs *RaftService) WaitForLeader(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if rs.GetLeaderHTTPAddr() != "" {
			log.Printf("Líder eleito: %s (HTTP: %s)",
				rs.GetLeaderID(),
				rs.GetLeaderHTTPAddr())
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for leader")
}



// adiciona servidor ao cluster
// httpAddr deve ser o endereço HTTP completo (ex: "http://server-b:8081")
func (rs *RaftService) AddVoter(serverID, httpAddr string) error {
	if !rs.IsLeader() {
		return fmt.Errorf("only leader can add voters")
	}

	log.Printf("Adicionando servidor: %s (HTTP: %s)", serverID, httpAddr)

	future := rs.raft.AddVoter(
		raft.ServerID(serverID),
		raft.ServerAddress(httpAddr), 
		0,
		10*time.Second,
	)

	if err := future.Error(); err != nil {
		return fmt.Errorf("failed to add voter: %v", err)
	}

	log.Printf("Servidor %s adicionado ao cluster", serverID)
	return nil
}


func (rs *RaftService) RemoveServer(serverID string) error {
	if !rs.IsLeader() {
		return fmt.Errorf("only leader can remove servers")
	}

	future := rs.raft.RemoveServer(raft.ServerID(serverID), 0, 10*time.Second)
	if err := future.Error(); err != nil {
		return fmt.Errorf("failed to remove server: %v", err)
	}

	log.Printf("Servidor %s removido", serverID)
	return nil
}

// retorna estatísticas
func (rs *RaftService) GetStats() map[string]string {
	return rs.raft.Stats()
}


// retorna lista de servidores
func (rs *RaftService) GetServers() []entities.ServerInfo {
	future := rs.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return nil
	}

	var servers []entities.ServerInfo
	for _, server := range future.Configuration().Servers {
		servers = append(servers, entities.ServerInfo{
			ID:       string(server.ID),
			Address:  string(server.Address), // HTTP address
			IsLeader: string(server.ID) == rs.GetLeaderID(),
		})
	}
	return servers
}



// desliga gracefully
func (rs *RaftService) Shutdown() error {
	log.Println("Desligando Raft...")
	future := rs.raft.Shutdown()
	return future.Error()
}



