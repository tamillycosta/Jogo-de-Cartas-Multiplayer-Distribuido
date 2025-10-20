package models

import (
	"errors"
	"fmt"
	"time" // <-- IMPORTAR time

	"Jogo-de-Cartas-Multiplayer-Distribuido/cmd/tui/comm"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type LoginModel struct {
	textInput textinput.Model
	width     int
	height    int
	status    string
	err       error
}

func NewLogin() LoginModel {
	ti := textinput.New()
	ti.Placeholder = "username"
	ti.Focus()
	ti.Width = 30
	ti.CharLimit = 20

	return LoginModel{
		textInput: ti,
	}
}

func (m LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			// Retorna a mensagem para o AppModel tratar
			return m, func() tea.Msg { return SwitchToMenuMsg{} }

		case "enter":
			username := m.textInput.Value()
			if username == "" {
				return m, nil
			}
			
			// --- INÍCIO DA MUDANÇA ---
			m.status = fmt.Sprintf("Enviando login para '%s'...", username)
			m.err = nil
			// Envia a mensagem para o AppModel tratar
			return m, func() tea.Msg { return DoLoginMsg{Username: username} }
			// --- FIM DA MUDANÇA ---
		}
		
	// --- ADICIONAR ESTES CASES ---
	// Recebe a resposta de autenticação
	case comm.AuthResponseMsg:
		m.status = ""
		if msg.Success {
			m.status = fmt.Sprintf("Sucesso! %s", msg.Message)
			m.textInput.Reset()
			
			// Retorna ao menu após 2 segundos
			cmd = tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return SwitchToMenuMsg{}
			})
			
		} else {
			m.err = errors.New(msg.Error)
		}
		return m, cmd

	// Recebe erros genéricos de conexão
	case comm.ErrorMsg:
		m.status = ""
		m.err = msg.Err
	// --- FIM DA ADIÇÃO ---

	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m LoginModel) View() string {
	titulo := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		Width(37).
		Align(lipgloss.Center).
		Render("LOGIN (via Pub/Sub)")

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
		// Verde para sucesso, Amarelo para "Enviando..."
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("220")) // Amarelo
		if m.err == nil && m.status != fmt.Sprintf("Enviando login para '%s'...", m.textInput.Value()) {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("70")) // Verde
		}
		feedback = style.Render(m.status)

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