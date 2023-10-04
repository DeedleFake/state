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

	var latest float64
	defer s.Listen(func(v float64) { latest = v })()
	assert.Equal(t, 7.5, latest)
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

func TestMutable(t *testing.T) {
	s := state.Mutable(1)
	assert.Equal(t, 1, state.Get[int](s))
	s.Set(2)
	assert.Equal(t, 2, state.Get[int](s))

	var latest int
	defer s.Listen(func(v int) { latest = v })()
	assert.Equal(t, 2, latest)
	s.Set(3)
	assert.Equal(t, 3, latest)
}

func TestUniq(t *testing.T) {
	m := state.Mutable(0)
	s := state.Uniq[int](m)

	var vals []int
	defer s.Listen(func(v int) { vals = append(vals, v) })()
	assert.DeepEqual(t, []int{0}, vals)
	m.Set(1)
	m.Set(1)
	m.Set(2)
	m.Set(1)
	m.Set(3)
	m.Set(3)
	m.Set(3)
	assert.DeepEqual(t, []int{0, 1, 2, 1, 3}, vals)

	assert.Equal(t, 3, state.Get(s))
}

func TestIter(t *testing.T) {
	m := state.Mutable(0)
	go func() {
		for i := range 10 {
			m.Set(i)
		}
	}()

	for v := range state.Iter(m) {
		t.Log(v)
	}
}
