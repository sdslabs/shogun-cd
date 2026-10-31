package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/http/apiutils"
	"github.com/kunalvirwal/shogun-cd/api/http/controllers"
	"github.com/kunalvirwal/shogun-cd/api/http/middlewares"
	"github.com/kunalvirwal/shogun-cd/api/http/response"
	"github.com/kunalvirwal/shogun-cd/internal/config"
	"github.com/kunalvirwal/shogun-cd/internal/crons"
	"github.com/kunalvirwal/shogun-cd/internal/models"
	"github.com/kunalvirwal/shogun-cd/internal/orchestrator"
	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
	"github.com/kunalvirwal/shogun-cd/internal/secrets"
	"github.com/kunalvirwal/shogun-cd/internal/target"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
	"gorm.io/gorm"
)

type cronRepo struct{ runs []models.CronRun }

func (cronRepo) List(context.Context) ([]models.Cron, error) {
	return []models.Cron{{ID: 1, Name: "daily", Schedule: "0 2 * * *", Payload: map[string]string{"TOKEN": "private-payload"}}}, nil
}
func (cronRepo) ListDeleted(context.Context) ([]models.Cron, error) {
	deleted := time.Now()
	return []models.Cron{{ID: 1, Name: "Deleted nightly", Payload: map[string]string{"TOKEN": "private-payload"}, DeletedAt: &deleted}}, nil
}
func (cronRepo) PermanentlyDelete(context.Context, uint) error { return nil }
func (cronRepo) Get(_ context.Context, id uint) (*models.Cron, error) {
	if id != 1 {
		return nil, gorm.ErrRecordNotFound
	}
	deleted := time.Now()
	return &models.Cron{ID: id, Name: "Deleted nightly", Pipeline: "deploy", Schedule: "0 2 * * *", Payload: map[string]string{"TOKEN": "private-payload"}, DeletedAt: &deleted}, nil
}
func (cronRepo) Save(context.Context, *models.Cron) error              { return nil }
func (cronRepo) Delete(context.Context, uint) error                    { return nil }
func (cronRepo) CreateRun(context.Context, *models.CronRun) error      { return nil }
func (cronRepo) FinishDispatch(context.Context, *models.CronRun) error { return nil }
func (r cronRepo) Runs(context.Context, uint, uint) ([]models.CronRun, error) {
	if r.runs != nil {
		return r.runs, nil
	}
	return []models.CronRun{{Status: "failed", Error: "pipeline is disabled"}}, nil
}
func (cronRepo) InterruptDispatches(context.Context) error { return nil }

type cronOrchestrator struct{}

func (cronOrchestrator) RunPipeline(context.Context, string, pipeline.TriggerKind, map[string]string) (uint, error) {
	return 1, nil
}

type cronSecretService struct{}

