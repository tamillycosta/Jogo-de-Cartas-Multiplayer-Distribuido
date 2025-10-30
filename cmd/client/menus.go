package main

import "fmt"




func (c *Client) showMenu() {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║         🎴  MAGICARDS  🎴             ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Printf("\n👤 Jogador: %s\n", c.username)
	fmt.Println("\n📋 Comandos Disponíveis:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  pack (p)        → Abrir pacote de cartas")
	fmt.Println("  queue (q)       → Entrar na fila")
	fmt.Println("  card <n> (c)    → Jogar carta (ex: card 0)")
	fmt.Println("  attack (a)      → Atacar com carta ativa")
	fmt.Println("  menu (m)        → Mostrar este menu")
	fmt.Println("  help (h)        → Ajuda")
	fmt.Println("  exit            → Sair do jogo")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func (c *Client) showHelp() {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║              📖 AJUDA 📖              ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println("\n🎮 Como Jogar:")
	fmt.Println("   1. Digite 'queue' para procurar partida")
	fmt.Println("   2. Quando for seu turno, escolha uma carta")
	fmt.Println("   3. Use 'card <número>' para jogar")
	fmt.Println("   4. Use 'attack' para atacar o oponente")
	fmt.Println("   5. Reduza o HP do oponente a 0 para vencer!")
	fmt.Println("\n💡 Dicas:")
	fmt.Println("   • Você pode usar atalhos: q, c, a")
	fmt.Println("   • 'card 0' joga a primeira carta do deck")
	fmt.Println("   • Só pode atacar quando tiver carta na mão")
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func printBanner() {
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║                                        ║")
	fmt.Println("║       🎴  MAGICARDS CLIENT  🎴        ║")
	fmt.Println("║                                        ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()
}

// Lida com erro de servidor caído
func (c *Client) handleServerDown() {
	clearScreen()
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║      ⚠️  SERVIDOR INDISPONÍVEL  ⚠️     ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println("\n❌ O servidor do oponente caiu ou está inacessível")
	fmt.Println("🔌 A partida será encerrada automaticamente")
	fmt.Println("\n💡 Você pode:")
	fmt.Println("   • Entrar na fila novamente (digite 'queue')")
	fmt.Println("   • Sair do jogo (digite 'exit')")
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	// Limpa estado da partida
	c.inMatch = false
	c.matchID = ""
	c.currentTurn = ""
	c.turnNumber = 0
}
