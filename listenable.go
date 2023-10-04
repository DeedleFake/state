package state

import (
	"sync"
	"sync/atomic"
)

// Listenable implements a registerable list of listening functions.
//
// Listenable's methods are thread-safe.
type Listenable[T any] struct {
	id  atomic.Uint32
	lis sync.Map
}

// Add registers a listener function, returning an ID that can be used
// to remove it later.
func (lis *Listenable[T]) Add(f func(T)) uint32 {
	id := lis.id.Add(1) - 1
	lis.lis.Store(id, f)

	return id
}

// Remove deregisters the listener function with the given ID.
func (lis *Listenable[T]) Remove(id uint32) {
	lis.lis.Delete(id)
}

// Send calls all of the registered listener functions with the given
// value. It does not return until all of the registered functions do.
//
// TODO: Don't send to the listener that triggered the send?
func (lis *Listenable[T]) Send(v T) {
	lis.lis.Range(func(_, f any) bool {
		f.(func(T))(v)
		return true
	})
}
