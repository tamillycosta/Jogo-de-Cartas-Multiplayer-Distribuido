package main

import (
	"bufio"
	"fmt"

	"math/big"
	"os"
	"strings"
)

func showMenu() {
	fmt.Println(strings.Repeat("═", 80))
	fmt.Println("           EXPLORADOR DE BLOCKCHAIN - MAGICARDS ")
	fmt.Println(strings.Repeat("═", 80))
	fmt.Println()
	fmt.Println("   RELATÓRIOS DO SISTEMA DE PACOTES:")
	fmt.Println("    [1] Resumo do Sistema")
	fmt.Println("    [2] Listar Todos os Pacotes")
	fmt.Println("    [3] Pacotes Recentes")
	fmt.Println()
	fmt.Println("   CONSULTAS DETALHADAS:")
	fmt.Println("    [4] Detalhes de um Pacote")
	fmt.Println("    [5] Histórico de um Jogador")
	fmt.Println("    [6] Histórico de uma Carta")
	fmt.Println("    [7] Detalhes de uma Transação")
	fmt.Println()
	fmt.Println("   RELATÓRIOS DAS PARTIDAS:")
	fmt.Println("    [8] Estatísticas de Partidas")
	fmt.Println("    [9] Estatísticas de um Jogador")
	fmt.Println()
	fmt.Println("    [0] Sair")
	fmt.Println()
	
	fmt.Println()
	fmt.Println(strings.Repeat("═", 80))
}

// ===== OPÇÃO 1: RESUMO DO SISTEMA =====
func showSystemReport() {
	fmt.Println(" RESUMO DO SISTEMA")
	fmt.Println(strings.Repeat("─", 80))

	summary, err := queryService.GetSystemReport(ctx)
	if err != nil {
		fmt.Printf(" Erro: %v\n", err)
		return
	}

	fmt.Printf("\n   Total de Pacotes:       %d\n", summary.TotalPackages)
	fmt.Printf("   Pacotes Abertos:        %d\n", summary.TotalPackagesOpened)
	fmt.Printf("   Pacotes Fechados:       %d\n", summary.TotalPackages-summary.TotalPackagesOpened)
	fmt.Printf("   Total de Cartas (NFTs): %d\n", summary.TotalCards)

	if summary.TotalPackages > 0 {
		percentOpen := float64(summary.TotalPackagesOpened) / float64(summary.TotalPackages) * 100
		fmt.Printf("\n   Taxa de Abertura:       %.1f%%\n", percentOpen)
	}

	fmt.Println()
}

// ===== OPÇÃO 2: LISTAR TODOS OS PACOTES =====
func listAllPackages() {
	fmt.Println(" LISTA DE PACOTES")
	fmt.Println(strings.Repeat("─", 80))

	summary, err := queryService.GetSystemReport(ctx)
	if err != nil {
		fmt.Printf(" Erro: %v\n", err)
		return
	}

	if summary.TotalPackages == 0 {
		fmt.Println("\n  ⚠️  Nenhum pacote encontrado.")
		return
	}

	fmt.Println()
	fmt.Printf("%-5s %-40s %-10s %-8s\n", "Nº", "Package ID", "Status", "Cartas")
	fmt.Println(strings.Repeat("─", 80))

	for i := 0; i < summary.TotalPackages; i++ {
		pkg, err := contracts.Package.GetPackageByIndex(
			client.CallOpts(),
			big.NewInt(int64(i)),
		)
		if err != nil {
			continue
		}

		status := " Fechado"
		if pkg.Opened {
			status = " Aberto"
		}

		// Trunca ID se muito longo
		displayID := pkg.Id
		if len(displayID) > 36 {
			displayID = displayID[:36]
		}

		fmt.Printf("%-5d %-40s %-10s %-8d\n",
			i+1,
			displayID,
			status,
			len(pkg.CardIds),
		)
	}
	fmt.Println()
}

