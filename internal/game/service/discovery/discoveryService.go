// implementa a descoberta automática de servidores
// em uma rede usando a biblioteca HashiCorp memberlist.
// Ele permite que servidores detectem uns aos outros, mantenham
// uma lista de servidores conhecidos (KnownServers) e se conectem
// a seeds quando disponíveis.


package discovery

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
	"sync"
	"github.com/hashicorp/memberlist"
	"encoding/json"
	"fmt"
	"time"
	"log"
	"io"
)

// Componente respossavel por descobrir novos servidores na rede 
type Discovery struct {
	MyInfo       *entities.ServerInfo
	KnownServers map[string]*entities.ServerInfo
	mu           sync.RWMutex
	memberlist   *memberlist.Memberlist
}


//cria uma nova instância de Discovery, inicializa o memberlist
// e configura eventos para gerenciar entradas e saídas de servidores.
// bindPort define a porta usada para gossip, e seedAddrs é a lista
// de servidores iniciais para tentar se conectar.
func New(myInfo *entities.ServerInfo, bindPort int, seedAddrs []string) (*Discovery, error) {
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

	// espera por outros servidores 
	if len(seedAddrs) > 0 {
		go d.connectToSeeds(seedAddrs) 
	} else {
		fmt.Printf("[Discovery] Nenhum seed configurado - esperando outros se conectarem\n")
	}

	return d, nil
}


// Tenta conectar este servidor a uma lista de seeds
// Se não conseguir, o servidor roda isolado.
func (d *Discovery) connectToSeeds(seedAddrs []string) {
	maxRetries := 5
	retryDelay := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("[Discovery] Tentativa %d/%d de conectar aos seeds: %v\n", 
			attempt, maxRetries, seedAddrs)
		
		n, err := d.memberlist.Join(seedAddrs)
		
		if err == nil && n > 0 {
			fmt.Printf("[Discovery] ✓ Conectado a %d servidor(es)\n", n)
			return
		}
		
		if attempt < maxRetries {
			fmt.Printf("[Discovery] Falha ao conectar, tentando novamente em %v...\n", retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2  
		}
	}
	
	fmt.Printf("[Discovery] ⚠ Não foi possível conectar aos seeds. Servidor rodando isolado.\n")
}


// GetMemberCount retorna número total de membros
func (d *Discovery) GetMemberCount() int {
	return d.memberlist.NumMembers()
}


// Seta configurações iniciais do member list
func setMemberlistConfig(myInfo entities.ServerInfo, bindPort int,config *memberlist.Config){
	config.Logger = log.New(io.Discard, "", 0)
	config.Name = myInfo.ID
	config.BindPort = bindPort
	config.AdvertisePort = bindPort
	
}


//Lista todos os servidores conhecidos 
func (d *Discovery) GetServers() []*entities.ServerInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	servers := make([]*entities.ServerInfo, 0, len(d.KnownServers))
	for _, s := range d.KnownServers {
		servers = append(servers, s)
	}
	return servers
}