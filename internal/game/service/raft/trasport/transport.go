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
// a biblioteca padrão do hashicorp/raft é por tcp então é preciso impelemtar esta estrutura para
// comunicação via rest
type HTTPTransport struct {
	localAddr  raft.ServerAddress
	httpClient *http.Client
	consumer   chan raft.RPC
}

func New(bindAddr string, timeout time.Duration) *HTTPTransport {
	return &HTTPTransport{
		localAddr: raft.ServerAddress(bindAddr),
		httpClient: &http.Client{
			Timeout: timeout,
		},
		consumer: make(chan raft.RPC, 128),
	}
}

// retorna canal para receber RPCs
func (h *HTTPTransport) Consumer() <-chan raft.RPC {
	return h.consumer
}

// LocalAddr retorna endereço local 
func (h *HTTPTransport) LocalAddr() raft.ServerAddress {
	return h.localAddr
}

// envia logs para os servidores  de forma sincorna
// um log(requsição) apos o outro 
func (h *HTTPTransport) AppendEntries(
	id raft.ServerID,
	target raft.ServerAddress,
	args *raft.AppendEntriesRequest,
	resp *raft.AppendEntriesResponse,
) error {
	url := fmt.Sprintf("%s/api/v1/raft/append-entries", target)
	return h.sendRPC(url, args, resp)
}

// envia requisição de voto via HTTP POST
func (h *HTTPTransport) RequestVote(
	id raft.ServerID,
	target raft.ServerAddress,
	args *raft.RequestVoteRequest,
	resp *raft.RequestVoteResponse,
) error {
	url := fmt.Sprintf("%s/api/v1/raft/request-vote", target)
	return h.sendRPC(url, args, resp)
}

// envia snapshot via HTTP POST
// método usado quando o líder precisa enviar o snapshot atual para um seguidor desatualizado
func (h *HTTPTransport) InstallSnapshot(
	id raft.ServerID,
	target raft.ServerAddress,
	args *raft.InstallSnapshotRequest,
	resp *raft.InstallSnapshotResponse,
	data io.Reader,
) error {
	snapshotData, err := io.ReadAll(data)
	if err != nil {
		return err
	}

	payload := struct {
		*raft.InstallSnapshotRequest
		Data []byte `json:"data"`
	}{
		InstallSnapshotRequest: args,
		Data:                   snapshotData,
	}

	url := fmt.Sprintf("%s/api/v1/raft/install-snapshot", target)
	return h.sendRPC(url, payload, resp)
}


// define handler para heartbeats
func (h *HTTPTransport) SetHeartbeatHandler(cb func(rpc raft.RPC)) {
	// Heartbeats são tratados via AppendEntries então n precisa de uma rota especifica
}


// 
func (h *HTTPTransport) TimeoutNow(
	id raft.ServerID,
	target raft.ServerAddress,
	args *raft.TimeoutNowRequest,
	resp *raft.TimeoutNowResponse,
) error {
	url := fmt.Sprintf("%s/api/v1/raft/timeout-now", target)
	return h.sendRPC(url, args, resp)
}






// envia uma chamada RPC via HTTP 
// (rpcs são chamadas remota de função — um servidor chama função em outro )
func (h *HTTPTransport) sendRPC(url string, request, response interface{}) error {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := h.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http error %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}


//  cria um pipeline para replicação otimizada 
// Permite enviar múltiplas(de forma assincrona) requsiões de envio de logs sem aguardar resposta individual
func (h *HTTPTransport) AppendEntriesPipeline(id raft.ServerID,target raft.ServerAddress) (raft.AppendPipeline, error) {
	return &httpPipeline{
		transport: h,
		target:    target,
		doneCh:    make(chan raft.AppendFuture, 128),
	}, nil
}









// ------------------- Auxiliares -----------


// EncodePeer codifica endereço do peer
func (h *HTTPTransport) EncodePeer(id raft.ServerID, addr raft.ServerAddress) []byte {
	return []byte(addr)
}

// DecodePeer decodifica endereço do peer
func (h *HTTPTransport) DecodePeer(buf []byte) raft.ServerAddress {
	return raft.ServerAddress(buf)
}

func (h *HTTPTransport) Close() error {
	close(h.consumer)
	return nil
}

//-------------- Handlers para rotas http para transporte ----------
// talvez seja bom seprar em outro arquivo?

// processa AppendEntries recebido via HTTP
func (h *HTTPTransport) HandleAppendEntries(req *raft.AppendEntriesRequest) (*raft.AppendEntriesResponse, error) {
	respCh := make(chan raft.RPCResponse, 1)

	rpc := raft.RPC{
		Command:  req,
		RespChan: respCh,
	}

	// Envia para consumer (Raft processa)
	select {
	case h.consumer <- rpc:
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


//  processa RequestVote recebido via HTTP
func (h *HTTPTransport) HandleRequestVote(req *raft.RequestVoteRequest) (*raft.RequestVoteResponse, error) {
	respCh := make(chan raft.RPCResponse, 1)

	rpc := raft.RPC{
		Command:  req,
		RespChan: respCh,
	}

	select {
	case h.consumer <- rpc:
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
func (h *HTTPTransport) HandleInstallSnapshot(req struct {
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
	case h.consumer <- rpc:
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


