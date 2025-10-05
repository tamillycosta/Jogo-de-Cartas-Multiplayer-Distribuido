	package discovery

	import (
		"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
		"encoding/json"
		"fmt"
		"github.com/hashicorp/memberlist"
	)

	// estrutura que implementa a interface memberlist.EventDelegate.
	// Ela é responsável por receber notificações de eventos de entrada e saída de nós (servidores)
	// no cluster, permitindo que o serviço Discovery mantenha uma lista atualizada dos servidores conhecidos.
	type eventDelegate struct {
		discovery *Discovery
	}

	// Chamado quando ALGUÉM ENTRA no cluster
	func (e *eventDelegate) NotifyJoin(node *memberlist.Node) {
		// Decodifica metadata do servidor
		var info entities.ServerInfo
		if err := json.Unmarshal(node.Meta, &info); err != nil {
			fmt.Printf("[Discovery] Erro ao decodificar metadata: %v\n", err)
			return
		}

		// Ignora a si mesmo
		if info.ID == e.discovery.MyInfo.ID {
			return
		}

		e.discovery.mu.Lock()
		e.discovery.KnownServers[info.ID] = &info
		e.discovery.mu.Unlock()
		
		fmt.Printf("[Discovery] ✓ Servidor entrou: %s (%s:%d) - Region: %s\n",
			info.ID, info.Address, info.Port, info.Region)
	}

// Chamado quando ALGUÉM SAI do cluster
func (e *eventDelegate) NotifyLeave(node *memberlist.Node) {
		var info entities.ServerInfo
		if err := json.Unmarshal(node.Meta, &info); err != nil {
			return
		}

		e.discovery.mu.Lock()
		delete(e.discovery.KnownServers, info.ID)
		e.discovery.mu.Unlock()

		fmt.Printf("[Discovery] ✗ Servidor saiu: %s\n", info.ID)
	}

// n é ultilizado 
	func (e *eventDelegate) NotifyUpdate(node *memberlist.Node) {
		
	}