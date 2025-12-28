package orchestrator

func (o *orchestrator) RunIndexer() {

}

// 	gitEventsChan := p.gitService.GetPullEvents()

// 	go func() {
// 		for {
// 			select {
// 			case changedFiles := <-gitEventsChan:
// 				// Make a copy since we're processing asynchronously
// 				// and the slice might be reused by git service
// 			}
// 		}
// 	}()
// }
