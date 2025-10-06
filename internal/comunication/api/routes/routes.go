package routes

import (
	

	"github.com/gin-gonic/gin"
)

// Rotas da aplicação 
func SetupRoutes(router *gin.Engine){
	v1 := router.Group("/api/v1")
	{
	  v1.GET("/info", )
	  v1.POST("/notify", )
	 
	}
	v2 := router.Group("/api/v2/notify")
	{
		v2.POST("/user-exits")
		v2.POST("/isLog")
	}
}