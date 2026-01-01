package webhooks

import "errors"

func (s *service) Delete(slug string) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.cache[slug]; !exists {
		return errors.New(WebhookNotFound)
	}

	delete(s.cache, slug)

	// [TODO] delete from DB

	return nil
}
