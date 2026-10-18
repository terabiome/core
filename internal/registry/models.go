package registry

import "github.com/terabiome/terabiome/internal/operations"

type OperationRegistry map[string]func() operations.Operation
