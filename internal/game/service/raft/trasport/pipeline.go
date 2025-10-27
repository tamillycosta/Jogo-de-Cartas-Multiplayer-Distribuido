package trasport

import (
	"fmt"
	"time"
	"sync"
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
	
	mu       sync.RWMutex
	closed   bool
	shutdownCh chan struct{}
}




// adiciona uma entrada ao pipeline
// retorna esturtura apprendFuture
// o heartbeat tbm e mandado pelo appendEntries
func (p *httpPipeline) AppendEntries(args *raft.AppendEntriesRequest, resp *raft.AppendEntriesResponse) (raft.AppendFuture, error) {
	// Verifica se pipeline foi fechado
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return nil, fmt.Errorf("pipeline closed")
	}
	p.mu.RUnlock()

	future := &appendFuture{
		start: time.Now(),
		args:  args,
		resp:  resp,
	}

	
	go func() {
		url := fmt.Sprintf("%s/api/v1/raft/append-entries", p.target)
		future.err = p.transport.sendRPC(url, args, resp)
		future.responded = time.Now()

		//  Verifica se pipeline ainda está aberto antes de enviar
		p.mu.RLock()
		closed := p.closed
		p.mu.RUnlock()

		if closed {
			
			return
		}

		
		select {
		case p.doneCh <- future:
			
		case <-p.shutdownCh:
			
			return
		case <-time.After(100 * time.Millisecond):
			
			return
		}
	}()

	return future, nil
}
func (p *httpPipeline) Consumer() <-chan raft.AppendFuture {
	return p.doneCh
}


func (p *httpPipeline) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	
	// Fecha canal de shutdown primeiro (sinaliza para goroutines pararem)
	close(p.shutdownCh)
	
	time.Sleep(50 * time.Millisecond)
	
	// Fecha canal de resultados
	close(p.doneCh)
	
	return nil
}

// appendFuture implementa raft.AppendFuture
type appendFuture struct {
	start     time.Time
	args      *raft.AppendEntriesRequest
	resp      *raft.AppendEntriesResponse
	err       error
	responded time.Time
}

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


type deferError struct {
	err error
}

func (d *deferError) Error() error {
	return d.err
}