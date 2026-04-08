package webhooks

import (
	"context"
	"errors"

	"github.com/kunalvirwal/shogun-cd/internal/store"
)

func (s *Service) SetStatus(ctx context.Context, slug string, isActive bool) error {

	if err := s.store.SetStatus(ctx, slug, isActive); err != nil {
		switch {
		case errors.Is(err, store.ErrRecordNotFound):
			return ErrHookNotFound

		default:
			return err
		}
	}

	s.mu.Lock()
	if v, ok := s.registry[slug]; ok {
		v.IsActive = isActive
	}
	s.mu.Unlock()

	return nil
}
