package webhooks

import "errors"

func (s *service) Resolve(slug string) (*Webhook, error) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	hook, exists := s.cache[slug]
	if !exists {
		return nil, errors.New(WebhookNotFound)
	}

	if !hook.IsActive {
		return nil, errors.New(WebhookInactive)
	}

	return hook, nil
}
