package models

// Mensagens customizadas para trocar de tela
type SwitchToLoginMsg struct{}
type SwitchToCadastroMsg struct{}
type SwitchToMenuMsg struct{}

// Mensagem para o AppModel realizar o cadastro
type DoRegisterMsg struct {
	Username string
}

// Mensagem para o AppModel realizar o login
type DoLoginMsg struct {
	Username string
}

type SwitchToLobbyMsg struct{}