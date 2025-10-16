package comands

const (

CommandCreatePackage CommandType = "CREATE_PACKAGE"
CommandLockPackage   CommandType = "LOCK_PACKAGE"
CommandOpenPackage   CommandType = "OPEN_PACKAGE"


CommandCreateCard    CommandType = "CREATE_CARD"
CommandTransferCard  CommandType = "TRANSFER_CARD"
)





// Comandos de Packages

type CreatePackageCommand struct {
	PackageID string   `json:"package_id"`
	CardIDs   []string `json:"card_ids"` 
}

type LockPackageCommand struct {
	PackageID string `json:"package_id"`
	PlayerID  string `json:"player_id"`
}

type OpenPackageCommand struct {
	PackageID string `json:"package_id"`
	PlayerID  string `json:"player_id"`
}

// Comandos de Cards

type CreateCardCommand struct {
	CardID     string `json:"card_id"`
	TemplateID string `json:"template_id"` // Ex: "starter_mage"
	PackageID  string `json:"package_id"`  // Pacote que contém a carta
}

type TransferCardCommand struct {
	CardID   string `json:"card_id"`
	PlayerID string `json:"player_id"` // Transfere carta para jogador
}