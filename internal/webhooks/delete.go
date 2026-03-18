package webhooks

import "context"

func (s *Service) Delete(ctx context.Context, slug string) error {

	if err := s.store.Delete(ctx, slug); err != nil {
		return err
	}

	s.mu.Lock()
	delete(s.registry, slug)
	s.mu.Unlock()

	return nil
}
