package webhooks

import (
	"context"
	"sync"

	"github.com/kunalvirwal/shogun-cd/internal/models"
)

// sync the local registry with database
// current implementation of this function is meant to be run during startup
func (s *Service) Load(ctx context.Context) error {
	dbHooks, err := s.store.Load(ctx)
	if err != nil {
		return err
	}

	newRegistry := make(map[string]*Webhook, len(dbHooks))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, dbHook := range dbHooks {
		wg.Add(1)

		isActive := false // Default safe value if db has NULL for the boolean pointer
		if dbHook.IsActive != nil {
			isActive = *dbHook.IsActive
		}

		go func(hook *models.Webhook, isActive bool) {
			defer wg.Done()
			wh := &Webhook{
				ID:        hook.ID,
				Slug:      hook.Slug,
				Secret:    hook.Secret, // [TODO] [CRITICAL] decrypt the secret to store in registry
				Pipeline:  hook.Pipeline,
				Alias:     hook.Alias,
				CreatedBy: hook.CreatedBy,
				CreatedAt: hook.CreatedAt,
				IsActive:  isActive,
			}
			mu.Lock()
			newRegistry[wh.Slug] = wh
			mu.Unlock()
		}(dbHook, isActive)
	}
	wg.Wait()

	s.mu.Lock()
	s.registry = newRegistry
	s.mu.Unlock()

	s.logger.LogInfo("%v Webhooks loaded from Database.", len(dbHooks))

	return nil
}
