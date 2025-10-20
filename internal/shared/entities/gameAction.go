package entities

// ----------------- TIPOS E ERROS -----------------

type GameAction struct {
	Type            string `json:"type"` // "play_card", "attack", "leave_match"
	CardID          string `json:"card_id,omitempty"`
	AttackerCardID  string `json:"attacker_card_id,omitempty"`
	TargetPlayerID  string `json:"target_player_id,omitempty"`
}

var (
	ErrNotYourTurn    = NewGameError("não é seu turno")
	ErrCardNotInHand  = NewGameError("carta não está na mão")
	ErrNotEnoughLife  = NewGameError("esta carta ja esta sem vida")
	ErrCardNotFound   = NewGameError("carta não encontrada")
	ErrOppoentCard = NewGameError("seu oponente ainda não tem carta em campo")
)

type GameError struct {
	Message string
}

func NewGameError(msg string) *GameError {
	return &GameError{Message: msg}
}

func (e *GameError) Error() string {
	return e.Message
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}