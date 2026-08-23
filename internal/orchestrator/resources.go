package orchestrator

import (
	"sort"

	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
	"github.com/kunalvirwal/shogun-cd/internal/target"
)

func (o *orchestrator) ListPipelines() []*pipeline.Pipeline {
	pipelines := o.Pipelines.Load()
	if pipelines == nil {
		return []*pipeline.Pipeline{}
	}

	result := make([]*pipeline.Pipeline, 0, len(*pipelines))
	for _, resource := range *pipelines {
		result = append(result, resource)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Metadata.Name < result[j].Metadata.Name
	})

	return result
}

func (o *orchestrator) ListTargets() []*target.Target {
	targets := o.Targets.Load()
	if targets == nil {
		return []*target.Target{}
	}

	result := make([]*target.Target, 0, len(*targets))
	for _, resource := range *targets {
		result = append(result, resource)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Metadata.Name < result[j].Metadata.Name
	})

	return result
}
