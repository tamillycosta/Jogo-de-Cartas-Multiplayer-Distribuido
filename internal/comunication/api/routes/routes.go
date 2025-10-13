package routes

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/comunication/api/handlers"
	
	"github.com/gin-gonic/gin"
)
	
	func SetupRoutes(router *gin.Engine, handler *handlers.Handler) {
		v1 := router.Group("/api/v1")
		{
			
			// ------------------ Informações do Servidor -----------------
			
			v1.GET("/info", handler.GetServerInfo)
			
			// ------------------ Autenticação (P2P via HTTP) ------------------
			
			auth := v1.Group("/auth")
			{
				auth.GET("/user-exists", handler.AuthHandler.UserExists)
			}
		
			
			// Raft ------------------- Gerenciamento do Cluster ------------------
			
			raft := v1.Group("/raft")
			{
				// Status e gerenciamento
				raft.GET("/status", handler.RaftHandler.GetStatus)
				raft.POST("/join", handler.RaftHandler.Join)
				raft.DELETE("/remove/:server_id", handler.RaftHandler.Remove)
				
				
				// -------------------- RPCs Internos do Raft -------------------
				// Chamados automaticamente por outros servidores
				
				raft.POST("/append-entries",handler.RaftHandler.AppendEntries)
				raft.POST("/request-vote", handler.RaftHandler.RequestVote)
				raft.POST("/install-snapshot", handler.RaftHandler.InstallSnapshot)
			}
		}
	}
	
