package registry

import (
	"github.com/terabiome/core/pkg/operations"
	"github.com/terabiome/core/pkg/operations/builtin"
)

func NewDefault() *operations.Registry {
	registry, err := builtin.NewRegistry()
	if err != nil {
		panic(err)
	}
	return registry
}
