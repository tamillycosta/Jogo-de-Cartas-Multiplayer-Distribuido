package packageService

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/aplication/usecases"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/repository"
	raftService "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/comands"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/session"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
)
// Seriviço para gerenciar pacotes 
type PackageService struct {
	packageRepo    *repository.PackageRepository
	cardRepo       *repository.CardRepository
	raft           *raftService.RaftService
	sessionManager *session.SessionManager
}

func New(
	packageRepo *repository.PackageRepository,cardRepo *repository.CardRepository,raft *raftService.RaftService,sessionManager *session.SessionManager) *PackageService {
	return &PackageService{
		packageRepo:    packageRepo,
		cardRepo:       cardRepo,
		raft:           raft,
		sessionManager: sessionManager,
	}
}

// OpenPackage abre um pacote para um jogador
// Verifica se a sessão do cliente è valida 
// Seleciona disponivel se tiver 
// Caso encontre pacote , muda o estado para bloqueado 
// Ao abrir pacote muda o estado para abeeto e direciona cartas para jogador 
func (ps *PackageService) OpenPackage(playerID string) error {
	log.Printf("📦 [PackageService] Tentando abrir pacote para player: %s", playerID)

	// 1. Verifica se é líder
	if !ps.raft.IsLeader() {
		return fmt.Errorf("apenas o líder pode abrir packages")
	}

	if !ps.sessionManager.IsPlayerLoggedIn(playerID) {
		return errors.New("usuário não está logado")
	}

	// Busca pacotes disponíveis
	packages, err := ps.packageRepo.GetAll()
	if err != nil {
		return fmt.Errorf("não foi possível carregar os pacotes: %v", err)
	}

	// Seleciona pacote disponível
	availablePackage, err := usecases.SelectAvailablePackage(packages)
	if err != nil {
		return fmt.Errorf("erro ao selecionar pacote: %v", err)
	}

	
	if availablePackage == nil {
		return errors.New("nenhum pacote disponível no momento")
	}

	log.Printf("📦 [PackageService] Pacote selecionado: %s", availablePackage.ID)



	// Processo para abertura 
	err = ps.blockPackage(availablePackage.ID, playerID)
	if err != nil {
		return fmt.Errorf("erro ao bloquear pacote: %v", err)
	}

	err = ps.openPackage(availablePackage.ID, playerID)
	if err != nil {
		return fmt.Errorf("erro ao abrir pacote: %v", err)
	}
	err = ps.transferCards(availablePackage.ID, playerID)
	if err != nil {
		return fmt.Errorf("erro ao transferir cartas: %v", err)
	}

	log.Printf("[PackageService] Package %s aberto por jogador %s", availablePackage.ID, playerID)
	return nil
}


// --------------------  Auxiliares  -----------------------


func (ps *PackageService) blockPackage(packageID, playerID string) error {
	log.Printf("[PackageService] Bloqueando pacote %s para player %s", packageID, playerID)

	lockCmd := comands.LockPackageCommand{
		PackageID: packageID,
		PlayerID:  playerID,
	}

	lockData, _ := json.Marshal(lockCmd)
	response, err := ps.raft.ApplyCommand(comands.Command{
		Type: comands.CommandLockPackage,
		Data: lockData,
	})

	if err != nil {
		return fmt.Errorf("raft apply error: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("lock failed: %s", response.Error)
	}

	return nil
}


func (ps *PackageService) openPackage(packageID, playerID string) error {
	log.Printf("PackageService] Abrindo pacote %s", packageID)

	openCmd := comands.OpenPackageCommand{
		PackageID: packageID,
		PlayerID:  playerID,
	}

	openData, _ := json.Marshal(openCmd)
	response, err := ps.raft.ApplyCommand(comands.Command{
		Type: comands.CommandOpenPackage,
		Data: openData,
	})

	if err != nil {
		return fmt.Errorf("raft apply error: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("open failed: %s", response.Error)
	}

	return nil
}


func (ps *PackageService) transferCards(packageID, playerID string) error {
	log.Printf(" [PackageService] Transferindo cartas do pacote %s para player %s", packageID, playerID)

	
	packageData, err := ps.packageRepo.FindByIdWithCards(packageID)  
    if err != nil {
        return fmt.Errorf("erro ao buscar pacote: %v", err)
    }

	if packageData == nil {
		return errors.New("pacote não encontrado")
	}

	// Transfere cada carta
	for _, card := range packageData.Cards {
		transferCmd := comands.TransferCardCommand{
			CardID:   card.ID,
			PlayerID: playerID,
		}

		transferData, _ := json.Marshal(transferCmd)
		response, err := ps.raft.ApplyCommand(comands.Command{
			Type: comands.CommandTransferCard,
			Data: transferData,
		})

		if err != nil {
			log.Printf("[PackageService] Erro ao transferir carta %s: %v", card.ID, err)
			continue
		}

		if !response.Success {
			log.Printf("[PackageService] Falha ao transferir carta %s: %s", card.ID, response.Error)
			continue
		}

		log.Printf("[PackageService] Carta %s transferida para player %s", card.ID, playerID)
	}

	return nil
}


// ----------------- Serviço para criação dos pacotes --------------------------


//  cria um pacote com 5 cartas aleatórias
func (ps *PackageService) CreatePackage() error {
	if !ps.raft.IsLeader() {
		return fmt.Errorf("apenas o líder pode criar packages")
	}

	packageID := uuid.New().String()

	// Gera 5 cartas aleatórias
	cardTemplates := usecases.GenerateRandomCards(5)
	cardIDs := make([]string, 5)

	log.Printf("[PackageService] Criando package %s com 5 cartas", packageID)

	// Cria pacote
	pkgCmd := comands.CreatePackageCommand{
		PackageID: packageID,
		CardIDs:   cardIDs,
	}

	pkgData, _ := json.Marshal(pkgCmd)
	response, err := ps.raft.ApplyCommand(comands.Command{
		Type: comands.CommandCreatePackage,
		Data: pkgData,
	})

	if err != nil || !response.Success {
		return fmt.Errorf("erro ao criar package: %v", err)
	}

	// Cria cartas 
	for i, templateID := range cardTemplates {
		cardID := uuid.New().String()
		cardIDs[i] = cardID

		cardCmd := comands.CreateCardCommand{
			CardID:     cardID,
			TemplateID: templateID,
			PackageID:  packageID,
		}

		cardData, _ := json.Marshal(cardCmd)
		_, err := ps.raft.ApplyCommand(comands.Command{
			Type: comands.CommandCreateCard,
			Data: cardData,
		})

		if err != nil {
			log.Printf("⚠️ Erro ao criar carta %s: %v", cardID, err)
		}
	}

	log.Printf("[PackageService] Package %s criado com sucesso!", packageID)
	return nil
}


//  retorna pacotes disponíveis 
func (ps *PackageService) GetAvailablePackages() ([]*entities.Package, error) {
	allPackages, err := ps.packageRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var available []*entities.Package
	for _, pkg := range allPackages {
		if pkg.Status == "avalible" {
			available = append(available, pkg)
		}
	}

	return available, nil
}	