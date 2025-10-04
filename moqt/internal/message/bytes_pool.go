package message

import (
	"sync"
)

var pool = NewPool(64, 256, 1024)

func NewPool(min, middle, max int) *Pool {
	if min <= 0 || middle <= 0 || max <= 0 {
		panic("min, middle, max must be greater than 0")
	}

	if min < middle && middle < max {
		p := &Pool{
			min:    min,
			middle: middle,
			max:    max,
			pools:  make([]*sync.Pool, 3),
		}
		p.pools[0] = &sync.Pool{
			New: func() any {
				b := make([]byte, min)
				return &b
			},
		}
		p.pools[1] = &sync.Pool{
			New: func() any {
				b := make([]byte, middle)
				return &b
			},
		}
		p.pools[2] = &sync.Pool{
			New: func() any {
				b := make([]byte, max)
				return &b
			},
		}

		return p
	} else {
		panic("min, middle, max must be in ascending order")
	}
}

type Pool struct {
	mu     sync.RWMutex
	pools  []*sync.Pool
	min    int
	middle int
	max    int
}

func (p *Pool) Get(cap int) []byte {
	if cap <= p.min {
		return p.getMin()
	} else if cap <= p.middle {
		return p.getMiddle()
	} else if cap <= p.max {
		return p.getMax()
	} else {
		return make([]byte, 0, cap)
	}
}

func (p *Pool) getMin() []byte {
	p.mu.RLock()
	pool := p.pools[0]
	p.mu.RUnlock()

	b := pool.Get().(*[]byte)
	return (*b)[:0]
}

func (p *Pool) getMiddle() []byte {
	p.mu.RLock()
	pool := p.pools[1]
	defer p.mu.RUnlock()

	b := pool.Get().(*[]byte)
	return (*b)[:0]
}

func (p *Pool) getMax() []byte {
	p.mu.RLock()
	pool := p.pools[2]
	p.mu.RUnlock()

	b := pool.Get().(*[]byte)
	return (*b)[:0]
}

func (p *Pool) Put(b []byte) {
	l := len(b)
	switch l {
	case p.min:
		p.mu.Lock()
		defer p.mu.Unlock()
		p.pools[0].Put(&b)
		return
	case p.middle:
		p.mu.Lock()
		defer p.mu.Unlock()
		p.pools[1].Put(&b)
		return
	case p.max:
		p.mu.Lock()
		defer p.mu.Unlock()
		p.pools[2].Put(&b)
		return
	default:
		// If the length does not match any pool, GC the slice
		return
	}
}