// ===== OPÇÃO 3: DETALHES DE UM PACOTE =====
func showPackageDetails() {
	packageID := readInput("Digite o Package ID: ")

	fmt.Println("\n DETALHES DO PACOTE")
	fmt.Println(strings.Repeat("─", 80))

	pkg, err := queryService.GetPackageReport(ctx, packageID)
	if err != nil {
		fmt.Printf(" Erro: %v\n", err)
		return
	}

	status := " Fechado"
	if pkg.Opened {
		status = " Aberto"
	}

	fmt.Printf("\n   Package ID:  %s\n", pkg.PackageID)
	fmt.Printf("   Status:      %s\n", status)

	if pkg.Opened {
		fmt.Printf("   Aberto Por:  %s\n", pkg.OpenedBy)
	}

	fmt.Printf("   Criado Em:   %s\n", formatTimestamp(pkg.CreatedAt))
	fmt.Printf("\n   Cartas no Pacote (%d):\n", len(pkg.CardIDs))
	fmt.Println(strings.Repeat("  ─", 40))

	for i, cardID := range pkg.CardIDs {
		card, err := queryService.GetCardReport(ctx, cardID)
		if err != nil {
			fmt.Printf("    %d. %s (⚠️  não mintada)\n", i+1, cardID)
			continue
		}

		fmt.Printf("    %d. Token #%d - %s\n", i+1, card.TokenID, card.TemplateID)
		fmt.Printf("       Dono: %s\n", shortenAddress(card.CurrentOwner))
	}
	fmt.Println()
}

// ===== OPÇÃO 4: HISTÓRICO DO JOGADOR =====
func showPlayerReport() {
	playerAddress := readInput("Digite o endereço do jogador (0x...): ")
	playerID := readInput("Digite o Player ID (opcional, Enter para pular): ")

	if playerID == "" {
		playerID = "N/A"
	}

	fmt.Println("\n HISTÓRICO DO JOGADOR")
	fmt.Println(strings.Repeat("─", 80))

	player, err := queryService.GetPlayerReport(ctx, playerID, playerAddress)
	if err != nil {
		fmt.Printf(" Erro: %v\n", err)
		return
	}

	fmt.Printf("\n   Player ID:     %s\n", player.PlayerID)
	fmt.Printf("   Endereço:      %s\n", player.Address)
	fmt.Printf("   Saldo:         %.6f ETH\n", player.BalanceETH)
	fmt.Printf("   Total Cartas:  %d\n", player.TotalCards)

	if len(player.Cards) > 0 {
		fmt.Println("\n   CARTAS DO JOGADOR:")
		fmt.Println(strings.Repeat("  ─", 40))

		for _, card := range player.Cards {
			fmt.Printf("\n    Token #%d\n", card.TokenID)
			fmt.Printf("      Template:  %s\n", card.TemplateID)
			fmt.Printf("      Pacote:    %s\n", shortenID(card.PackageID))
			fmt.Printf("      Mintada:   %s\n", formatTimestamp(card.MintedAt))
		}
	} else {
		fmt.Println("\n  ⚠️  Jogador não possui cartas.")
	}
	fmt.Println()
}

