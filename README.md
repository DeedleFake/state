state
=====

state provides a very simple mechanism for creating, manipulating, and listening to observable state in Go. All state is based around interface

```go
type State[T any] interface {
	Listen(func(T)) CancelFunc
}
```

This registers a listener to be called when new values are available, and returns a function that deregisters the listener. The function is also called immediately with the current value of the state, and the `Listen()` method does not return until the passed function does.

Mutable state adds a `Set(T)` method that sets the current state and informs all of the listeners.

On top of these, a large number of utility functions are provided, most producing derived forms of state that modify the state in some way before passing it to listeners.
