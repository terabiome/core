package workflow

import (
	"context"
	"testing"
	"uuid"
)

func TestDispatchContextRoundTrip(t *testing.T) {
	t.Parallel()

	want := Dispatch{
		UUID:    uuid.MustParse("018f0000-0000-7000-8000-000000000001"),
		UOWUUID: uuid.MustParse("018f0000-0000-7000-8000-000000000002"),
	}

	got, ok := DispatchFromContext(WithDispatch(context.Background(), want))
	if !ok {
		t.Fatal("DispatchFromContext() ok = false, want true")
	}
	if got != want {
		t.Fatalf("DispatchFromContext() = %+v, want %+v", got, want)
	}
}

func TestExecutorWithDispatch(t *testing.T) {
	t.Parallel()

	want := Dispatch{
		UUID:    uuid.MustParse("018f0000-0000-7000-8000-000000000001"),
		UOWUUID: uuid.MustParse("018f0000-0000-7000-8000-000000000002"),
	}

	exec := NewWorkflowExecutor(nil, nil, 1, nil)
	got, ok := DispatchFromContext(exec.WithDispatch(context.Background(), want))
	if !ok {
		t.Fatal("DispatchFromContext() ok = false, want true")
	}
	if got != want {
		t.Fatalf("DispatchFromContext() = %+v, want %+v", got, want)
	}
}

func TestDispatchContextAbsent(t *testing.T) {
	t.Parallel()

	if _, ok := DispatchFromContext(context.Background()); ok {
		t.Fatal("DispatchFromContext() ok = true, want false")
	}
}

func TestWithDispatchIgnoresZero(t *testing.T) {
	t.Parallel()

	if _, ok := DispatchFromContext(WithDispatch(context.Background(), Dispatch{})); ok {
		t.Fatal("DispatchFromContext() ok = true, want false for zero dispatch")
	}
}
