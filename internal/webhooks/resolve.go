package webhooks

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
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

	vals := make(map[string]string) // currently accepts flattened json
	err = json.Unmarshal(body, &vals)
	if err != nil {
		return nil, fmt.Errorf("%w : %v", InvalidJSONError, err)
	}

	p := hook.Pipeline
	go func(p string, vals map[string]string) {
		s.logger.LogInfo("Webhook triggers pipeline : %v", p)
		if err := s.orchestrator.RunPipeline(p, pipeline.WebhookTriggerKind, vals); err != nil {
			s.logger.LogError(err)
		}
	}(p, vals)

	return hook, nil
}
