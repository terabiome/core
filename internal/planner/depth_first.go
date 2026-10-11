package planner

import yamlcontracts "github.com/terabiome/terabiome/internal/contracts/yaml"

type depthFirstStrategy struct{}

func (s depthFirstStrategy) generateBatches(uow yamlcontracts.UnitOfWork) []yamlcontracts.UnitOfWorkSequenceBatch {
	graph := uow.GetSequenceDependencyGraph()

	primitiveBatches := s.groupByDepthUsingGraph(uow.Sequences, graph.FirstDegreeGraph)
	batches := make([]yamlcontracts.UnitOfWorkSequenceBatch, 0, len(primitiveBatches))
	for idx, primitiveBatch := range primitiveBatches {
		for subIdx := range primitiveBatch {
			primitiveBatch[subIdx].BatchID = subIdx
		}
		batches = append(batches, yamlcontracts.UnitOfWorkSequenceBatch{
			ID:        idx,
			Sequences: primitiveBatch,
		})
	}

	return batches
}

// groupByDepthUsingGraph uses the pre-validated adjacency list to find max depth.
func (s depthFirstStrategy) groupByDepthUsingGraph(sequences []yamlcontracts.UnitOfWorkSequence, graph map[string]map[string]struct{}) [][]yamlcontracts.UnitOfWorkSequence {
	if len(sequences) == 0 {
		return nil
	}

	nameToIdx := make(map[string]int)
	for i, seq := range sequences {
		nameToIdx[seq.Name] = i
	}

	depths := make([]int, len(sequences))
	for i, seq := range sequences {
		// Use the graph to find dependencies instead of raw DependsOn slice
		depths[i] = s.calculateDepthFromGraph(seq.Name, graph, nameToIdx, make(map[string]bool))
	}

	maxDepth := 0
	for _, d := range depths {
		if d > maxDepth {
			maxDepth = d
		}
	}

	batches := make([][]yamlcontracts.UnitOfWorkSequence, maxDepth+1)
	for i, seq := range sequences {
		batches[depths[i]] = append(batches[depths[i]], seq)
	}

	return batches
}

func (s depthFirstStrategy) calculateDepthFromGraph(name string, graph map[string]map[string]struct{}, nameToIdx map[string]int, visited map[string]bool) int {
	if visited[name] {
		return 0
	}
	visited[name] = true

	deps, exists := graph[name]
	if !exists || len(deps) == 0 {
		return 0
	}

	maxParentDepth := -1
	for depName := range deps {
		parentDepth := s.calculateDepthFromGraph(depName, graph, nameToIdx, visited)
		if parentDepth > maxParentDepth {
			maxParentDepth = parentDepth
		}
	}

	if maxParentDepth == -1 {
		return 0
	}
	return maxParentDepth + 1
}
