package txmanager

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

type fakeTx struct {
	pgx.Tx
}

func TestTxFromContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tx, err := TxFromContext(ctx)
	if err != nil {
		t.Fatalf("TxFromContext() = %v", err)
	}
	if tx != nil {
		t.Fatalf("TxFromContext() = %v, want nil", tx)
	}

	ctx = context.WithValue(ctx, txContextKey{}, "not a tx")
	if _, err := TxFromContext(ctx); err == nil {
		t.Fatal("TxFromContext() = nil, want type mismatch error")
	}
}

func TestContextWithTxRejectsNil(t *testing.T) {
	t.Parallel()

	if _, err := ContextWithTx(context.Background(), nil); err == nil {
		t.Fatal("ContextWithTx(nil) = nil, want error")
	}
}

func TestWrapReusesExistingTx(t *testing.T) {
	t.Parallel()

	ctx, err := ContextWithTx(context.Background(), &fakeTx{})
	if err != nil {
		t.Fatalf("ContextWithTx() = %v", err)
	}

	var got context.Context
	err = NewTxManager(nil).Wrap(ctx, func(fnCtx context.Context) error {
		got = fnCtx
		return nil
	})
	if err != nil {
		t.Fatalf("Wrap() = %v", err)
	}
	if got != ctx {
		t.Fatal("Wrap() did not pass the original context")
	}
}

func TestWrapPropagatesNestedError(t *testing.T) {
	t.Parallel()

	ctx, err := ContextWithTx(context.Background(), &fakeTx{})
	if err != nil {
		t.Fatalf("ContextWithTx() = %v", err)
	}

	want := errors.New("boom")
	if err := NewTxManager(nil).Wrap(ctx, func(context.Context) error { return want }); !errors.Is(err, want) {
		t.Fatalf("Wrap() = %v, want %v", err, want)
	}
}
