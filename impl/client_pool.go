package impl

import (
	"fmt"
	"sync"
	"time"

	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
)

const (
	DefaultPoolSize   = 10
	MaxPoolSize       = 50
	ConnectionTimeout = 30 * time.Second
	IdleTimeout       = 5 * time.Minute
)

// RPCClientPool connection pool
type RPCClientPool struct {
	url         string
	pool        chan *rpcclient.RPCClient
	size        int
	maxSize     int
	mu          sync.Mutex
	created     int
	timeout     time.Duration
	idleTimeout time.Duration
}

// NewRPCClientPool creates a new connection pool
func NewRPCClientPool(url string, size int) *RPCClientPool {
	if size <= 0 {
		size = DefaultPoolSize
	}
	if size > MaxPoolSize {
		size = MaxPoolSize
	}

	return &RPCClientPool{
		url:         url,
		pool:        make(chan *rpcclient.RPCClient, size),
		size:        size,
		maxSize:     size,
		timeout:     ConnectionTimeout,
		idleTimeout: IdleTimeout,
	}
}

// getClient gets a connection from the pool
func (p *RPCClientPool) getClient() (*rpcclient.RPCClient, error) {
	select {
	case client := <-p.pool:
		// Get connection from pool
		return client, nil
	default:
		// No available connection in pool, create new one
		p.mu.Lock()
		if p.created < p.maxSize {
			p.created++
			p.mu.Unlock()

			client, err := rpcclient.NewRPCClient(p.url)
			if err != nil {
				p.mu.Lock()
				p.created--
				p.mu.Unlock()
				return nil, err
			}
			return client, nil
		}
		p.mu.Unlock()

		// Wait for available connection in pool
		select {
		case client := <-p.pool:
			return client, nil
		case <-time.After(p.timeout):
			return nil, fmt.Errorf("timeout waiting for available connection")
		}
	}
}

// putClient returns a connection to the pool
func (p *RPCClientPool) putClient(client *rpcclient.RPCClient) {
	if client == nil {
		return
	}

	select {
	case p.pool <- client:
		// Successfully returned to pool
	default:
		// Pool is full, close connection
		p.mu.Lock()
		p.created--
		p.mu.Unlock()
		client.Close()
	}
}

// Close closes the connection pool
func (p *RPCClientPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	close(p.pool)
	for client := range p.pool {
		if client != nil {
			client.Close()
		}
	}
}
