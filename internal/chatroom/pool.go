package chatroom

import "sync"

type Pool struct {
	data map[int]*ChatRoom
	mu   sync.RWMutex
}

func NewPool() *Pool {
	return &Pool{
		data: make(map[int]*ChatRoom),
	}
}

func (p *Pool) GetOrSet(id int) *ChatRoom {
	p.mu.RLock()
	cr, ok := p.data[id]
	p.mu.RUnlock()
	if !ok {
		p.mu.Lock()
		cr, ok = p.data[id]
		if !ok {
			cr = NewCR()
			p.data[id] = cr
		}
		p.mu.Unlock()
	}

	return cr
}
