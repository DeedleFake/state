package state

import "sync"

// State represents a value that can change over time.
type State[T any] interface {
	// Listen registers a listener function to be called whenever the
	// State's value changes.
	//
	// When Listen is first called, it immediately calls the provided
	// function with the current value of the state and does not return
	// until after the provided function returns.
	Listen(func(T)) CancelFunc
}

// getter is an optional interface defined by States to allow for
// quicker one-time fetching of their current value.
type getter[T any] interface {
	Get() T
}

// Setter represents a value that can be changed. Setting a value is
// thread-safe with reading it.
type Setter[T any] interface {
	Set(v T)
}

// MutableState is a State that can be changed. There is no guarantee
// about the relationship between the Set method returning and the
// State's listeners being informed of the update.
type MutableState[T any] interface {
	State[T]
	Setter[T]
}

// Get returns the current value of a given state. If the state passed
// is nil, it returns the zero-value of T.
//
// If the state passed has a Get method that returns a single value of
// type T, the return of that method is returned.
func Get[T any](s State[T]) (v T) {
	if s == nil {
		return v
	}

	if g, ok := s.(getter[T]); ok {
		return g.Get()
	}

	var once sync.Once
	cancel := s.Listen(func(c T) {
		once.Do(func() {
			v = c
		})
	})
	cancel()
	return v
}

type CancelFunc func()

type static[T any] struct {
	v T
}

// Static returns a state that never changes. If a value is definitely
// constant, using this type of state is more efficient than a mutable
// one.
func Static[T any](v T) State[T] {
	return static[T]{v: v}
}

func (s static[T]) Listen(listener func(T)) CancelFunc {
	listener(s.v)
	return func() {}
}

func (s static[T]) Get() T {
	return s.v
}

type mutable[T any] struct {
	lis Listenable[T]
	m   sync.RWMutex
	v   T
}

func Mutable[T any](v T) MutableState[T] {
	return &mutable[T]{v: v}
}

func (s *mutable[T]) Listen(f func(T)) CancelFunc {
	s.m.RLock()
	defer s.m.RUnlock()

	done := make(chan struct{})
	lf := func(v T) {
		select {
		case <-done:
		default:
			f(v)
		}
	}

	id := s.lis.Add(lf)
	lf(s.v)

	var cancel sync.Once
	return func() {
		cancel.Do(func() { close(done) })
		s.lis.Remove(id)
	}
}

func (s *mutable[T]) Set(v T) {
	s.m.Lock()
	defer s.m.Unlock()

	s.v = v
	s.lis.Send(v)
}

func (s *mutable[T]) Get() T {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.v
}

type derived[T, F any] struct {
	from State[F]
	m    func(F) T
}

// Derived returns a read-only state that derives its values from
// another state, passing them through the mapping function m. In
// other words, when from's value changes, the derived state's
// listeners will be called with that new value passed through m.
func Derived[T, F any, FS State[F]](from FS, m func(F) T) State[T] {
	return derived[T, F]{
		from: from,
		m:    m,
	}
}

func (s derived[T, F]) Listen(f func(T)) CancelFunc {
	return s.from.Listen(func(v F) {
		f(s.m(v))
	})
}

func (s derived[T, F]) Get() T {
	return s.m(Get(s.from))
}

type mutator[T, F any] struct {
	from MutableState[F]
	gm   func(F) T
	sm   func(T) F
}

// Mutator returns a mutable derived state. Like Derived, the returned
// state runs successive values through mapGet, but it can also be set
// and, when set, runs the new value through mapSet before setting the
// underlying state.
func Mutator[T, F any, FS MutableState[F]](from FS, mapGet func(F) T, mapSet func(T) F) MutableState[T] {
	return mutator[T, F]{
		from: from,
		gm:   mapGet,
		sm:   mapSet,
	}
}

func (s mutator[T, F]) Listen(f func(T)) CancelFunc {
	return s.from.Listen(func(v F) {
		f(s.gm(v))
	})
}

func (s mutator[T, F]) Set(v T) {
	s.from.Set(s.sm(v))
}

func (s mutator[T, F]) Get() T {
	return s.gm(Get(s.from))
}

type uniq[T any] struct {
	from  State[T]
	equal func(T, T) bool
}

// Uniq returns a State that wraps from. Listeners of the returned
// state will only see the first of successive values that are equal
// to each other, similar to the Unix uniq command.
func Uniq[T comparable, S State[T]](from S) State[T] {
	return uniq[T]{
		from: from,
		equal: func(v1, v2 T) bool {
			return v1 == v2
		},
	}
}

// UniqFunc is like Uniq, but uses a custom comparison function to
// determine equality.
func UniqFunc[T any, S State[T]](from S, equal func(T, T) bool) State[T] {
	return uniq[T]{
		from:  from,
		equal: equal,
	}
}

func (u uniq[T]) Listen(f func(T)) CancelFunc {
	var prev T
	var ok bool
	return u.from.Listen(func(v T) {
		if ok && u.equal(v, prev) {
			return
		}

		prev = v
		ok = true
		f(v)
	})
}

func (u uniq[T]) Get() T {
	return Get(u.from)
}

type Seq[T any] func(yield func(T) bool)

func Iter[T any](s State[T]) Seq[T] {
	return func(yield func(T) bool) {
		done := make(chan struct{})

		var cancel CancelFunc
		cancel = s.Listen(func(v T) {
			select {
			case <-done:
				return
			default:
				if !yield(v) {
					cancel()
					close(done)
				}
			}
		})

		<-done
	}
}
