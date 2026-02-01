package webhooks

import "context"

type WebhookFilter func(*Webhook) bool

func (s *Service) Find(ctx context.Context, filter WebhookFilter) []*Webhook {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*Webhook, 0)

	for _, w := range s.registry {
		if filter == nil || filter(w) {
			results = append(results, w)
		}
	}

	return results
}

func FilterByPipeline(pipeline string) WebhookFilter {
	return func(w *Webhook) bool {
		return w.Pipeline == pipeline
	}
}
