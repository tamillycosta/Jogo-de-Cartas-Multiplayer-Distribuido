package authhandler

import ("Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service"
	"github.com/gin-gonic/gin"

)

type Authhandler struct{
	gameServer *service.GameServer
}


func New(gameServer *service.GameServer)  *Authhandler{
	return &Authhandler{
		gameServer: gameServer,
	}
}


// GET /api/v1/user-exists
//Retorna 
func (*Authhandler) UserExists(ctx *gin.Context){

}