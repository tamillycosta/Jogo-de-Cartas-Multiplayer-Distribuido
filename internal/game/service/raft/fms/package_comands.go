package fms

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft/comands"
	"encoding/json"
	"log"
)

// ----------------- RAFT PARA CRIAÇÃO DE  PACOTES E CARTAS ----------------------


func (f *GameFSM)  applyCreatePackage(data json.RawMessage) *comands.ApplyResponse  {
	var cmd comands.CreatePackageCommand

	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	pkg := &entities.Package{
		ID: cmd.PackageID,
		Status: "avalible",

	}

	if _, err := f.packageRepository.CreateWithID(pkg); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	
	return &comands.ApplyResponse{Success: true, Data: pkg}
}



func (f *GameFSM) applyCreateCard(data json.RawMessage) *comands.ApplyResponse {
	var cmd comands.CreateCardCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	template, exists := entities.BaseCards[cmd.TemplateID]
	if !exists {
		return &comands.ApplyResponse{Success: false, Error: "template not found"}
	}

	card := &entities.Card{
		ID:        cmd.CardID,
		Name:      template.Nome,
		Type:      template.Type,
		Power:     template.Power,
		Rarity:    template.Rarity,
		Health: template.Health,
		PackageID: &cmd.PackageID,
	}

	if _, err := f.cardRepository.CreateWithID(card); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	log.Printf("[FSM] Card criado: %s (%s) no package %s", 
		cmd.CardID, cmd.TemplateID, cmd.PackageID)
	return &comands.ApplyResponse{Success: true, Data: card}
}

//------------------------ RAFT PARA ABERTURA DE PACOTES  ------------------------

// bloqueia um pacote para ele n ser selecionado por outro para abertura duas vezes 
func (f *GameFSM) applyLockPackage(data json.RawMessage)  *comands.ApplyResponse {
	var cmd comands.LockPackageCommand

	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	// se não for encontrado pacote com esse id 
	// retorna erro 
	pkg, err := f.packageRepository.FindById(cmd.PackageID)
	if err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}
	if pkg == nil {
		return &comands.ApplyResponse{Success: false, Error: "package not found"}
	}
	// se este pacote não estiver mais disponivel 
	// retorna erro 
	if pkg.Status != "avalible" {
		return &comands.ApplyResponse{Success: false, Error: "package not available"}
	}

	// Bloqueia package
	if err := f.packageRepository.UpdatePackageStatus(cmd.PackageID, "locked"); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	log.Printf("[FSM] Package bloqueado: %s para player %s", cmd.PackageID, cmd.PlayerID)
	return &comands.ApplyResponse{Success: true, Data: map[string]string{
		"package_id": cmd.PackageID,
		"status":     "locked",
	}}
}

//  modifica o statdo do pacote de bloaqueado para aberto 
func (f *GameFSM) applyOpenPackage(data json.RawMessage) *comands.ApplyResponse {
	var cmd comands.OpenPackageCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	// Marca package como aberto
	if err := f.packageRepository.UpdatePackageStatus(cmd.PackageID, "opened"); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	log.Printf("[FSM] Package aberto: %s por player %s", cmd.PackageID, cmd.PlayerID)
	return &comands.ApplyResponse{Success: true, Data: map[string]string{
		"package_id": cmd.PackageID,
		"status":     "opened",
	}}
}


// Relaciona uma carta com um player (finaliza processo de aberutura de pacote)
// desvincula a carta com o pacote que ele pertencia 
func (f *GameFSM) applyTransferCard(data json.RawMessage) *comands.ApplyResponse {
	var cmd comands.TransferCardCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return &comands.ApplyResponse{Success: false, Error: err.Error()}
	}

	// transferência de carta para jogador e remoção de id do pacote 
	f.cardRepository.UpdateCardStatus(cmd.PlayerID, cmd.CardID)

	log.Printf("[FSM] Card transferido: %s pa	ra player %s", cmd.CardID, cmd.PlayerID)
	return &comands.ApplyResponse{Success: true, Data: "card transferred"}
}


