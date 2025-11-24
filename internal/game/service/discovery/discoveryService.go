package discovery

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/util"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/memberlist"
)

type Discovery struct {
	MyInfo       *entities.ServerInfo
	KnownServers map[string]*entities.ServerInfo
	mu           sync.RWMutex
	memberlist   *memberlist.Memberlist
}

func SetUpDiscovery(myInfo *entities.ServerInfo) (*Discovery, error) {
	gossipPort := util.GetPortFromEnv("GOSSIP_PORT", 7947)
	return New(myInfo, gossipPort)
}

func New(myInfo *entities.ServerInfo, bindPort int) (*Discovery, error) {
	d := &Discovery{
		MyInfo:       myInfo,
		KnownServers: make(map[string]*entities.ServerInfo),
	}

	// Configura memberlist
	config := memberlist.DefaultLocalConfig()
	config.Name = myInfo.ID
	config.BindPort = bindPort
	config.AdvertisePort = bindPort
	
	// IMPORTANTE: Liga logs para debug
	config.LogOutput = log.Writer()

	metadata, err := json.Marshal(myInfo)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar metadata: %w", err)
	}

	config.Delegate = &delegate{meta: metadata}
	config.Events = &eventDelegate{discovery: d}

	list, err := memberlist.Create(config)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar memberlist: %w", err)
	}

	d.memberlist = list

	log.Printf("[Discovery] Servidor %s iniciado na porta %d (gossip)", myInfo.ID, bindPort)
	log.Printf("[Discovery] Aguardando broadcasts na porta 9000...")

	// Inicia descoberta automática
	go d.startAutoDiscovery(bindPort)

	return d, nil
}
func (d *Discovery) startAutoDiscovery(gossipPort int) {
	broadcastPort := util.GetPortFromEnv("DISCOVERY_PORT", 9000)

	// conjunto para rastrear servidores já processados
	discoveredServers := make(map[string]bool)
	var discMu sync.Mutex

	// goroutine para escutar broadcasts
	go func() {
		addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", broadcastPort))
		conn, err := net.ListenUDP("udp4", addr)
		if err != nil {
			log.Printf("[Discovery] Erro ao iniciar listener UDP: %v\n", err)
			return
		}
		defer conn.Close()

		buf := make([]byte, 1024)
		for {
			n, remoteAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}

			var msg entities.ServerInfo
			if err := json.Unmarshal(buf[:n], &msg); err != nil {
				continue
			}

			// ignora o próprio servidor
			if msg.ID == d.MyInfo.ID {
				continue
			}

			// verifica se já foi processado (independente do Join ter dado certo)
			discMu.Lock()
			if discoveredServers[msg.ID] {
				discMu.Unlock()
				continue
			}
			discoveredServers[msg.ID] = true
			discMu.Unlock()

			// tenta conectar
			addr := fmt.Sprintf("%s:%d", remoteAddr.IP.String(), gossipPort)
			
				log.Printf("[Discovery] Novo servidor detectado: %s em %s", msg.ID, addr)
		
				d.KnownServers[msg.ID] = &msg
			}
		
	}()

	// 2goroutine para enviar broadcasts
	go func() {
		conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{
			IP:   net.ParseIP("255.255.255.255"),
			Port: broadcastPort,
		})
		if err != nil {
			log.Printf("[Discovery] Erro ao iniciar broadcaster UDP: %v\n", err)
			return
		}
		defer conn.Close()

		data, _ := json.Marshal(d.MyInfo)
		for {
			_, _ = conn.Write(data)
			time.Sleep(3 * time.Second)
		}
	}()
}



func (d *Discovery) GetKnownServers() map[string]*entities.ServerInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Cria cópia para evitar race conditions
	servers := make(map[string]*entities.ServerInfo)
	for k, v := range d.KnownServers {
		servers[k] = v
	}
	return servers
}


func (d *Discovery) GetMemberlistNodes() []*memberlist.Node {
	return d.memberlist.Members()
}

