package main

import "fmt"

func (c *Client) showMenu() {
	// Banner padronizado com 42 caracteres de largura interna
	fmt.Println("\n╔══════════════════════════════════════════╗")
	fmt.Println("║           🎴   MAGICARDS   🎴            ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	
	fmt.Printf("\n 👤 Jogador: %s\n", c.username)
	
	fmt.Println("\n 📋 COMANDOS DISPONÍVEIS:")
	fmt.Println(" ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	fmt.Println("  🔹 LOBBY & CARTAS")
	fmt.Printf("  %-15s → %s\n", "pack (p)", "Abrir novo pacote de cartas")
	fmt.Printf("  %-15s → %s\n", "list (ls)", "Listar inventário (ou 'ls <user>')")
	fmt.Printf("  %-15s → %s\n", "trade (t)", "Trocar: t <meu_id> <user> <dele_id>")
	fmt.Printf("  %-15s → %s\n", "queue (q)", "Entrar na fila de partida")
	
	fmt.Println("\n  ⚔️  EM PARTIDA")
	fmt.Printf("  %-15s → %s\n", "card <n> (c)", "Jogar carta da mão (ex: 'c 0')")
	fmt.Printf("  %-15s → %s\n", "attack (a)", "Atacar o oponente")
	fmt.Printf("  %-15s → %s\n", "leave (l)", "Desistir da partida atual")
	
	fmt.Println("\n  ⚙️  SISTEMA")
	fmt.Printf("  %-15s → %s\n", "menu (m)", "Mostrar este menu")
	fmt.Printf("  %-15s → %s\n", "help (h)", "Ajuda detalhada")
	fmt.Printf("  %-15s → %s\n", "exit", "Sair do jogo")
	fmt.Println(" ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func (c *Client) showHelp() {
	fmt.Println("\n╔══════════════════════════════════════════╗")
	fmt.Println("║               📖  AJUDA  📖              ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	
	fmt.Println("\n 🎮 Fluxo do Jogo:")
	fmt.Println("   1. Use 'pack' para ganhar cartas se não tiver nenhuma.")
	fmt.Println("   2. Use 'queue' para encontrar um oponente.")
	fmt.Println("   3. Na partida, jogue uma carta ('c 0') para colocá-la em campo.")
	fmt.Println("   4. No próximo turno, use 'attack' para atacar a carta ou vida.")
	
	fmt.Println("\n 🤝 Sistema de Trocas (Trade):")
	fmt.Println("   • Para trocar, você precisa saber o ID da carta que quer.")
	fmt.Println("   • Use 'list <nome_amigo>' para ver as cartas dele e copiar o ID.")
	fmt.Println("   • Comando: trade <ID_SUA_CARTA> <NOME_AMIGO> <ID_CARTA_DELE>")
	
	fmt.Println("\n 💡 Atalhos:")
	fmt.Println("   • q = queue   |  c = card")
	fmt.Println("   • a = attack  |  t = trade")
	fmt.Println("   • p = pack    |  l = leave")
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}



// Lida com erro de servidor caído
func (c *Client) handleServerDown() {
	clearScreen()
	fmt.Println("\n╔══════════════════════════════════════════╗")
	fmt.Println("║       ⚠️  SERVIDOR INDISPONÍVEL ⚠️       ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println("\n ❌ O servidor do oponente caiu ou está inacessível")
	fmt.Println(" 🔌 A partida será encerrada automaticamente")
	fmt.Println("\n 💡 Você pode:")
	fmt.Println("   • Entrar na fila novamente (digite 'queue')")
	fmt.Println("   • Sair do jogo (digite 'exit')")
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	// Limpa estado da partida
	c.inMatch = false
	c.matchID = ""
	c.currentTurn = ""
	c.turnNumber = 0
}

func printLogo(){
	art := `
	███╗   ███╗ █████╗  ██████╗ ██╗ ██████╗ █████╗ ██████╗ ██████╗ ███████╗
	████╗ ████║██╔══██╗██╔════╝ ██║██╔════╝██╔══██╗██╔══██╗██╔══██╗██╔════╝
	██╔████╔██║███████║██║  ███╗██║██║     ███████║██████╔╝██║  ██║███████╗
	██║╚██╔╝██║██╔══██║██║   ██║██║██║     ██╔══██║██╔══██╗██║  ██║╚════██║
	██║ ╚═╝ ██║██║  ██║╚██████╔╝██║╚██████╗██║  ██║██║  ██║██████╔╝███████║
	╚═╝     ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝ ╚══════╝
	`

	fmt.Println(art)
}