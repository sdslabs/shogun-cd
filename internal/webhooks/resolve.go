package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
)

// fetches webhook info, verifies payload, and runs the pipeline
func (s *Service) Resolve(ctx context.Context, slug string, headers http.Header, body []byte) (*Webhook, error) {

	s.mu.RLock()
	hook, exists := s.registry[slug]
	s.mu.RUnlock()

	if !exists {
		return nil, ErrHookNotFound
	}

	if !hook.IsActive {
		return nil, ErrHookInactive
	}

	ok, err := hook.VerifyPayload(ctx, headers, body)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrAuthFailed
	}

	vals := make(map[string]string) // currently accepts flattened json
	err = json.Unmarshal(body, &vals)
	if err != nil {
		return nil, fmt.Errorf("%w : %v", ErrInvalidJSON, err)
	}

	p := hook.Pipeline
	s.logger.LogInfo("Webhook triggers pipeline : %v", p)
	if err := s.orchestrator.RunPipeline(p, pipeline.WebhookTriggerKind, vals); err != nil {
		s.logger.LogError(err)
		return nil, err
	}

	return hook, nil
}
