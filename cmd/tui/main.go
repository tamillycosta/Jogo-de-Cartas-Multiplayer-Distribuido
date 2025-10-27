package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	// Ajuste os caminhos de importação para o nome do seu módulo
	"Jogo-de-Cartas-Multiplayer-Distribuido/cmd/tui/comm"
	"Jogo-de-Cartas-Multiplayer-Distribuido/cmd/tui/models"
)

// AppModel é o modelo principal que gerencia as telas E a conexão
type AppModel struct {
	client       *comm.Client // Nosso cliente WebSocket
	currentModel tea.Model
	width        int
	height       int
	status       string // Status global (conectando, conectado, erro)
}

func initialModel() AppModel {
	// Conecta em qualquer servidor, o backend cuida do resto
	// (Baseado no seu docker-compose.yml)
	client := comm.NewClient("ws://localhost:8080/ws")

	return AppModel{
		client:       client,
		currentModel: models.NewMenu(), //
		width:        80,
		height:       24,
		status:       "Conectando...",
	}
}

func (m AppModel) Init() tea.Cmd {
	// Combina o Init do modelo atual com o início da conexão
	return tea.Batch(m.currentModel.Init(), m.client.Connect())
}

// --- FUNÇÃO UPDATE MODIFICADA PARA O LOOP DE RE-ESCUTA ---
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// --- Ciclo de Vida da Conexão ---
	case comm.ConnectedMsg:
		m.status = fmt.Sprintf("Conectado! (ID: %s)", msg.ClientID)
		// Auto-subscribe no tópico de resposta de auth
		authTopic := fmt.Sprintf("auth.response.%s", msg.ClientID) //
		cmds = append(cmds,
			m.client.Subscribe(authTopic),
			m.client.Listen(), // Inicia a *primeira* escuta
		)

	case comm.ErrorMsg:
		m.status = fmt.Sprintf("Erro de Conexão: %v", msg.Err)
		// Para de escutar em caso de erro

	// --- Roteamento de Ações dos Modelos ---
	case models.DoRegisterMsg:
		m.status = "Registrando..."
		// O AppModel é quem publica, usando seu client
		cmds = append(cmds, m.client.Publish(
			"auth.create_account", // Tópico de cadastro
			map[string]string{"username": msg.Username}, // Payload
		))

		case models.DoLoginMsg:
			m.status = "Fazendo login..."
			cmds = append(cmds, m.client.Publish(
				"auth.login",
				map[string]string{"username": msg.Username},
			))

	// --- Troca de Telas ---
	case models.SwitchToLoginMsg:
		m.currentModel = models.NewLogin()
		m.currentModel, cmd = m.currentModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		cmds = append(cmds, cmd)

	case models.SwitchToCadastroMsg:
		m.currentModel = models.NewCadastro()
		m.currentModel, cmd = m.currentModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		cmds = append(cmds, cmd)

	case models.SwitchToMenuMsg:
		m.currentModel = models.NewMenu() //
		m.currentModel, cmd = m.currentModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		cmds = append(cmds, cmd)

	case models.SwitchToLobbyMsg:
		m.currentModel = models.NewLobby()
		m.currentModel, cmd = m.currentModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		cmds = append(cmds, cmd)

	// --- MENSAGENS VINDAS DO LISTENER (LOOP) ---
	case comm.AuthResponseMsg:
		// 1. Passa a mensagem para o modelo filho
		m.currentModel, cmd = m.currentModel.Update(msg)
		cmds = append(cmds, cmd)
		// 2. RE-INICIA A ESCUTA para a próxima mensagem
		cmds = append(cmds, m.client.Listen())

	case comm.NoOpMsg:
		// 1. Mensagem ignorada (não passa para o filho)
		// 2. Apenas RE-INICIA A ESCUTA
		cmds = append(cmds, m.client.Listen())

	// --- Roteamento Padrão ---
	default:
		// Se for tea.KeyMsg, tea.WindowSizeMsg, etc.
		// passa para o modelo filho.
		if wsMsg, ok := msg.(tea.WindowSizeMsg); ok {
			m.width = wsMsg.Width
			m.height = wsMsg.Height
		}

		m.currentModel, cmd = m.currentModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// --- FIM DA MODIFICAÇÃO ---

func (m AppModel) View() string {
	// Delega a view para o modelo atual
	// Poderia adicionar um rodapé com m.status aqui se quisesse
	return m.currentModel.View()
}

func main() {
	// Habilitar log em arquivo para depuração
	f, err := tea.LogToFile("tui.log", "debug")
	if err != nil {
		fmt.Println("Erro ao criar log:", err)
		os.Exit(1)
	}
	defer f.Close()

	log.Println("Iniciando TUI...")

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Println("Erro ao rodar programa:", err)
		fmt.Println("Erro:", err)
		os.Exit(1)
	}
}