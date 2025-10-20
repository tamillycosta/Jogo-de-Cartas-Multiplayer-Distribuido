package models

import (
	"errors"
	"fmt"
	
	// Ajuste os caminhos de importação para o nome do seu módulo
	"Jogo-de-Cartas-Multiplayer-Distribuido/cmd/tui/comm" 

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type CadastroModel struct {
	textInput textinput.Model
	width     int
	height    int
	status    string // Para feedback: "Registrando..."
	err       error  // Para feedback de erro
}

func NewCadastro() CadastroModel {
	ti := textinput.New()
	ti.Placeholder = "username"
	ti.Focus()
	ti.Width = 30
	ti.CharLimit = 20

	return CadastroModel{
		textInput: ti,
	}
}

func (m CadastroModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CadastroModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			// Volta para o menu
			return m, func() tea.Msg { return SwitchToMenuMsg{} }

		case "enter":
			username := m.textInput.Value()
			if username == "" {
				return m, nil
			}
			// Envia a mensagem para o AppModel tratar
			m.status = fmt.Sprintf("Enviando '%s'...", username)
			m.err = nil
			return m, func() tea.Msg { return DoRegisterMsg{Username: username} }
		}

	// --- Respostas do Servidor (via AppModel) ---
	
	// Recebe a resposta de autenticação
	case comm.AuthResponseMsg:
		m.status = ""
		if msg.Success {
			m.status = fmt.Sprintf("Sucesso! %s", msg.Message)
			m.textInput.Reset()
		} else {
			m.err = errors.New(msg.Error)
		}

	// Recebe erros genéricos de conexão
	case comm.ErrorMsg:
		m.status = ""
		m.err = msg.Err

	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m CadastroModel) View() string {
	titulo := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		Width(37).
		Align(lipgloss.Center).
		Render("CADASTRO (via Pub/Sub)")

	campo := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Render(m.textInput.View())

	ajuda := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("enter: confirmar • esc: voltar")

	// Feedback
	var feedback string
	if m.status != "" {
		feedback = lipgloss.NewStyle().Foreground(lipgloss.Color("70")).Render(m.status) // Verde
	} else if m.err != nil {
		feedback = lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Render(m.err.Error()) // Vermelho
	}

	conteudo := lipgloss.JoinVertical(
		lipgloss.Left,
		titulo,
		"",
		campo,
		"",
		feedback,
		"",
		ajuda,
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		conteudo,
	)
}
