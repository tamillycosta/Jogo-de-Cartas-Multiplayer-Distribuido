	package discovery

	import (
		"Jogo-de-Cartas-Multiplayer-Distribuido/internal/shared/entities"
		"encoding/json"
		"log"
		"github.com/hashicorp/memberlist"
	)

	// estrutura que implementa a interface memberlist.EventDelegate.
	// Ela é responsável por receber notificações de eventos de entrada e saída de nós (servidores)
	// no cluster, permitindo que o serviço Discovery mantenha uma lista atualizada dos servidores conhecidos.
	
type eventDelegate struct {
	discovery *Discovery
}

func (e *eventDelegate) NotifyJoin(node *memberlist.Node) {
	log.Printf("🤝 [Discovery] Servidor entrou no cluster: %s (%s)", node.Name, node.Address())

	// Adiciona aos servidores conhecidos
	if node.Name != e.discovery.MyInfo.ID {
		var serverInfo entities.ServerInfo
		if err := json.Unmarshal(node.Meta, &serverInfo); err == nil {
			e.discovery.mu.Lock()
			e.discovery.KnownServers[node.Name] = &serverInfo
			e.discovery.mu.Unlock()
			
			log.Printf("✅ [Discovery] %s adicionado aos servidores conhecidos", node.Name)
		}
	}
}

func (e *eventDelegate) NotifyLeave(node *memberlist.Node) {
	log.Printf("👋 [Discovery] Servidor saiu do cluster: %s", node.Name)

	e.discovery.mu.Lock()
	delete(e.discovery.KnownServers, node.Name)
	e.discovery.mu.Unlock()
}

func (e *eventDelegate) NotifyUpdate(node *memberlist.Node) {
	log.Printf("🔄 [Discovery] Servidor atualizado: %s", node.Name)
}
