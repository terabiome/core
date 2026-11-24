package services

import (
	"context"
	"fmt"
)

type baseService struct {
	txManager TransactionManager
}

func newBaseService(txManager TransactionManager) baseService {
	return baseService{txManager: txManager}
}

type idempotencyContextKeyType string

const (
	IdempotencyContextKey idempotencyContextKeyType = "repositories.idempotency_context_key"
)

func loadIdempotencyKeyFromCtx(ctx context.Context) (string, error) {
	key := ctx.Value(IdempotencyContextKey)
	if key == nil {
		return "", nil
	}

	idempotencyKey, ok := key.(string)
	if !ok {
		return "", fmt.Errorf("type mismatch: %T (instead of [string])", key)
	}

	return idempotencyKey, nil
}

func setIdempotencyKeyInCtx(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, IdempotencyContextKey, value)
}
