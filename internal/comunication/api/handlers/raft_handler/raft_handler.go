package rafthandler

import (
	raftService "Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/service/raft"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
)

// RaftHandler gerencia endpoints do Raft
type RaftHandler struct {
	raft *raftService.RaftService
}

func New(raft *raftService.RaftService) *RaftHandler {
	return &RaftHandler{
		raft: raft,
	}
}

//
// ------------------- Endpoints de Gerenciamento -------------------
// 

// GET /api/v1/raft/status
func (h *RaftHandler) GetStatus(ctx *gin.Context) {
	stats := h.raft.GetStats()

	ctx.JSON(http.StatusOK, gin.H{
		"is_leader":   h.raft.IsLeader(),
		"leader_id":   h.raft.GetLeaderID(),
		"leader_addr": h.raft.GetLeaderHTTPAddr(),
		"stats":       stats,
		"servers":     h.raft.GetServers(),
	})
}

// POST /api/v1/raft/join
func (h *RaftHandler) Join(ctx *gin.Context) {
	if !h.raft.IsLeader() {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":       "not leader",
			"leader_addr": h.raft.GetLeaderHTTPAddr(),
		})
		return
	}

	var req struct {
		ServerID string `json:"server_id" binding:"required"`
		HTTPAddr string `json:"http_addr" binding:"required"` // HTTP address!
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//  Garante que o endereço tem http://
	httpAddr := req.HTTPAddr
	if !strings.HasPrefix(httpAddr, "http://") && !strings.HasPrefix(httpAddr, "https://") {
		httpAddr = "http://" + httpAddr
	}

	log.Printf("[Raft] Adicionando servidor %s com endereço: %s", req.ServerID, httpAddr)

	if err := h.raft.AddVoter(req.ServerID, httpAddr); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "server added to cluster",
		"server_id": req.ServerID,
		"http_addr": httpAddr,
	})
}

// DELETE /api/v1/raft/remove/:server_id
func (h *RaftHandler) Remove(ctx *gin.Context) {
	if !h.raft.IsLeader() {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":       "not leader",
			"leader_addr": h.raft.GetLeaderHTTPAddr(),
		})
		return
	}

	serverID := ctx.Param("server_id")
	if err := h.raft.RemoveServer(serverID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "server removed",
		"server_id": serverID,
	})
}

// 
// ------------------- Endpoints de RPCs Internos do Raft -------------------
// (Chamados por outros servidores)
// 

// POST /api/v1/raft/append-entries
// Recebe AppendEntries de outro servidor
func (h *RaftHandler) AppendEntries(ctx *gin.Context) {
	var req raft.AppendEntriesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transport := h.raft.GetTransport()
	resp, err := transport.HandleAppendEntries(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// POST /api/v1/raft/request-vote
// Recebe RequestVote de outro servidor
func (h *RaftHandler) RequestVote(ctx *gin.Context) {
	var req raft.RequestVoteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transport := h.raft.GetTransport()
	resp, err := transport.HandleRequestVote(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}


// POST /api/v1/raft/install-snapshot
// Recebe InstallSnapshot de outro servidor
func (h *RaftHandler) InstallSnapshot(ctx *gin.Context) {
	var req struct {
		*raft.InstallSnapshotRequest
		Data []byte `json:"data"`
	}
	
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transport := h.raft.GetTransport()
	resp, err := transport.HandleInstallSnapshot(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}