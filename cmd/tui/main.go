package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"Jogo-de-Cartas-Multiplayer-Distribuido/cmd/tui/comm"
	"Jogo-de-Cartas-Multiplayer-Distribuido/cmd/tui/models"
)

type AppModel struct {
	client       *comm.Client
	currentModel tea.Model
	width        int
	height       int
	status       string
	PlayerID     string
}

func initialModel() AppModel {
	client := comm.NewClient("ws://localhost:8080/ws")

	return AppModel{
		client:       client,
		currentModel: models.NewMenu(),
		width:        80,
		height:       24,
		status:       "Conectando...",
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(m.currentModel.Init(), m.client.Connect())
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// --- Ciclo de Vida da Conexão ---
	case comm.ConnectedMsg:
		m.status = fmt.Sprintf("Conectado! (ID: %s)", msg.ClientID)
		
		responseTopic := fmt.Sprintf("response.%s", msg.ClientID) // Tópico de Auth (provado pelo seu log)
		packageTopic := fmt.Sprintf("package.response.%s", msg.ClientID) // Tópico de Pacote (provado pelo backend)
		
		log.Printf("[AppModel] Inscrevendo-se em: %s e %s", responseTopic, packageTopic)

		// --- CORREÇÃO DO PÂNICO: Use tea.Sequence ---
		// Executa os comandos de *escrita* (Subscribe) em ordem,
		// e só no fim inicia a *leitura* (Listen).
		cmds = append(cmds, tea.Sequence(
			m.client.Subscribe(responseTopic),
			m.client.Subscribe(packageTopic),
			m.client.Listen(),
		))
		// --- FIM DA CORREÇÃO ---

	case comm.ErrorMsg:
		m.status = fmt.Sprintf("Erro de Conexão: %v", msg.Err)

	// --- Roteamento de Ações dos Modelos ---
	case models.DoRegisterMsg:
		m.status = "Registrando..."
		// --- CORREÇÃO DO TRAVAMENTO: Re-adiciona o Listen() ---
		// Publicar (escrita) e Escutar (leitura) em paralelo é SEGURO.
		cmds = append(cmds,
			m.client.Publish(
				"auth.create_account",
				map[string]string{"username": msg.Username},
			),
			m.client.Listen(), // <-- Essencial para não ficar "surdo"
		)
		// --- FIM DA CORREÇÃO ---

	case models.DoLoginMsg:
		m.status = "Fazendo login..."
		// --- CORREÇÃO DO TRAVAMENTO: Re-adiciona o Listen() ---
		cmds = append(cmds,
			m.client.Publish(
				"auth.login",
				map[string]string{"username": msg.Username},
			),
			m.client.Listen(), // <-- Essencial para não ficar "surdo"
		)
		// --- FIM DA CORREÇÃO ---

	case models.DoOpenPackageMsg:
		if m.PlayerID == "" {
			log.Println("[AppModel] Erro: Tentativa de abrir pacote sem PlayerID!")
			cmds = append(cmds, func() tea.Msg {
				return comm.PackageResponseMsg{Success: false, Error: "Erro: PlayerID não encontrado."}
			})
		} else {
			log.Printf("[AppModel] Publicando 'package.open_pack' para PlayerID: %s", m.PlayerID)
			// --- CORREÇÃO DO TRAVAMENTO: Re-adiciona o Listen() ---
			cmds = append(cmds,
				m.client.Publish(
					"package.open_pack",
					map[string]string{"player_id": m.PlayerID},
				),
				m.client.Listen(), // <-- Essencial para não ficar "surdo"
			)
			// --- FIM DA CORREÇÃO ---
		}

	// --- Troca de Telas ---
	// (Precisa de Listen() para iniciar o loop na nova tela)
	case models.SwitchToLoginMsg:
		m.currentModel = models.NewLogin()
		m.currentModel, cmd = m.currentModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		cmds = append(cmds, cmd)
		cmds = append(cmds, m.client.Listen())

	case models.SwitchToCadastroMsg:
		m.currentModel = models.NewCadastro()
		m.currentModel, cmd = m.currentModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		cmds = append(cmds, cmd)
		cmds = append(cmds, m.client.Listen())

	case models.SwitchToMenuMsg:
		m.currentModel = models.NewMenu()
		m.currentModel, cmd = m.currentModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		cmds = append(cmds, cmd)
		cmds = append(cmds, m.client.Listen())

	case models.SwitchToLobbyMsg:
		log.Printf("[AppModel] Trocando para Lobby.")
		m.currentModel = models.NewLobby()
		m.currentModel, cmd = m.currentModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		cmds = append(cmds, cmd)
		cmds = append(cmds, m.client.Listen())

	case models.SwitchToPackageOpeningMsg:
		log.Printf("[AppModel] Trocando para PackageOpening.")
		m.currentModel = models.NewPackageOpening()
		m.currentModel, cmd = m.currentModel.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		cmds = append(cmds, cmd)
		// O Init() da tela vai disparar DoOpenPackageMsg,
		// e o 'case DoOpenPackageMsg' (acima) já inclui o Listen().

	// --- MENSAGENS VINDAS DO LISTENER (LOOP) ---
	// (Estes reiniciam o loop de escuta)
	case comm.AuthResponseMsg:
		log.Printf("[AppModel] Recebida AuthResponseMsg. Repassando para o modelo filho...")
		
		if msg.Success && msg.PlayerID != "" {
			m.PlayerID = msg.PlayerID
			log.Printf("[AppModel] PlayerID armazenado: %s", m.PlayerID)
		}

		m.currentModel, cmd = m.currentModel.Update(msg)
		cmds = append(cmds, cmd)

		// Se o filho (login) falhou (cmd == nil), reinicia a escuta.
		// Se o filho (login) teve sucesso (cmd != nil),
		// o Tick/SwitchToLobbyMsg tratará de reiniciar a escuta.
		if cmd == nil {
			cmds = append(cmds, m.client.Listen())
		}

	case comm.PackageResponseMsg:
		log.Printf("[AppModel] Recebida PackageResponseMsg. Repassando para o modelo filho...")
		m.currentModel, cmd = m.currentModel.Update(msg)
		cmds = append(cmds, cmd)
		cmds = append(cmds, m.client.Listen()) // Continua escutando na tela de resultado

	case comm.NoOpMsg:
		log.Printf("[AppModel] NoOp recebido. Reiniciando escuta.")
		cmds = append(cmds, m.client.Listen()) // Continua o loop

	// --- Roteamento Padrão ---
	default:
		if wsMsg, ok := msg.(tea.WindowSizeMsg); ok {
			m.width = wsMsg.Width
			m.height = wsMsg.Height
		}

		m.currentModel, cmd = m.currentModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m AppModel) View() string {
	return m.currentModel.View()
}

func main() {
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