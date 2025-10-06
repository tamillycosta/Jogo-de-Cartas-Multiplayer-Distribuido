package interfaces



// Representa todos os metodos implementados para comunicação dos servidores 
type Communication interface {
    AskForExistUsername(serverAddres string, port int) bool
    AskForNewPlayer(serverAddress string)
}