package orchestrator

import "context"

func (o *orchestrator) Start() {
	ctx := context.Background()

	// Clone repo and start poller
	o.gitService.CloneAndStartPoller(ctx)

	// Start Indexer

	// Start Git Watcher

}