// ===== OPÇÃO 5: HISTÓRICO DA CARTA =====
func showCardHistory() {
	cardID := readInput("Digite o Card ID: ")

	fmt.Println("\n HISTÓRICO DA CARTA")
	fmt.Println(strings.Repeat("─", 80))

	history, err := queryService.GetCardHistory(ctx, cardID)
	if err != nil {
		fmt.Printf(" Erro: %v\n", err)
		return
	}

	fmt.Printf("\n   Card ID:       %s\n", history.CardID)
	fmt.Printf("   Token ID:      #%d\n", history.TokenID)
	fmt.Printf("   Template:      %s\n", history.TemplateID)
	fmt.Printf("   Pacote Origem: %s\n", shortenID(history.PackageID))
	fmt.Printf("   Dono Atual:    %s\n", shortenAddress(history.CurrentOwner))
	fmt.Printf("   Mintada Em:    %s\n", formatTimestamp(history.MintedAt))

	if len(history.Transfers) > 0 {
		fmt.Println("\n   HISTÓRICO DE TRANSFERÊNCIAS:")
		fmt.Println(strings.Repeat("  ─", 40))

		for i, transfer := range history.Transfers {
			if i == 0 && transfer.From == "0x0000000000000000000000000000000000000000" {
				// Mint inicial
				fmt.Printf("\n     [MINT] Bloco #%d\n", transfer.BlockNumber)
				fmt.Printf("       ➜ Para: %s\n", shortenAddress(transfer.To))
			} else {
				// Transferência normal
				fmt.Printf("\n     [TRANSFERÊNCIA %d] Bloco #%d\n", i, transfer.BlockNumber)
				fmt.Printf("       ➜ De:   %s\n", shortenAddress(transfer.From))
				fmt.Printf("       ➜ Para: %s\n", shortenAddress(transfer.To))
			}
			fmt.Printf("        TX:   %s\n", shortenHash(transfer.TxHash))
			fmt.Printf("        Data: %s\n", formatTimestamp(transfer.Timestamp))
		}

		fmt.Printf("\n   Total de movimentações: %d\n", len(history.Transfers))
	} else {
		fmt.Println("\n    Nenhuma transferência registrada.")
	}
	fmt.Println()
}

// ===== OPÇÃO 6: DETALHES DA TRANSAÇÃO =====
func showTransactionDetails() {
	txHash := readInput("Digite o hash da transação (0x...): ")

	fmt.Println("\n DETALHES DA TRANSAÇÃO")
	fmt.Println(strings.Repeat("─", 80))

	tx, err := queryService.GetTransactionDetails(ctx, txHash)
	if err != nil {
		fmt.Printf(" Erro: %v\n", err)
		return
	}

	fmt.Printf("\n TX Hash:     %s\n", tx.TxHash)
	fmt.Printf("   Bloco:       #%d\n", tx.BlockNumber)
	fmt.Printf("   De:          %s\n", tx.From)
	fmt.Printf("   Para:        %s\n", tx.To)
	fmt.Printf("   Gas Usado:   %d\n", tx.GasUsed)
	fmt.Printf("   Status:      %s\n", tx.Status)
	fmt.Println()
}

// ===== OPÇÃO 7: BUSCAR POR ENDEREÇO =====
func searchByAddress() {
	address := readInput("Digite o endereço (0x...): ")

	fmt.Println("\n BUSCANDO INFORMAÇÕES...")
	fmt.Println(strings.Repeat("─", 80))

	// Buscar cartas
	player, err := queryService.GetPlayerReport(ctx, "unknown", address)
	if err != nil {
		fmt.Printf(" Erro: %v\n", err)
		return
	}

	fmt.Printf("\n Endereço: %s\n", address)
	fmt.Printf("   Saldo:    %.6f ETH\n", player.BalanceETH)
	fmt.Printf("   Cartas:   %d\n", player.TotalCards)

	if player.TotalCards > 0 {
		fmt.Println("\n  Templates das cartas:")
		for i, card := range player.Cards {
			fmt.Printf("    %d. %s (Token #%d)\n", i+1, card.TemplateID, card.TokenID)
		}
	}
	fmt.Println()
}

// ===== OPÇÃO 8: ATIVIDADES RECENTES =====
func showRecentActivity() {
	fmt.Println(" ATIVIDADES RECENTES")
	fmt.Println(strings.Repeat("─", 80))

	summary, err := queryService.GetSystemReport(ctx)
	if err != nil {
		fmt.Printf(" Erro: %v\n", err)
		return
	}

	// Últimos 5 pacotes
	fmt.Println("\n   ÚLTIMOS PACOTES CRIADOS:")
	start := summary.TotalPackages - 5
	if start < 0 {
		start = 0
	}

	for i := summary.TotalPackages - 1; i >= start && i >= 0; i-- {
		pkg, err := contracts.Package.GetPackageByIndex(
			client.CallOpts(),
			big.NewInt(int64(i)),
		)
		if err != nil {
			continue
		}

		status := "Fechado"
		if pkg.Opened {
			status = "Aberto"
		}

		fmt.Printf("    • %s - %s\n", shortenID(pkg.Id), status)
	}

	fmt.Println()
}

