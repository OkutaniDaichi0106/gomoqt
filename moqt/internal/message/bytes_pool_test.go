package message

import (
	"testing"
)

func TestNewPool(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p := NewPool(64, 256, 1024)

		if p.min != 64 || p.middle != 256 || p.max != 1024 {
			t.Errorf("Pool values incorrect: min=%d, middle=%d, max=%d", p.min, p.middle, p.max)
		}
	})

	t.Run("panic min <= 0", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("NewPool should panic with min <= 0")
			}
		}()
		NewPool(0, 256, 1024)
	})

	t.Run("panic middle <= 0", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("NewPool should panic with middle <= 0")
			}
		}()
		NewPool(64, 0, 1024)
	})

	t.Run("panic max <= 0", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("NewPool should panic with max <= 0")
			}
		}()
		NewPool(64, 256, 0)
	})

	t.Run("panic not ascending", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("NewPool should panic with not ascending order")
			}
		}()
		NewPool(256, 64, 1024)
	})
}

func TestPoolGet(t *testing.T) {
	p := NewPool(64, 256, 1024)

	t.Run("get min", func(t *testing.T) {
		b := p.Get(32)
		if cap(b) != 64 {
			t.Errorf("Get(32) should return slice with cap 64, got %d", cap(b))
		}
	})

	t.Run("get middle", func(t *testing.T) {
		b := p.Get(128)
		if cap(b) != 256 {
			t.Errorf("Get(128) should return slice with cap 256, got %d", cap(b))
		}
	})

	t.Run("get max", func(t *testing.T) {
		b := p.Get(512)
		if cap(b) != 1024 {
			t.Errorf("Get(512) should return slice with cap 1024, got %d", cap(b))
		}
	})

	t.Run("get large", func(t *testing.T) {
		b := p.Get(2048)
		if cap(b) != 2048 {
			t.Errorf("Get(2048) should return slice with cap 2048, got %d", cap(b))
		}
	})
}

func TestPoolPut(t *testing.T) {
	p := NewPool(64, 256, 1024)

	t.Run("put min", func(t *testing.T) {
		b := make([]byte, 64)
		p.Put(b)
		// No panic, and can get back
		got := p.Get(32)
		if cap(got) != 64 {
			t.Errorf("After Put, Get should return pooled slice")
		}
	})

	t.Run("put middle", func(t *testing.T) {
		b := make([]byte, 256)
		p.Put(b)
		got := p.Get(128)
		if cap(got) != 256 {
			t.Errorf("After Put, Get should return pooled slice")
		}
	})

	t.Run("put max", func(t *testing.T) {
		b := make([]byte, 1024)
		p.Put(b)
		got := p.Get(512)
		if cap(got) != 1024 {
			t.Errorf("After Put, Get should return pooled slice")
		}
	})

	t.Run("put wrong size", func(t *testing.T) {
		b := make([]byte, 128)
		p.Put(b) // Should not panic, just GC
	})
}
