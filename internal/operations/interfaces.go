package operations

import (
	"context"
)

// Operation is the atomic unit of work.
// Other packages/repositories implement this to enroll their own Operation (will be moving to pkg or public soon).
// PreCheck() should always run after SetParameters().
type Operation interface {
	Name() string
	SetParameters(parameters map[string]any) error
	PreCheck(env *OperationEnvironment) error
	Run(ctx context.Context, env *OperationEnvironment) error
}
