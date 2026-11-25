package workflow

import operationspkg "github.com/terabiome/core/pkg/operations"

type Registry interface {
	Get(name string) (operationspkg.Operation, error)
}
