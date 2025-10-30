package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LobbyModel struct {
	cursor int
	opcoes []string
	width  int
	height int
}

func NewLobby() LobbyModel {
	return LobbyModel{
		cursor: 0,
		// Adiciona "Voltar ao Menu" como uma opção
		opcoes: []string{"Entrar na Partida", "Abrir Pacote", "Trocar Cartas", "Voltar ao Menu"},
	}
}

func (m LobbyModel) Init() tea.Cmd {
	return nil
}

func (m LobbyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.opcoes)-1 {
				m.cursor++
			}

		case "enter":
			// Por enquanto, apenas o "Voltar ao Menu" faz algo
			switch m.cursor {
			case 0: // Entrar na Partida
				// Não implementado
				return m, nil
			case 1: // Abrir Pacote
				return m, func() tea.Msg { return SwitchToPackageOpeningMsg{} }
			case 2: // Trocar Cartas
				// Não implementado
				return m, nil
			case 3: // Voltar ao Menu
				return m, func() tea.Msg { return SwitchToMenuMsg{} }
			}
		case "esc": // Permite voltar com Esc também
			return m, func() tea.Msg { return SwitchToMenuMsg{} }
		}
	}
	return m, nil
}

func (m LobbyModel) View() string {
	titulo := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")). // Cor roxa
		Width(37).
		Align(lipgloss.Center).
		Render("LOBBY")

	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170"))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	menuStyle := lipgloss.NewStyle().Width(25).Align(lipgloss.Center) // Aumenta a largura
	var menu string
	for i, opcao := range m.opcoes {
		if m.cursor == i {
			menu += menuStyle.Render(selectedStyle.Render("→ "+opcao)) + "\n"
		} else {
			menu += menuStyle.Render(normalStyle.Render("  "+opcao)) + "\n"
		}
	}

	conteudo := lipgloss.JoinVertical(
		lipgloss.Center,
		titulo,
		"",
		menu,
		"",
		helpStyle.Render("↑/↓: navegar • enter: selecionar • esc: voltar menu • q: sair"),
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		conteudo,
	)
}