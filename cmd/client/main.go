package main

import (
	"bufio"
	
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"math/rand"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn       *websocket.Conn
	clientID   string
	matchID    string
	playerID   string
	username   string
	currentTurn string
	turnNumber  int
	inMatch     bool
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	
	clearScreen()
	printBanner()
	// Lista de servidores possíveis
	servers := []string{
		"ws://localhost:8080/ws",
		"ws://localhost:8081/ws",
		"ws://localhost:8082/ws",
	}

	// Sorteia um servidor
	selectedServer := servers[rand.Intn(len(servers))]

	fmt.Println("🌐 Conectando ao servidor...")

	conn, _, err := websocket.DefaultDialer.Dial(selectedServer, nil)
	if err != nil {
		log.Fatal("❌ Erro ao conectar:", err)
	}

	client := &Client{conn: conn}

	// Recebe client_id
	var msg map[string]interface{}
	conn.ReadJSON(&msg)
	client.clientID = msg["client_id"].(string)
	fmt.Printf("✅ Conectado! ID: %s\n", client.clientID)

	// Goroutine para receber mensagens
	go client.listen()

	for {
		clearScreen()
		fmt.Println("🔐 Sistema de Autenticação")
		fmt.Println("1) Login")
		fmt.Println("2) Criar Conta")
		fmt.Print("Escolha: ")
	
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
	
		if choice == "1" || choice == "2" {
			if choice == "2" {
				fmt.Print("\n🆕 Digite o nome de usuário para criar conta: ")
				username, _ := reader.ReadString('\n')
				username = strings.TrimSpace(username)
	
				client.subscribe("response." + client.clientID)
				client.subscribe("package.response." + client.clientID)
				client.createAccount(username)
				
				fmt.Println("\n⏳ Criando conta...")
				time.Sleep(1 * time.Second)
			}
	
			fmt.Print("\n🎮 Digite seu nome de usuário para login: ")
			client.subscribe("response." + client.clientID)
			client.subscribe("package.response." + client.clientID)
			username, _ := reader.ReadString('\n')
			username = strings.TrimSpace(username)
	
			client.username = username
			client.login(username)
			break
		}
	
		fmt.Println("❌ Opção inválida!")
		time.Sleep(1 * time.Second)
	}

	
	
	// Aguarda um pouco para processar login
	fmt.Println("\n⏳ Processando login...")
	time.Sleep(1 * time.Second)

	clearScreen()
	client.showMenu()

	for {
		if client.inMatch {
			fmt.Print("\n⚔️ > ")
		} else {
			fmt.Print("\n🎯 > ")
		}
		
		cmdLine, _ := reader.ReadString('\n')
		cmdLine = strings.TrimSpace(cmdLine)
		parts := strings.Split(cmdLine, " ")

		if len(parts) == 0 {
			continue
		}

		switch parts[0] {

		case "pack", "p":
			if client.inMatch {
				fmt.Println("⚠️ Você não pode abrir pacotes durante uma partida!")
				continue
			}
			client.openPack()
			
		case "queue", "q":
			if client.inMatch {
				fmt.Println("⚠️ Você já está em uma partida!")
				continue
			}
			client.joinQueue()
			
		case "card", "c":
			if !client.inMatch {
				fmt.Println("⚠️ Você não está em uma partida!")
				continue
			}
			if len(parts) < 2 {
				fmt.Println("❌ Uso: card <index> (ex: card 0)")
				continue
			}
			client.playCard(parts[1])

		case "leave", "l":{
			if !client.inMatch {
				fmt.Println("⚠️ Você não está em uma partida!")
				continue
			}
			client.leaveMatch()
		}
			
		case "attack", "a":
			if !client.inMatch {
				fmt.Println("⚠️ Você não está em uma partida!")
				continue
			}
			client.attack()
			
		case "menu", "m":
			clearScreen()
			client.showMenu()
			
		case "exit", "quit":
			fmt.Println("\n👋 Até logo!")
			client.conn.Close()
			return
			
		case "help", "h":
			client.showHelp()
			
		default:
			fmt.Println("❌ Comando inválido. Digite 'help' para ver os comandos.")
		}
	}
}

func (c *Client) listen() {
	for {
		var msg map[string]interface{}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("\n🔌 Conexão perdida:", err)
			os.Exit(1)
		}

		msgType, _ := msg["type"].(string)
		if msgType == "" {
			msgType, _ = msg["topic"].(string)
		}

		switch msgType {
		case "response":
			c.handleResponse(msg)

		case "package.response":
			c.handlePackageResponse(msg)

		case "queue_joined":
			queueSize, _ := msg["queue_size"].(float64)
			fmt.Printf("\nVocê entrou na fila! (jogadores: %.0f)\n", queueSize)
			fmt.Println(" Procurando oponente...")

		case "match_found":
			c.handleMatchFound(msg)

		case "game_update":
			c.handleGameUpdate(msg)

		case "subscribed":
			
		    
		case "error":
			
			if c.inMatch && c.isServerDownError(msg) {
				c.handleServerDown()
			} else {
				errMsg, _ := msg["error"].(string)
				fmt.Printf("\n Erro: %s\n", errMsg)
			}

		default:
			log.Printf("Mensagem não tratada: %s", msgType)
		}
	}
}


func (c *Client) handleResponse(msg map[string]interface{}) {
    data, ok := msg["data"].(map[string]interface{})
    if !ok {
        return
    }

    respType, _ := data["type"].(string)
    success, _ := data["success"].(bool)

    switch respType {

    case "account_created":
        if success {
            message, _ := data["message"].(string)
            fmt.Printf("\n %s\n", message)
            fmt.Println("Agora você pode fazer login!")
        } else {
            errMsg, _ := data["error"].(string)
            fmt.Printf("\n Erro ao criar conta: %s\n", errMsg)
        }

    case "login_response":
        if success {
            fmt.Println(" Login realizado com sucesso!")
            if player, ok := data["player"].(map[string]interface{}); ok {
                if playerID, ok := player["id"].(string); ok {
                    c.playerID = playerID
                }
            }
        } else {
            errMsg, _ := data["error"].(string)
            fmt.Printf("\n Erro no login: %s\n", errMsg)
        }
    }
}


func (c *Client) handlePackageResponse(msg map[string]interface{}) {
	data, ok := msg["data"].(map[string]interface{})
	if !ok {
		fmt.Println(" Resposta inválida do servidor")
		return
	}

	success, _ := data["success"].(bool)
	
	if success {
		message, _ := data["message"].(string)
		fmt.Printf("\n✨ %s\n", message)
		fmt.Println("🎴 Novas cartas adicionadas ao seu deck!")
	} else {
		errorMsg, _ := data["error"].(string)
		fmt.Printf("\n Erro ao abrir pacote: %s\n", errorMsg)
	}
}


func (c *Client) isServerDownError(msg map[string]interface{}) bool {
	errMsg, ok := msg["error"].(string)
	if !ok {
		return false
	}
	
	// Lista de indicadores de servidor caído
	serverDownIndicators := []string{
		"dial tcp",
		"connection refused",
		"server misbehaving",
		"no such host",
		"connection reset",
		"timeout",
		"EOF",
	}
	
	errMsgLower := strings.ToLower(errMsg)
	for _, indicator := range serverDownIndicators {
		if strings.Contains(errMsgLower, indicator) {
			return true
		}
	}
	
	return false
}


