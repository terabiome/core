package manifest_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/terabiome/core/pkg/contracts/manifest"
)

func TestPublicManifestAPI(t *testing.T) {
	t.Parallel()

	uow := manifest.UnitOfWork{
		Name:     "bootstrap",
		Strategy: manifest.UnitOfWorkStrategyDepthFirst,
		Type:     manifest.UnitOfWorkTypeWorkflow,
		Targets: []manifest.UnitOfWorkTarget{
			{Topics: []string{"dummy.dispatch"}},
		},
		Sequences: []manifest.UnitOfWorkSequence{
			{Name: "os-hardening", Steps: []manifest.UnitOfWorkStep{{Name: "shell.cmd"}}},
		},
	}

	if err := uow.Validate(validator.New()); err != nil {
		t.Fatalf("Validate() = %v, want nil", err)
	}
	if uow.CheckIfContainCycle() {
		t.Fatal("CheckIfContainCycle() = true, want false")
	}
}
