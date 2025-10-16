package comands


//  representa dados para criar usuário
type CreateUserCommand struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// representa dados para deletar usuário
type DeleteUserCommand struct {
	UserID string `json:"user_id"`
}


type UpdateUserCommand struct{
	
}