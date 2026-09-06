package evaluate

import "testing"

func TestDecideDeterministic(t *testing.T) {
	first := Decide("myflag", "user-1", 50, true)
	for i := 0; i < 100; i++ {
		if got := Decide("myflag", "user-1", 50, true); got != first {
			t.Fatalf("non-deterministic result: first=%v, got=%v", first, got)
		}
	}
}

func TestDecideDisabledAlwaysFalse(t *testing.T) {
	for _, p := range []int{0, 10, 50, 100} {
		if Decide("f", "u", p, false) {
			t.Fatalf("enabled=false must always be false, got true for rollout %d", p)
		}
	}
}

func TestDecideRolloutZeroFalse(t *testing.T) {
	if Decide("f", "u", 0, true) {
		t.Fatal("rollout 0 must be false")
	}
}

func TestDecideRolloutHundredTrue(t *testing.T) {
	if !Decide("f", "u", 100, true) {
		t.Fatal("rollout 100 must be true")
	}
}
