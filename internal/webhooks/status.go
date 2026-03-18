package webhooks

import "context"

func (s *Service) SetStatus(ctx context.Context, slug string, isActive bool) error {

	if err := s.store.SetStatus(ctx, slug, isActive); err != nil {
		return err
	}

	s.mu.Lock()
	if v, ok := s.registry[slug]; ok {
		v.IsActive = isActive
	}
	s.mu.Unlock()

	return nil
}
