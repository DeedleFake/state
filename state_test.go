package state_test

import (
	"testing"

	"deedles.dev/state"
	"gotest.tools/v3/assert"
)

func TestStatic(t *testing.T) {
	s := state.Static("some data")
	assert.Equal(t, "some data", state.Get(s))
}

func TestDerived(t *testing.T) {
	s := state.Derived(state.Static(3), func(v int) float64 { return float64(v) * 2.5 })
	assert.Equal(t, 7.5, state.Get(s))
}

func TestNilGet(t *testing.T) {
	assert.Equal(t, "", state.Get[string](nil))
}

type testManualGetState[T any] struct {
	s state.State[T]
}

func (s *testManualGetState[T]) Listen(f func(T)) state.CancelFunc {
	return s.s.Listen(f)
}

func TestManualGet(t *testing.T) {
	s := &testManualGetState[int]{s: state.Static(3)}
	assert.Equal(t, 3, state.Get[int](s))
}
