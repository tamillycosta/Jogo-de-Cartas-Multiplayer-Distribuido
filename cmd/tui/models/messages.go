package models

// Mensagens customizadas para trocar de tela
type SwitchToLoginMsg struct{}
type SwitchToCadastroMsg struct{}
type SwitchToMenuMsg struct{}
type SwitchToLobbyMsg struct{}

// --- NOVA MENSAGEM ---
// Mensagem para o AppModel trocar para a tela de abertura de pacote
type SwitchToPackageOpeningMsg struct{}

// Mensagem para o AppModel realizar o cadastro
type DoRegisterMsg struct {
	Username string
}

// Mensagem para o AppModel realizar o login
type DoLoginMsg struct {
	Username string
}

// --- NOVA MENSAGEM ---
// Mensagem para o AppModel enviar o pedido de abrir pacote via WS
type DoOpenPackageMsg struct{}

// --- NOVA MENSAGEM ---
// Mensagem que o AppModel envia para a tela (PackageOpeningModel) com o resultado
type PackageResponseMsg struct {
	Success bool
	Message string
	Error   string
}