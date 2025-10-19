package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MenuModel struct {
	cursor int
	opcoes []string
	width  int
	height int
}

func NewMenu() MenuModel {
	return MenuModel{
		cursor: 0,
		opcoes: []string{"Login", "Cadastro", "Sair"},
	}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			// Retorna mensagens personalizadas para trocar de tela
			// Note que agora elas vêm do pacote 'models'
			switch m.cursor {
			case 0:
				// Retorna a mensagem para o AppModel tratar
				return m, func() tea.Msg { return SwitchToLoginMsg{} }
			case 1:
				// Retorna a mensagem para o AppModel tratar
				return m, func() tea.Msg { return SwitchToCadastroMsg{} }
			case 2:
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m MenuModel) View() string {
	banner := `
 ███╗   ███╗ █████╗ ██████╗ ██╗ ██████╗ █████╗ ██████╗ ██████╗ ███████╗
 ████╗ ████║██╔══██╗██╔════╝ ██║██╔════╝██╔══██╗██╔══██╗██╔══██╗██╔════╝
 ██╔████╔██║███████║██║  ███╗██║██║     ███████║██████╔╝██║  ██║███████╗
 ██║╚██╔╝██║██╔══██║██║   ██║██║██║     ██╔══██║██╔══██╗██║  ██║╚════██║
 ██║ ╚═╝ ██║██║  ██║╚██████╔╝██║╚██████╗██║  ██║██║  ██║██████╔╝███████║
 ╚═╝     ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝ ╚══════╝
`

	bannerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170"))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	menuStyle := lipgloss.NewStyle().Width(20).Align(lipgloss.Center)
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
		bannerStyle.Render(banner),
		"",
		menu,
		"",
		helpStyle.Render("↑/↓: navegar • enter: selecionar • q: sair"),
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		conteudo,
	)
}

// REMOVIDO: As definições de SwitchToLoginMsg, etc., agora
// estão em 'models/messages.go'