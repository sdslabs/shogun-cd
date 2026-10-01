package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/http/response"
	"github.com/kunalvirwal/shogun-cd/internal/models"
	"github.com/kunalvirwal/shogun-cd/internal/orchestrator"
	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
	pipelineSteps "github.com/kunalvirwal/shogun-cd/internal/pipeline/steps"
	"github.com/kunalvirwal/shogun-cd/internal/store"
	"github.com/kunalvirwal/shogun-cd/internal/target"
)

func TestListAllPipelinesIncludesLatestRunSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	startedAt := time.Date(2026, time.August, 24, 10, 30, 0, 0, time.UTC)
	finishedAt := startedAt.Add(12 * time.Second)
	succeeded := true

	orch := &systemTestOrchestrator{pipelines: []*pipeline.Pipeline{{
		Metadata: pipeline.Metadata{Name: "deploy-api", Enabled: true},
		Spec: pipeline.Spec{
			Triggers: []pipeline.Trigger{{Type: string(pipeline.WebhookTriggerKind)}},
			Steps:    []pipelineSteps.StepWrapper{{Step: &pipelineSteps.ExecStep{Target: "{{TARGET}}"}}},
		},
	}}}
	pipelineStore := &systemTestPipelineStore{latest: map[string]*models.PipelineRun{
		"deploy-api": {
			ID:          42,
			Pipeline:    "deploy-api",
			TriggerKind: string(pipeline.WebhookTriggerKind),
			Status:      models.PipelineRunStatusSucceeded,
			Success:     &succeeded,
			StartedAt:   startedAt,
			FinishedAt:  &finishedAt,
		},
	}}
	handler := &Handler{
		response: response.NewResponder(false),
		store:    &store.Store{Pipeline: pipelineStore},
		orch:     orch,
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/system/pipelines", nil)
	handler.ListAllPipelines(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Data []struct {
			Name    string `json:"name"`
			LastRun *struct {
				ID     uint   `json:"id"`
				Status string `json:"status"`
			} `json:"last_run"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Data) != 1 || payload.Data[0].Name != "deploy-api" {
		t.Fatalf("unexpected pipeline response: %+v", payload.Data)
	}
	if payload.Data[0].LastRun == nil || payload.Data[0].LastRun.ID != 42 || payload.Data[0].LastRun.Status != "succeeded" {
		t.Fatalf("unexpected latest run summary: %+v", payload.Data[0].LastRun)
	}
}

func TestListAllTargetsIncludesAccessSecretReference(t *testing.T) {
	gin.SetMode(gin.TestMode)
	targetResource := &target.Target{
		Metadata: target.Metadata{Name: "production", Type: target.ServerType},
		Spec: target.Spec{
			Host:         "deploy.example.com",
			User:         "shogun",
			Port:         22,
			AccessSecret: "PRODUCTION_SSH_KEY",
		},
	}
	handler := &Handler{
		response: response.NewResponder(false),
		orch:     &systemTestOrchestrator{targets: []*target.Target{targetResource}},
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/system/targets", nil)
	handler.ListAllTargets(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload struct {
		Data []struct {
			Name         string `json:"name"`
			AccessSecret string `json:"access_secret"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Data) != 1 || payload.Data[0].Name != "production" || payload.Data[0].AccessSecret != "PRODUCTION_SSH_KEY" {
		t.Fatalf("unexpected target response: %+v", payload.Data)
	}
}

func TestPipelineStepDefinitionIncludesAllManifestFields(t *testing.T) {
	tests := []struct {
		name string
		step pipelineSteps.Step
		want string
	}{
		{
			name: "mutate",
			step: &pipelineSteps.MutateStep{
				TriggerWhen: string(pipeline.WebhookTriggerKind),
				Changes: []pipelineSteps.Change{{
					File:        "services/api/compose.yaml",
					UpdateField: "services.api.image",
					Value:       "{{IMAGE}}",
				}},
			},
			want: `{"changes":[{"file":"services/api/compose.yaml","update_field":"services.api.image","value":"{{IMAGE}}"}],"trigger_when":"ci_webhook"}`,
		},
		{
			name: "sync",
			step: &pipelineSteps.SyncStep{
				TriggerWhen: string(pipeline.GitChangesTriggerKind),
				Target:      "production",
				Files:       []pipelineSteps.FileUpdate{{Src: "compose.yaml", Dst: "~/compose.yaml"}},
			},
			want: `{"files":[{"src":"compose.yaml","dst":"~/compose.yaml"}],"target":"production","trigger_when":"git_changes"}`,
		},
		{
			name: "exec",
			step: &pipelineSteps.ExecStep{
				TriggerWhen: string(pipeline.WebhookTriggerKind),
				Target:      "production",
				Commands:    []string{"docker compose pull", "docker compose up -d"},
			},
			want: `{"commands":["docker compose pull","docker compose up -d"],"target":"production","trigger_when":"ci_webhook"}`,
		},
		{
			name: "apply",
			step: &pipelineSteps.ApplyStep{
				TriggerWhen: string(pipeline.GitChangesTriggerKind),
				Target:      "cluster",
				Files:       []string{"deployment.yaml", "service.yaml"},
			},
			want: `{"files":["deployment.yaml","service.yaml"],"target":"cluster","trigger_when":"git_changes"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := json.Marshal(pipelineStepDefinition(test.step))
			if err != nil {
				t.Fatalf("encode definition: %v", err)
			}
			if string(encoded) != test.want {
				t.Fatalf("unexpected definition:\nwant: %s\n got: %s", test.want, encoded)
			}
		})
	}
}

type systemTestOrchestrator struct {
	pipelines []*pipeline.Pipeline
	targets   []*target.Target
}

var _ orchestrator.Orchestrator = (*systemTestOrchestrator)(nil)

func (*systemTestOrchestrator) Start()                      {}
func (*systemTestOrchestrator) RunIndexer()                 {}
func (*systemTestOrchestrator) StartPoller(context.Context) {}
func (*systemTestOrchestrator) RunPipeline(context.Context, string, pipeline.TriggerKind, map[string]string) (uint, error) {
	return 0, nil
}
func (*systemTestOrchestrator) PipelineExists(string) bool            { return true }
func (o *systemTestOrchestrator) ListPipelines() []*pipeline.Pipeline { return o.pipelines }
func (o *systemTestOrchestrator) ListTargets() []*target.Target       { return o.targets }

type systemTestPipelineStore struct {
	latest map[string]*models.PipelineRun
}

var _ store.PipelineStore = (*systemTestPipelineStore)(nil)

func (*systemTestPipelineStore) CreatePipelineRun(context.Context, *models.PipelineRun) error {
	return nil
}
func (*systemTestPipelineStore) UpdatePipelineRun(context.Context, *models.PipelineRun) error {
	return nil
}
func (*systemTestPipelineStore) CreatePipelineRunStep(context.Context, *models.PipelineRunStep) error {
	return nil
}
func (*systemTestPipelineStore) UpdatePipelineRunStep(context.Context, *models.PipelineRunStep) error {
	return nil
}
func (*systemTestPipelineStore) FetchPipelineRunDetails(context.Context, string, uint) ([]models.PipelineRun, error) {
	return nil, nil
}
func (s *systemTestPipelineStore) FetchLatestPipelineRuns(context.Context, []string) (map[string]*models.PipelineRun, error) {
	return s.latest, nil
}
