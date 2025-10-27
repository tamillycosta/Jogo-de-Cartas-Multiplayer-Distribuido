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
				auth.GET("/is-player-logged-in", handler.AuthHandler.IsPlayerLoggedIn)
				auth.POST("/create-account", handler.AuthHandler.CreateAccount)
			}
		
			// ------------------- Gerecimaneto de pacotes ------------------------
			packages := v1.Group("/package")
			{
				packages.POST("/open-package", handler.PackageHandler.OpenPackage)
			}

			trade := v1.Group("/trade")
			{
				trade.POST("/execute", handler.TradeHandler.ExecuteTrade)
			}
			

			// ------------------- Gerecimaneto de partidas ------------------------
			matchs := v1.Group("/match")
			{
				matchs.POST("/global/join", handler.MatchHandler.HandleJoinGlobalQueue)
				matchs.POST("/created", handler.MatchHandler.HandleRemoteMatchNotification)
				matchs.POST("/sync", handler.MatchHandler.HandleMatchSync)
				matchs.POST("/action", handler.MatchHandler.HandleMatchAction)
				matchs.POST("/heartbeat", handler.MatchHandler.HandleHeartbeat)
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
	
