package webhooks

func (s *Service) Delete(slug string) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.registry[slug]; !exists {
		return HookNotFoundError
	}

	delete(s.registry, slug)

	// [TODO] delete from DB

	return nil
}