func (cronSecretService) ResolveSecrets(_ context.Context, value string) (string, error) {
	return value, nil
}
func (cronOrchestrator) ListPipelines() []*pipeline.Pipeline { return nil }
func TestCronPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	cfg.ApiConfig.JWT.Secret = "test-secret"
	res := response.NewResponder(false)
	logger := utils.NewLogger(0, false)
	service, err := crons.NewCronService(cronRepo{}, cronOrchestrator{}, logger, "Asia/Kolkata", cronSecretService{})
	if err != nil {
		t.Fatal(err)
	}
	handler := controllers.NewHandler(logger, cfg, res, nil, nil, nil, nil, nil, service)
	router := gin.New()
	initRoutes(router, middlewares.NewManager(logger, cfg, res, nil), handler)
	for _, tc := range []struct {
		role, method, path string
		want               int
	}{
		{"", "GET", "/crons", 401},
		{"", "GET", "/crons/deleted", 401},
		{"", "GET", "/crons/1", 401},
		{"", "GET", "/crons/1/runs", 401},
		{"", "POST", "/admin/crons", 401},
		{"", "PATCH", "/admin/crons/1", 401},
		{"", "DELETE", "/admin/crons/1", 401},
		{"", "DELETE", "/admin/crons/1/permanent", 401},
		{"admin", "PATCH", "/admin/crons/1", 400},
		{"user", "GET", "/crons", 200},
		{"admin", "GET", "/crons", 200}, {"user", "GET", "/crons/1", 200}, {"user", "GET", "/crons/deleted", 200}, {"admin", "GET", "/crons/deleted", 200}, {"admin", "GET", "/crons/1", 200}, {"user", "GET", "/crons/999", 404}, {"user", "GET", "/crons/1/runs", 200},
		{"user", "POST", "/admin/crons", 403}, {"user", "PATCH", "/admin/crons/1", 403}, {"user", "DELETE", "/admin/crons/1", 403}, {"user", "DELETE", "/admin/crons/1/permanent", 403}, {"admin", "DELETE", "/admin/crons/1/permanent", 200},
		{"admin", "POST", "/admin/crons", 400}, {"admin", "DELETE", "/admin/crons/1", 200},
	} {
		t.Run(tc.role+tc.method+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader("{}"))
			req.Header.Set("Content-Type", "application/json")
			if tc.role != "" {
				token, e := apiutils.GenerateToken("test@example.com", tc.role, cfg.ApiConfig.JWT.Secret, 1)
				if e != nil {
					t.Fatal(e)
				}
				req.Header.Set("Authorization", "Bearer "+token)
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if tc.method == "GET" && (tc.path == "/crons" || tc.path == "/crons/1" || tc.path == "/crons/deleted") && tc.role != "" {
				if tc.path == "/crons/1" && (!strings.Contains(rec.Body.String(), "Deleted nightly") || !strings.Contains(rec.Body.String(), "deleted_at")) {
					t.Fatal("deleted metadata missing")
				}
				if tc.role == "user" && strings.Contains(rec.Body.String(), "private-payload") {
					t.Fatal("reader received payload")
				}
				if tc.role == "admin" && !strings.Contains(rec.Body.String(), "private-payload") {
					t.Fatal("admin payload missing")
				}
			}
			if rec.Code != tc.want {
				t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCronRunPageSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, count := range []int{0, 10, 11} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			rows := make([]models.CronRun, count)
			for i := range rows {
				rows[i] = models.CronRun{ID: uint(count - i), Status: "failed"}
			}
			logger := utils.NewLogger(0, false)
			service, err := crons.NewCronService(cronRepo{runs: rows}, cronOrchestrator{}, logger, "UTC", cronSecretService{})
			if err != nil {
				t.Fatal(err)
			}
			responder := response.NewResponder(false)
			h := controllers.NewHandler(logger, &config.Config{}, responder, nil, nil, nil, nil, nil, service)
			router := gin.New()
			router.GET("/crons/:cron/runs", h.CronRuns)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest("GET", "/crons/1/runs", nil))
			if rec.Code != 200 {
				t.Fatal(rec.Body.String())
			}
			var body struct {
				Data struct {
					Runs    []models.CronRun `json:"runs"`
					HasMore bool             `json:"has_more"`
				} `json:"data"`
			}
			if err = json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			want := count
			if want > 10 {
				want = 10
			}
			if len(body.Data.Runs) != want || body.Data.HasMore != (count > 10) {
				t.Fatalf("wrong page: %+v", body.Data)
			}
			if count > 0 && body.Data.Runs[0].ID != uint(count) {
				t.Fatal("newest run missing")
			}
		})
	}
}

var _ orchestrator.Orchestrator = cronOrchestrator{}
var _ secrets.SecretService = cronSecretService{}

func (cronOrchestrator) Start()                        { panic("unexpected Start call") }
func (cronOrchestrator) RunIndexer()                   { panic("unexpected RunIndexer call") }
func (cronOrchestrator) StartPoller(context.Context)   { panic("unexpected StartPoller call") }
func (cronOrchestrator) PipelineExists(string) bool    { panic("unexpected PipelineExists call") }
func (cronOrchestrator) ListTargets() []*target.Target { panic("unexpected ListTargets call") }

func (cronSecretService) SetSecrets(context.Context, []secrets.CreateParams) error {
	panic("unexpected SetSecrets call")
}
func (cronSecretService) FindMany(context.Context, *secrets.FilterParams) ([]secrets.Secret, error) {
	panic("unexpected FindMany call")
}
func (cronSecretService) FetchSecret(context.Context, string) (string, error) {
	panic("unexpected FetchSecret call")
}
func (cronSecretService) DeleteSecret(context.Context, string) error {
	panic("unexpected DeleteSecret call")
}
