package orchestrator

import (
	"context"
	"time"
)

func (o *orchestrator) Start() {
	ctx := context.Background()

	// Clone repo and start poller
	o.gitService.Clone(ctx)
	o.RunIndexer()
	go o.StartPoller(ctx)
	// vals := make(map[string]string)
	// vals["IMAGE"] = "my-web-app:v1.3.4"
	// vals["SERVICE_NAME"] = "web"
	// // [TEST] Run a pipeline
	// o.RunPipeline("deploy-my-app", pipeline.WebhookTriggerKind, vals)

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
