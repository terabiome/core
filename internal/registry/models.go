package registry

import "github.com/terabiome/core/internal/operations"

type OperationRegistry map[string]func() operations.Operation
