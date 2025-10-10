package discovery

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/util"
	"encoding/json"
	"fmt"
	"io"
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


func SetUpDiscovery(myInfo *entities.ServerInfo) (*Discovery, error){
	gossipPort := util.GetPortFromEnv("GOSSIP_PORT", 7947)
	return  New(myInfo, gossipPort)
}

func New(myInfo *entities.ServerInfo, bindPort int) (*Discovery, error) {
	d := &Discovery{
		MyInfo:       myInfo,
		KnownServers: make(map[string]*entities.ServerInfo),
	}

	config := memberlist.DefaultLocalConfig()
	setMemberlistConfig(*myInfo, bindPort, config)

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

	fmt.Printf("[Discovery] Memberlist iniciado: %s na porta %d\n", myInfo.ID, bindPort)

	// 🔥 inicia descoberta automática (sem seeds)
	go d.startAutoDiscovery(bindPort)

	return d, nil
}

func (d *Discovery) startAutoDiscovery(gossipPort int) {
	const broadcastPort = 9000

	// cache pra evitar flood
	lastSeen := make(map[string]time.Time)
	const floodDelay = 10 * time.Second

	// 1️⃣ goroutine para escutar broadcasts
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

			if msg.ID == d.MyInfo.ID {
				continue // ignora a si mesmo
			}

			key := fmt.Sprintf("%s-%s", msg.ID, remoteAddr.IP.String())

			d.mu.Lock()
			// evita flood: se já vimos há pouco tempo, ignora
			if t, ok := lastSeen[key]; ok && time.Since(t) < floodDelay {
				d.mu.Unlock()
				continue
			}
			lastSeen[key] = time.Now()

			if _, ok := d.KnownServers[msg.ID]; !ok {
				addr := fmt.Sprintf("%s:%d", remoteAddr.IP.String(), gossipPort)
				if _, err := d.memberlist.Join([]string{addr}); err == nil {
					d.KnownServers[msg.ID] = &msg
					log.Printf("[Discovery] 📡 Detectado servidor %s em %s", msg.ID, addr)
				}
			}
			d.mu.Unlock()
		}
	}()

	// 2️⃣ goroutine para enviar broadcasts
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
			conn.Write(data)
			time.Sleep(3 * time.Second)
		}
	}()
}



func setMemberlistConfig(myInfo entities.ServerInfo, bindPort int, config *memberlist.Config) {
	config.Logger = log.New(io.Discard, "", 0)
	config.Name = myInfo.ID
	config.BindPort = bindPort
	config.AdvertisePort = bindPort
}
