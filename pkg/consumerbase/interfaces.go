package consumerbase

import "github.com/terabiome/core/pkg/operations"

type Registry interface {
	Get(name string) (operations.Operation, error)
}
