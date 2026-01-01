package webhooks

func (s *service) Find(filter func(*Webhook) bool) []*Webhook {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*Webhook, 0)

	for _, w := range s.cache {
		if filter == nil || filter(w) {
			results = append(results, w)
		}
	}

	return results
}

func FilterByPipeline(pipeline string) func(*Webhook) bool {
	return func(w *Webhook) bool {
		return w.Pipeline == pipeline
	}
}
