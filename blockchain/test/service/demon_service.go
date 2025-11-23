package service

import (
	"context"
	"fmt"

	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/blockchain/service"
	"strings"
	"time"
)
// Serviço ultilizado para fazer e apresentar os status da blockchain para o usário 
type BlockchainDemoService struct {
	queryService   *BlockchainQueryService
	packageService *service.PackageChainService
	cardService    *service.CardChainService
}

func NewBlockchainDemoService(
	queryService *BlockchainQueryService,
	packageService *service.PackageChainService,
	cardService *service.CardChainService,
) *BlockchainDemoService {
	return &BlockchainDemoService{
		queryService:   queryService,
		packageService: packageService,
		cardService:    cardService,
	}
}

// ===== DEMONSTRAÇÃO COMPLETA =====

func (ds *BlockchainDemoService) RunFullDemo(
	ctx context.Context,
	packageID string,
	playerID string,
	playerAddress string,
	txHash string,
) {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("               RELATÓRIO COMPLETO DA BLOCKCHAIN ")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// 1. Relatório do Sistema
	ds.printSystemReport(ctx)

	// 2. Relatório do Pacote 
	ds.printPackageReport(ctx, packageID)

	// 3. Relatório do Jogador
	ds.printPlayerReport(ctx, playerID, playerAddress)

	// 4. Relatório da Transação
	if txHash != "" {
		ds.printTransactionReport(ctx, txHash)
	}

	// 5. Verificação de Propriedade das Cartas
	ds.printOwnershipVerification(ctx, packageID, playerAddress)

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("               FIM DO RELATÓRIO")
	fmt.Println(strings.Repeat("=", 80))
}

func (ds *BlockchainDemoService) printSystemReport(ctx context.Context) {
	fmt.Println("┌" + strings.Repeat("─", 78) + "┐")
	fmt.Println("│" + centerText(" RESUMO DO SISTEMA", 78) + "│")
	fmt.Println("├" + strings.Repeat("─", 78) + "┤")

	summary, err := ds.queryService.GetSystemReport(ctx)
	if err != nil {
		fmt.Printf("│ ⚠️  Erro ao buscar resumo: %v%s│\n", err, strings.Repeat(" ", 50))
	} else {
		fmt.Printf("│  Total de Pacotes:        %d%s│\n", summary.TotalPackages, padRight("", 50))
		fmt.Printf("│  Pacotes Abertos:         %d%s│\n", summary.TotalPackagesOpened, padRight("", 50))
		fmt.Printf("│  Total de Cartas (NFTs):  %d%s│\n", summary.TotalCards, padRight("", 50))
	}

	fmt.Println("└" + strings.Repeat("─", 78) + "┘")
	fmt.Println()
}




func (ds *BlockchainDemoService) printPackageReport(ctx context.Context, packageID string) {
	fmt.Println("┌" + strings.Repeat("─", 78) + "┐")
	fmt.Println("│" + centerText(" INFORMAÇÕES DO PACOTE", 78) + "│")
	fmt.Println("├" + strings.Repeat("─", 78) + "┤")

	pkg, err := ds.queryService.GetPackageReport(ctx, packageID)
	if err != nil {
		fmt.Printf("│ ⚠️  Erro ao buscar pacote: %v%s│\n", err, strings.Repeat(" ", 40))
	} else {
		status := " Fechado"
		if pkg.Opened {
			status = " Aberto"
		}

		fmt.Printf("│  Package ID:    %s │\n", padRight(pkg.PackageID, 60))
		fmt.Printf("│  Status:        %s%s│\n", status, padRight("", 52))
		fmt.Printf("│  Aberto Por:    %s │\n", padRight(pkg.OpenedBy, 60))
		fmt.Printf("│  Criado Em:     %s │\n", padRight(formatTimestamp(pkg.CreatedAt), 60))
		fmt.Println("│" + strings.Repeat(" ", 78) + "│")
		fmt.Println("│  Cartas no Pacote:" + strings.Repeat(" ", 58) + "│")

		for i, cardID := range pkg.CardIDs {
			fmt.Printf("│    %d. %s │\n", i+1, padRight(cardID, 70))
		}
	}

	fmt.Println("└" + strings.Repeat("─", 78) + "┘")
	fmt.Println()
}

