package webhooks

import (
	"net/http"
)

// fetches webhook info, verifies payload, and runs the pipeline
func (s *Service) Resolve(slug string, headers http.Header, body []byte) (*Webhook, error) {

	s.mu.RLock()
	hook, exists := s.registry[slug]
	s.mu.RUnlock()

	if !exists {
		return nil, HookNotFoundError
	}

	if !hook.IsActive {
		return nil, HookInactiveError
	}

	ok, err := hook.VerifyPayload(headers, body)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, AuthFailedError
	}

	// run the pipeline

	return hook, nil
}
