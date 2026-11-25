package operations_test

import (
	"context"
	"testing"

	"github.com/terabiome/core/pkg/operations"
)

type testOperation struct{}

func (o *testOperation) Name() string { return "test.op" }

func (o *testOperation) SetParameters(parameters map[string]any) error { return nil }

func (o *testOperation) PreCheck(env *operations.OperationEnvironment) error { return nil }

func (o *testOperation) Run(ctx context.Context, env *operations.OperationEnvironment) error {
	return nil
}

func TestRegistryRegisterAndGet(t *testing.T) {
	t.Parallel()

	registry := operations.NewRegistry()
	if err := registry.Register("test.op", func() operations.Operation {
		return &testOperation{}
	}); err != nil {
		t.Fatalf("Register() = %v, want nil", err)
	}

	op, err := registry.Get("test.op")
	if err != nil {
		t.Fatalf("Get() = %v, want nil", err)
	}
	if op.Name() != "test.op" {
		t.Fatalf("Name() = %q, want test.op", op.Name())
	}
}

func TestRegistryRejectsDuplicate(t *testing.T) {
	t.Parallel()

	registry := operations.NewRegistry()
	factory := func() operations.Operation { return &testOperation{} }

	if err := registry.Register("test.op", factory); err != nil {
		t.Fatalf("Register() = %v, want nil", err)
	}
	if err := registry.Register("test.op", factory); err == nil {
		t.Fatal("Register() = nil, want a duplicate error")
	}
}

func TestRegistryUnknownOperation(t *testing.T) {
	t.Parallel()

	if _, err := operations.NewRegistry().Get("missing"); err == nil {
		t.Fatal("Get() = nil, want an unknown operation error")
	}
}