func (ds *BlockchainDemoService) printPlayerReport(ctx context.Context, playerID string, playerAddress string) {
	fmt.Println("┌" + strings.Repeat("─", 78) + "┐")
	fmt.Println("│" + centerText(" INFORMAÇÕES DO JOGADOR", 78) + "│")
	fmt.Println("├" + strings.Repeat("─", 78) + "┤")

	player, err := ds.queryService.GetPlayerReport(ctx, playerID, playerAddress)
	if err != nil {
		fmt.Printf("│ ⚠️  Erro ao buscar jogador: %v%s│\n", err, strings.Repeat(" ", 40))
	} else {
		fmt.Printf("│  Player ID:      %s │\n", padRight(player.PlayerID, 58))
		fmt.Printf("│  Endereço:       %s │\n", padRight(player.Address, 58))
		fmt.Printf("│  Saldo:          %.6f ETH%s│\n", player.BalanceETH, padRight("", 46))
		fmt.Printf("│  Total de Cartas: %d%s│\n", player.TotalCards, padRight("", 56))
		fmt.Println("│" + strings.Repeat(" ", 78) + "│")

		if len(player.Cards) > 0 {
			fmt.Println("│   Cartas do Jogador (NFTs):" + strings.Repeat(" ", 47) + "│")
			fmt.Println("│" + strings.Repeat(" ", 78) + "│")

			for _, card := range player.Cards {
				fmt.Printf("│    ┌─ Token ID: %d%s│\n", card.TokenID, padRight("", 60))
				fmt.Printf("│    │  Card ID:     %s │\n", padRight(card.CardID, 58))
				fmt.Printf("│    │  Template:    %s │\n", padRight(card.TemplateID, 58))
				fmt.Printf("│    │  Pacote:      %s │\n", padRight(card.PackageID, 58))
				fmt.Printf("│    │  Dono Atual:  %s │\n", padRight(card.CurrentOwner, 58))
				fmt.Printf("│    └─ Mintada Em:  %s │\n", padRight(formatTimestamp(card.MintedAt), 58))
				fmt.Println("│" + strings.Repeat(" ", 78) + "│")
			}
		}
	}

	fmt.Println("└" + strings.Repeat("─", 78) + "┘")
	fmt.Println()
}

func (ds *BlockchainDemoService) printTransactionReport(ctx context.Context, txHash string) {
	fmt.Println("┌" + strings.Repeat("─", 78) + "┐")
	fmt.Println("│" + centerText(" DETALHES DA TRANSAÇÃO", 78) + "│")
	fmt.Println("├" + strings.Repeat("─", 78) + "┤")

	tx, err := ds.queryService.GetTransactionDetails(ctx, txHash)
	if err != nil {
		fmt.Printf("│ ⚠️  Erro ao buscar transação: %v%s│\n", err, strings.Repeat(" ", 35))
	} else {
		statusIcon := "❌"
		if tx.Status == "Sucesso" {
			statusIcon = "✅"
		}

		fmt.Printf("│  TX Hash:        %s │\n", padRight(tx.TxHash, 58))
		fmt.Printf("│  Bloco:          %d%s│\n", tx.BlockNumber, padRight("", 56))
		fmt.Printf("│  De (From):      %s │\n", padRight(tx.From, 58))
		fmt.Printf("│  Para (To):      %s │\n", padRight(tx.To, 58))
		fmt.Printf("│  Gas Usado:      %d%s│\n", tx.GasUsed, padRight("", 56))
		fmt.Printf("│  Status:         %s %s%s│\n", statusIcon, tx.Status, padRight("", 50))
	}

	fmt.Println("└" + strings.Repeat("─", 78) + "┘")
	fmt.Println()
}

func (ds *BlockchainDemoService) printOwnershipVerification(ctx context.Context, packageID string, playerAddress string) {
	fmt.Println("┌" + strings.Repeat("─", 78) + "┐")
	fmt.Println("│" + centerText(" VERIFICAÇÃO DE PROPRIEDADE", 78) + "│")
	fmt.Println("├" + strings.Repeat("─", 78) + "┤")

	pkg, err := ds.queryService.GetPackageReport(ctx, packageID)
	if err != nil {
		fmt.Printf("│ ⚠️  Erro ao buscar pacote: %v%s│\n", err, strings.Repeat(" ", 40))
		fmt.Println("└" + strings.Repeat("─", 78) + "┘")
		return
	}

	fmt.Println("│  Verificando se todas as cartas do pacote pertencem ao jogador..." + strings.Repeat(" ", 10) + "│")
	fmt.Println("│" + strings.Repeat(" ", 78) + "│")

	allMatch := true
	for _, cardID := range pkg.CardIDs {
		card, err := ds.queryService.GetCardReport(ctx, cardID)
		if err != nil {
			fmt.Printf("│     %s - Não encontrada%s│\n", padRight(cardID[:20]+"...", 30), padRight("", 25))
			allMatch = false
			continue
		}

		ownerMatch := strings.EqualFold(card.CurrentOwner, playerAddress)
		icon := "❌"
		status := "NÃO PERTENCE"
		if ownerMatch {
			icon = "✅"
			status = "PERTENCE"
		} else {
			allMatch = false
		}

		fmt.Printf("│    %s Token #%d - %s%s│\n", icon, card.TokenID, status, padRight("", 45))
	}

	fmt.Println("│" + strings.Repeat(" ", 78) + "│")

	if allMatch {
		fmt.Println("│   VERIFICAÇÃO COMPLETA: Todas as cartas pertencem ao jogador!" + strings.Repeat(" ", 13) + "│")
	} else {
		fmt.Println("│    ATENÇÃO: Algumas cartas não pertencem ao jogador esperado!" + strings.Repeat(" ", 13) + "│")
	}

	fmt.Println("└" + strings.Repeat("─", 78) + "┘")
	fmt.Println()
}

// ===== HELPERS =====

func centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}
	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text + strings.Repeat(" ", width-len(text)-padding)
}

func padRight(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}
	return text + strings.Repeat(" ", width-len(text))
}


func formatTimestamp(ts uint64) string {
	if ts == 0 {
		return "N/A"
	}
	t := time.Unix(int64(ts), 0)
	return t.Format("02/01/2006 15:04:05")
}
