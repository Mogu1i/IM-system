package pool

import "sync"

type ConnectionPool struct {
	max    int
	active int
	mu     sync.RWMutex
}

func NewConnectionPool(max int) *ConnectionPool {
	if max <= 0 {
		max = 1
	}
	return &ConnectionPool{max: max}
}

func (p *ConnectionPool) Acquire() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.active >= p.max {
		return false
	}
	p.active++
	return true
}

func (p *ConnectionPool) Release() {
	p.mu.Lock()
	if p.active > 0 {
		p.active--
	}
	p.mu.Unlock()
}

func (p *ConnectionPool) Active() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.active
}
