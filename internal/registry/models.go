package registry

import (
	operationspkg "github.com/terabiome/core/pkg/operations"
)

type OperationRegistry map[string]func() operationspkg.Operation
