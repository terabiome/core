package manifest

import (
	"testing"
	"time"
)

func TestCheckIfContainCycle(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		uow      UnitOfWork
		hasCycle bool
	}{
		{
			name: "independent sequences",
			uow: UnitOfWork{Sequences: []UnitOfWorkSequence{
				{Name: "os-hardening"},
				{Name: "other"},
			}},
		},
		{
			name: "chain",
			uow: UnitOfWork{Sequences: []UnitOfWorkSequence{
				{Name: "os-hardening"},
				{Name: "k3s-foundation", DependsOn: []string{"os-hardening"}},
			}},
		},
		{
			name: "diamond",
			uow: UnitOfWork{Sequences: []UnitOfWorkSequence{
				{Name: "root"},
				{Name: "left", DependsOn: []string{"root"}},
				{Name: "right", DependsOn: []string{"root"}},
				{Name: "join", DependsOn: []string{"left", "right"}},
			}},
		},
		{
			name: "mutual cycle",
			uow: UnitOfWork{Sequences: []UnitOfWorkSequence{
				{Name: "a", DependsOn: []string{"b"}},
				{Name: "b", DependsOn: []string{"a"}},
			}},
			hasCycle: true,
		},
		{
			name: "self cycle",
			uow: UnitOfWork{Sequences: []UnitOfWorkSequence{
				{Name: "a", DependsOn: []string{"a"}},
			}},
			hasCycle: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := checkCycleOrFail(t, tc.uow)
			if got != tc.hasCycle {
				t.Fatalf("CheckIfContainCycle() = %v, want %v", got, tc.hasCycle)
			}
		})
	}
}

// checkCycleOrFail fails if the detector does not return. A finished node used to be
// left in the graph and appended to the queue on every pass.
func checkCycleOrFail(t *testing.T, uow UnitOfWork) bool {
	t.Helper()

	done := make(chan bool, 1)
	go func() {
		done <- uow.CheckIfContainCycle()
	}()

	select {
	case hasCycle := <-done:
		return hasCycle
	case <-time.After(time.Second):
		t.Fatal("CheckIfContainCycle did not return")
		return false
	}
}
