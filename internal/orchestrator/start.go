package orchestrator

import (
	"context"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
)

func (o *orchestrator) Start() {
	ctx := context.Background()

	// Clone repo and start poller
	o.gitService.Clone(ctx)
	o.RunIndexer()
	go o.StartPoller(ctx)

	// Start Git Watcher
	gitEventsChan := o.gitService.GetPullEvents()
	go func() {
		for {
			select {
			case changedFiles := <-gitEventsChan:

				if len(changedFiles) == 0 {
					continue
				}
				// Make a copy since we're processing asynchronously
				// and the slice might be reused by git service
				files := make([]string, len(changedFiles))
				copy(files, changedFiles)
				o.RunIndexer()
				// [TODO] Now run pipelines affected by these changed files
				o.logger.LogInfo("Changed files: %v", files)
				pipelines := o.Pipelines.Load()
				triggeredPipelines := make(PipelineMap)
				for _, p := range *pipelines {
					i := -1
					if len(p.Spec.Triggers) == 0 {
						continue
					} else {
						if p.Spec.Triggers[0].Type == string(pipeline.GitChangesTriggerKind) {
							i = 0
						} else if len(p.Spec.Triggers) > 1 && p.Spec.Triggers[1].Type == string(pipeline.GitChangesTriggerKind) {
							i = 1
						}
					}
					if i != -1 {
						for _, f := range files {
							triggerd := false
							for _, path := range p.Spec.Triggers[i].Paths {

								match, err := doublestar.Match(path, f)
								if err != nil {
									o.logger.LogNewError("Error matching path pattern in pipeline %s: %v", p.Metadata.Name, err)
									continue
								}
								if match {
									triggeredPipelines[p.Metadata.Name] = p
									o.logger.LogInfo("Pipeline %s queued to run by git trigger via change in file %s matching path pattern %s", p.Metadata.Name, f, path)
									triggerd = true
									break
								}
							}
							if triggerd {
								break
							}
						}
					}
				}
				// Run matched pipelines
				for _, p := range triggeredPipelines {
					if _, err := o.RunPipeline(ctx, p.Metadata.Name, pipeline.GitChangesTriggerKind, nil); err != nil {
						o.logger.LogNewError("Failed to start pipeline %s: %v", p.Metadata.Name, err)
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// StartPoller polls the git repository at regular intervals for new remote changes
func (o *orchestrator) StartPoller(ctx context.Context) {

	o.logger.LogInfo("Starting git poller for repository: %s", o.gitService.GetRepoURL())
	ticker := time.NewTicker(time.Duration(o.gitService.GetPollingInterval()) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			o.mu.Lock()
			if err := o.gitService.Pull(ctx); err != nil {
				o.logger.LogNewError("Git pull failed: ", err)
			}
			o.mu.Unlock()
		case <-ctx.Done():
			o.logger.LogInfo("Git poller stopped.")
			return
		}
	}

}
