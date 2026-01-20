package webhooks

func (s *Service) SetStatus(slug string, isActive bool) error {

	// [TODO] db call to change webhook status

	s.mu.Lock()
	if v, ok := s.registry[slug]; ok {
		v.IsActive = isActive
	}
	s.mu.Unlock()

	return nil
}
