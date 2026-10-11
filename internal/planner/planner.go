package planner

import (
	"errors"

	yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"
)

func GenerateSequenceBatch(uow yamlcontracts.UnitOfWork) ([][]yamlcontracts.UnitOfWorkSequence, error) {
	switch uow.Strategy {
	case yamlcontracts.UnitOfWorkStrategyDepthFirst:
		return depthFirstStrategy{}.generateBatches(uow), nil
	}
	return nil, errors.New("unsupported strategy")
}
