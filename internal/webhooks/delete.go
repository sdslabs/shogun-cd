package webhooks

func (s *Service) Delete(slug string) error {

	// [TODO] delete from DB

	s.mu.Lock()
	delete(s.registry, slug)
	s.mu.Unlock()

	return nil
}
