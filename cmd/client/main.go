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
		printLogo()
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
			client.subscribe("trade.response." + client.clientID)
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
		client.printPrompt()
		
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

		case "list", "ls", "inv":
			targetUser := ""
			if len(parts) > 1 {
				targetUser = parts[1]
			}
			client.listCards(targetUser)

		case "trade", "t":
			if client.inMatch {
        		fmt.Println("⚠️ Você não pode trocar cartas durante uma partida!")
        		continue
    		}

			// Valida argumentos: trade <uuid> <nome> <uuid>
			if len(parts) < 3 {
				fmt.Println("❌ Uso correto: give <ID_DA_SUA_CARTA> <NOME_DO_JOGADOR> <ID_DA_CARTA_DELE>")
				fmt.Println("   Dica: Use o comando 'list' para copiar o ID da carta.")
				continue
			}

			myCardUUID := parts[1]
			targetUser := parts[2]
			wantedCardUUID := parts[3]

			client.tradeCard(myCardUUID, targetUser, wantedCardUUID)
			
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

		case "trade.response":
			c.handleTradeResponse(msg)

		case "inventory_list":
			c.handleInventoryList(msg)
			c.printPrompt()

		case "response":
			c.handleResponse(msg)

		case "package.response":
			c.handlePackageResponse(msg)
			c.printPrompt()

		case "queue_joined":
			queueSize, _ := msg["queue_size"].(float64)
			fmt.Printf("\nVocê entrou na fila! (jogadores: %.0f)\n", queueSize)
			fmt.Println(" Procurando oponente...")
			c.printPrompt()

		case "match_found":
			c.handleMatchFound(msg)
			c.printPrompt()

		case "game_update":
			c.handleGameUpdate(msg)
			c.printPrompt()

		case "subscribed":
			
		    
		case "error":
			
			if c.inMatch && c.isServerDownError(msg) {
				c.handleServerDown()
			} else {
				errMsg, _ := msg["error"].(string)
				fmt.Printf("\n Erro: %s\n", errMsg)
				c.printPrompt()
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

func (c *Client) handleInventoryList(msg map[string]interface{}) {
    var cards []interface{}
    if rawCards, ok := msg["cards"].([]interface{}); ok {
        cards = rawCards
    }

    // Tenta ler o nome do usuário alvo (se houver)
    targetUser, _ := msg["target_username"].(string)

    // Exibe o título apropriado
    if targetUser != "" {
        fmt.Printf("\n📦 INVENTÁRIO DE %s\n", targetUser)
    } else {
        fmt.Println("\n📦 SEU INVENTÁRIO")
    }

    // Verifica se está vazio com mensagens personalizadas
    if len(cards) == 0 {
        if targetUser != "" {
            fmt.Printf("\n⚠️  %s não possui cartas visíveis no momento.\n", targetUser)
        } else {
            fmt.Println("\n⚠️  Você não possui cartas no momento. Abra pacotes para começar!")
        }
        return
    }

    // Configuração de larguras fixas para garantir alinhamento
    const (
        wIdx    = 3
        wName   = 22
        wRarity = 10 
        wPower  = 5
        wUUID   = 36 
    )

    // Função auxiliar para desenhar linhas horizontais
    printLine := func(start, mid, end, fill string) {
        fmt.Print(start)
        fmt.Print(strings.Repeat(fill, wIdx+2))
        fmt.Print(mid)
        fmt.Print(strings.Repeat(fill, wName+2))
        fmt.Print(mid)
        fmt.Print(strings.Repeat(fill, wRarity+2))
        fmt.Print(mid)
        fmt.Print(strings.Repeat(fill, wPower+2))
        fmt.Print(mid)
        fmt.Print(strings.Repeat(fill, wUUID+2))
        fmt.Println(end)
    }

    // Topo da tabela
    fmt.Println()
    printLine("╔", "╦", "╗", "═")

    // Cabeçalho
    fmt.Printf("║ %-*s ║ %-*s ║ %-*s ║ %-*s ║ %-*s ║\n",
        wIdx, "ID",
        wName, "Nome",
        wRarity, "Raridade",
        wPower, "Pwr",
        wUUID, "UUID (Para Troca)")

    // Separador
    printLine("╠", "╬", "╣", "═")

    // Linhas de dados
    for _, item := range cards {
        cardMap := item.(map[string]interface{})

        // Trata o nome para não quebrar a tabela
        name := cardMap["name"].(string)
        if len(name) > wName {
            name = name[:wName-3] + "..."
        }

        rarity := cardMap["rarity"].(string)
        power := cardMap["power"].(float64)
        uuid := cardMap["id"].(string)

        fmt.Printf("║ %-*d ║ %-*s ║ %-*s ║ %-*.0f ║ %-*s ║\n",
            wIdx, int(cardMap["index"].(float64)),
            wName, name,
            wRarity, rarity,
            wPower, power,
            wUUID, uuid)
    }

    // Rodapé
    printLine("╚", "╩", "╝", "═")
    
    // Dica contextual
    if targetUser != "" {
        fmt.Println("💡 Dica: Copie o UUID da carta que você quer e use no comando 'trade'.")
    } else {
        fmt.Println("💡 Dica: Copie o UUID da carta que você quer ofertar no comando 'trade'.")
    }
}

func (c *Client) handleTradeResponse(msg map[string]interface{}) {
    data, ok := msg["data"].(map[string]interface{})
    if !ok { 
        return 
    }

    success, _ := data["success"].(bool)
    
    if success {
        message, _ := data["message"].(string)
        fmt.Println("\n✅ SUCESSO NA TRANSFERÊNCIA!")
        fmt.Printf("   %s\n", message)
    } else {
        errorMsg, _ := data["error"].(string)
        fmt.Println("\n❌ FALHA NA TRANSFERÊNCIA")
        fmt.Printf("   Erro: %s\n", errorMsg)
    }
    fmt.Print("\n🎯 > ") // Restaura o prompt
}

func (c *Client) printPrompt() {
	if c.inMatch {
		fmt.Print("\n⚔️ > ")
	} else {
		fmt.Print("\n🎯 > ")
	}
}