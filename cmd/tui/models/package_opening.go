package models

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/cmd/tui/comm" // Importa as mensagens de comunicação
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Define os estados da tela
type packageState int

const (
	stateOpening packageState = iota // Mostrando spinner
	stateResult                      // Mostrando sucesso ou erro
)

type PackageOpeningModel struct {
	spinner spinner.Model
	state   packageState
	message string // Mensagem de sucesso ou erro
	width   int
	height  int
}

func NewPackageOpening() PackageOpeningModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205")) // Cor do spinner

	return PackageOpeningModel{
		spinner: s,
		state:   stateOpening,
		message: "Abrindo pacote...",
	}
}

// Init é chamado quando a tela é criada.
// Inicia o spinner E envia a mensagem para o AppModel (main.go)
// solicitando a abertura do pacote.
func (m PackageOpeningModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg { return DoOpenPackageMsg{} }, // Envia a mensagem para o AppModel
	)
}

func (m PackageOpeningModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		// Se estivermos no estado de resultado, permite voltar ao lobby
		case "enter", "esc":
			if m.state == stateResult {
				return m, func() tea.Msg { return SwitchToLobbyMsg{} }
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	// Mensagem recebida do AppModel (vinda do WebSocket)
	case comm.PackageResponseMsg:
		m.state = stateResult // Muda para o estado de resultado
		if msg.Success {
			m.message = "✅ Sucesso! " + msg.Message
		} else {
			m.message = "❌ Erro: " + msg.Error
		}
		return m, nil // Para o spinner

	// Atualiza o spinner
	default:
		if m.state == stateOpening {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m PackageOpeningModel) View() string {
	var s string

	if m.state == stateOpening {
		s = m.spinner.View() + " " + m.message
	} else {
		// Define o estilo da mensagem de resultado
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("160")) // Vermelho (Erro)

		// <--- 2. ALTERE lipgloss.HasPrefix PARA strings.HasPrefix ---
		if strings.HasPrefix(m.message, "✅") {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("70")) // Verde (Sucesso)
		}

		s = style.Render(m.message)
		s += "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("(Pressione Enter ou Esc para voltar ao Lobby)")
	}

	// Centraliza a visão
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(s),
	)
}