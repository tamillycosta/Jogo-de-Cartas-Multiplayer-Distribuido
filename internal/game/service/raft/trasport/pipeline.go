package trasport

import (
	"fmt"
	"time"

	"github.com/hashicorp/raft"
)

// implementação para AppendEnties com pipiline
// esta funcionalidade da biblioteca da hashicorp/raft permite o envio assincrono de logs
// é uma forma mais efiente e é automtaticamente selecionada
//  pela lib quando a acumulo de logs a serem enviados
type httpPipeline struct {
	transport *HTTPTransport
	target    raft.ServerAddress
	doneCh    chan raft.AppendFuture
}


// adiciona uma entrada ao pipeline
// retorna esturtura apprendFuture
// o heartbeat tbm e mandado pelo appendEntries
func (p *httpPipeline) AppendEntries(args *raft.AppendEntriesRequest, resp *raft.AppendEntriesResponse) (raft.AppendFuture, error) {
	future := &appendFuture{
		start: time.Now(),
		args:  args,
		resp:  resp,
	}

	// Envia de forma assíncrona
	go func() {
		url := fmt.Sprintf("%s/api/v1/raft/append-entries", p.target)
		future.err = p.transport.sendRPC(url, args, resp)
		future.responded = time.Now()

		select {
		case p.doneCh <- future:
		default:

		}
	}()

	return future, nil
}

//  retorna canal de futures completados
func (p *httpPipeline) Consumer() <-chan raft.AppendFuture {
	return p.doneCh
}

// fecha o pipeline
func (p *httpPipeline) Close() error {
	close(p.doneCh)
	return nil
}

// implementa raft.AppendFuture
// esta esturura é usada para retornar informações sobre o AppendEntries com pipeline
type appendFuture struct {
	start     time.Time
	args      *raft.AppendEntriesRequest
	resp      *raft.AppendEntriesResponse
	err       error
	responded time.Time
}

// metodoes obirgatorios da esturutra
func (a *appendFuture) Error() error {
	return a.err
}

func (a *appendFuture) Start() time.Time {
	return a.start
}

func (a *appendFuture) Request() *raft.AppendEntriesRequest {
	return a.args
}

func (a *appendFuture) Response() *raft.AppendEntriesResponse {
	return a.resp
}

// Implementa IndexFuture para compatibilidade
type deferError struct {
	err error
}

func (d *deferError) Error() error {
	return d.err
}
