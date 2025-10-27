package trasport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/raft"
)

// HTTPTransport implementa raft.Transport usando HTTP REST
type HTTPTransport struct {
	httpAddr   string
	httpClient *http.Client
	consumer   chan raft.RPC 
}

func New(httpAddr string, timeout time.Duration) *HTTPTransport {
	return &HTTPTransport{
		httpAddr: httpAddr,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		consumer: make(chan raft.RPC, 256), 
	}
}


func (t *HTTPTransport) AppendEntriesPipeline(id raft.ServerID, target raft.ServerAddress) (raft.AppendPipeline, error) {
	return &httpPipeline{
		transport:  t,
		target:     target,
		doneCh:     make(chan raft.AppendFuture, 128),
		shutdownCh: make(chan struct{}),
		closed:     false,
	}, nil
}


// ---------------------- implementa rotas http para ações  base da lib raft ----------------------

func (t *HTTPTransport) AppendEntries(id raft.ServerID, target raft.ServerAddress, args *raft.AppendEntriesRequest, resp *raft.AppendEntriesResponse) error {
	url := fmt.Sprintf("%s/api/v1/raft/append-entries", target)
	return t.sendRPC(url, args, resp)
}


func (t *HTTPTransport) RequestVote(id raft.ServerID, target raft.ServerAddress, args *raft.RequestVoteRequest, resp *raft.RequestVoteResponse) error {
	url := fmt.Sprintf("%s/api/v1/raft/request-vote", target)
	return t.sendRPC(url, args, resp)
}


func (t *HTTPTransport) InstallSnapshot(id raft.ServerID, target raft.ServerAddress, args *raft.InstallSnapshotRequest, resp *raft.InstallSnapshotResponse, data io.Reader) error {
	url := fmt.Sprintf("%s/api/v1/raft/install-snapshot", target)
	
	body := &bytes.Buffer{}
	body.Write([]byte(fmt.Sprintf("%v", args)))
	
	httpResp, err := t.httpClient.Post(url, "application/octet-stream", body)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("snapshot rejected: %d", httpResp.StatusCode)
	}

	return json.NewDecoder(httpResp.Body).Decode(resp)
}

func (t *HTTPTransport) TimeoutNow(id raft.ServerID, target raft.ServerAddress, args *raft.TimeoutNowRequest, resp *raft.TimeoutNowResponse) error {
	url := fmt.Sprintf("%s/api/v1/raft/timeout-now", target)
	return t.sendRPC(url, args, resp)
}


//----------------- impelmentados pela obrição da lib ----------------

func (t *HTTPTransport) EncodePeer(id raft.ServerID, addr raft.ServerAddress) []byte {
	return []byte(addr)
}

func (t *HTTPTransport) DecodePeer(buf []byte) raft.ServerAddress {
	return raft.ServerAddress(buf)
}


func (t *HTTPTransport) SetHeartbeatHandler(cb func(rpc raft.RPC)) {
	// Não implementado
}


//  envia uma chamada RPC via HTTP
func (t *HTTPTransport) sendRPC(url string, req interface{}, resp interface{}) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpResp, err := t.httpClient.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("request failed (%d): %s", httpResp.StatusCode, string(body))
	}

	if err := json.NewDecoder(httpResp.Body).Decode(resp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// ---------------------------- HANDLERS PARA ROTAS HTTP -----------------

// processa AppendEntries recebido via HTTP
func (t *HTTPTransport) HandleAppendEntries(req *raft.AppendEntriesRequest) (*raft.AppendEntriesResponse, error) {
	respCh := make(chan raft.RPCResponse, 1)

	rpc := raft.RPC{
		Command:  req,
		RespChan: respCh,
	}

	// Envia para consumer (Raft processa)
	select {
	case t.consumer <- rpc:
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout sending to consumer")
	}

	// Aguarda resposta
	select {
	case resp := <-respCh:
		if resp.Error != nil {
			return nil, resp.Error
		}
		return resp.Response.(*raft.AppendEntriesResponse), nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

// processa RequestVote recebido via HTTP
func (t *HTTPTransport) HandleRequestVote(req *raft.RequestVoteRequest) (*raft.RequestVoteResponse, error) {
	respCh := make(chan raft.RPCResponse, 1)

	rpc := raft.RPC{
		Command:  req,
		RespChan: respCh,
	}

	select {
	case t.consumer <- rpc:
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout sending to consumer")
	}

	select {
	case resp := <-respCh:
		if resp.Error != nil {
			return nil, resp.Error
		}
		return resp.Response.(*raft.RequestVoteResponse), nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

// processa InstallSnapshot recebido via HTTP
func (t *HTTPTransport) HandleInstallSnapshot(req struct {
	*raft.InstallSnapshotRequest
	Data []byte `json:"data"`
}) (*raft.InstallSnapshotResponse, error) {
	respCh := make(chan raft.RPCResponse, 1)

	// Cria reader a partir dos dados
	reader := bytes.NewReader(req.Data)

	rpc := raft.RPC{
		Command:  req.InstallSnapshotRequest,
		Reader:   reader,
		RespChan: respCh,
	}

	select {
	case t.consumer <- rpc:
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout sending to consumer")
	}

	select {
	case resp := <-respCh:
		if resp.Error != nil {
			return nil, resp.Error
		}
		return resp.Response.(*raft.InstallSnapshotResponse), nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response")
	}
}




// --------------------------- Auxiliares  ----------------------------

func (t *HTTPTransport) LocalAddr() raft.ServerAddress {
	return raft.ServerAddress(t.httpAddr)
}


func (t *HTTPTransport) Close() error {
	close(t.consumer)
	return nil
}

// Consumer retorna canal para processar RPCs recebidos
func (t *HTTPTransport) Consumer() <-chan raft.RPC {
	return t.consumer
}