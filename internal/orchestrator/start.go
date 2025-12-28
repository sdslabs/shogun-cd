package orchestrator

import (
	"context"
)

func (o *orchestrator) Start() {
	ctx := context.Background()

	// Clone repo and start poller
	o.gitService.CloneAndStartPoller(ctx)
	o.RunIndexer()

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
	// Start Indexer
}
