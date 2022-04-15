package state

import "sync"

// Listenable implements a registerable list of listening functions.
//
// Listenable's methods are thread-safe.
type Listenable[T any] struct {
	m   sync.RWMutex
	id  uint32
	lis map[uint32]func(T)
}

// Add registers a listener function, returning an ID that can be used
// to remove it later.
func (lis *Listenable[T]) Add(f func(T)) uint32 {
	lis.m.Lock()
	defer lis.m.Unlock()

	if lis.lis == nil {
		lis.lis = make(map[uint32]func(T))
	}

	id := lis.id
	lis.id++
	lis.lis[id] = f

	return id
}

// Remove deregisters the listener function with the given ID.
func (lis *Listenable[T]) Remove(id uint32) {
	lis.m.Lock()
	defer lis.m.Unlock()

	delete(lis.lis, id)
}

// Send calls all of the registered listener functions with the given
// value. It does not return until all of the registered functions do.
//
// TODO: Don't send to the listener that triggered the send?
func (lis *Listenable[T]) Send(v T) {
	lis.m.RLock()
	defer lis.m.RUnlock()

	for _, f := range lis.lis {
		f(v)
	}
}