// =================== PARTIDAS =======================

func showMatchStatistics() {
	fmt.Println(" ESTATÍSTICAS DE PARTIDAS")
	fmt.Println(strings.Repeat("─", 80))

	stats, err := matchService.GetSystemStats(ctx)
	if err != nil {
		fmt.Printf(" Erro: %v\n", err)
		return
	}

	fmt.Printf("\n   RESUMO GERAL:\n")
	fmt.Printf("     Total de Partidas:  %d\n", stats["total_matches"])
	fmt.Printf("     Partidas Locais:    %d\n", stats["local_matches"])
	fmt.Printf("     Partidas Remotas:   %d\n", stats["remote_matches"])
	fmt.Printf("     Em Andamento:       %d\n", stats["active"])
	fmt.Printf("     Finalizadas:        %d\n", stats["finished"])

	if stats["total_matches"] > 0 {
		localPercent := float64(stats["local_matches"]) / float64(stats["total_matches"]) * 100
		remotePercent := float64(stats["remote_matches"]) / float64(stats["total_matches"]) * 100

		fmt.Printf("\n   DISTRIBUIÇÃO:\n")
		fmt.Printf("     Local:  %.1f%%\n", localPercent)
		fmt.Printf("     Remota: %.1f%%\n", remotePercent)
	}

	fmt.Println()
}

func showPlayerMatchStats() {
	playerID := readInput("Digite o Player ID: ")

	fmt.Println("\n ESTATÍSTICAS DO JOGADOR")
	fmt.Println(strings.Repeat("─", 80))

	
	stats, err := matchService.GetPlayerStats(ctx, playerID)
	if err != nil {
		fmt.Printf(" Erro: %v\n", err)
		return
	}

	fmt.Printf("\n  Player ID: %s\n", playerID)
	fmt.Printf("\n  ESTATÍSTICAS:\n")
	fmt.Printf("     Total de Partidas: %d\n", stats.TotalMatches)
	fmt.Printf("     Vitórias:          %d\n", stats.Wins)
	fmt.Printf("     Derrotas:          %d\n", stats.Losses)
	fmt.Printf("     Taxa de Vitória:   %d%%\n", stats.WinRate)

	if stats.TotalMatches > 0 {
		fmt.Printf("\n   PERFORMANCE:\n")

		var performance string
		switch {
		case stats.WinRate >= 70:
			performance = " Excelente"
		case stats.WinRate >= 50:
			performance = " Bom"
		case stats.WinRate >= 30:
			performance = "  Regular"
		default:
			performance = " Precisa Melhorar"
		}

		fmt.Printf("     Avaliação: %s\n", performance)
	}

	fmt.Println()
}

// ===== HELPERS =====

func readInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func waitForEnter() {
	fmt.Println("\n⏎  Pressione Enter para continuar...")
	bufio.NewReader(os.Stdin).ReadString('\n')
	clearScreen()
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func formatTimestamp(ts uint64) string {
	if ts == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d", ts)
}

func shortenAddress(addr string) string {
	if len(addr) <= 10 {
		return addr
	}
	return addr[:6] + "..." + addr[len(addr)-4:]
}

func shortenID(id string) string {
	if len(id) <= 20 {
		return id
	}
	return id[:8] + "..." + id[len(id)-8:]
}

func shortenHash(hash string) string {
	if len(hash) <= 14 {
		return hash
	}
	return hash[:10] + "..." + hash[len(hash)-4:]
}
