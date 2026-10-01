package dto

type StartPipelineRunInput struct {
	Values map[string]string `json:"values"`
}

type StartPipelineRunData struct {
	RunID uint `json:"run_id"`
}
