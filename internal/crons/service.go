package crons

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/kunalvirwal/shogun-cd/internal/models"
	"github.com/kunalvirwal/shogun-cd/internal/orchestrator"
	"github.com/kunalvirwal/shogun-cd/internal/pipeline"
	"github.com/kunalvirwal/shogun-cd/internal/secrets"
	"github.com/kunalvirwal/shogun-cd/internal/store"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
	cron "github.com/robfig/cron/v3"
)

type entry struct {
	definition models.Cron
	schedule   cron.Schedule
	next       time.Time
}

// CronService manages cron definitions, run history, and scheduled dispatches.
type CronService interface {
	List() []Summary
	ListDeleted(ctx context.Context) ([]models.Cron, error)
	Get(ctx context.Context, id uint) (*models.Cron, error)
	Save(ctx context.Context, id uint, in Input) (*models.Cron, error)
	Delete(ctx context.Context, id uint) error
	PermanentlyDelete(ctx context.Context, id uint) error
	Runs(ctx context.Context, id, before uint) ([]models.CronRun, error)
	Timezone() string
	Start(ctx context.Context)
}

var _ CronService = (*service)(nil)

type service struct {
	mu            sync.Mutex
	store         store.CronStore
	orch          orchestrator.Orchestrator
	logger        utils.Logger
	secretService secrets.SecretService
	location      *time.Location
	entries       map[uint]*entry
}
type Summary struct {
	models.Cron
	NextRun *time.Time `json:"next_run,omitempty"`
}
type Input struct {
	Name     string            `json:"name"`
	Pipeline string            `json:"pipeline"`
	Schedule string            `json:"schedule"`
	Payload  map[string]string `json:"payload"`
	Enabled  bool              `json:"enabled"`
}

var ErrInvalidInput = errors.New("invalid cron configuration")

var keyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func Parse(expression string, location *time.Location) (cron.Schedule, error) {
	if len(strings.Fields(expression)) != 5 || strings.Contains(expression, "=") {
		return nil, fmt.Errorf("use five cron fields: minute hour day-of-month month day-of-week; timezone overrides are not allowed")
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(expression)
	if err != nil {
		return nil, err
	}
	schedule.(*cron.SpecSchedule).Location = location
	if schedule.Next(time.Now()).IsZero() {
		return nil, fmt.Errorf("schedule has no future occurrence")
	}
	return schedule, nil
}

func NewCronService(cronStore store.CronStore, orch orchestrator.Orchestrator, logger utils.Logger, timezone string, secretService secrets.SecretService) (CronService, error) {
	if secretService == nil {
		return nil, fmt.Errorf("cron secret service is required")
	}
	if timezone == "" {
		timezone = "UTC"
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}
	s := &service{
		store:         cronStore,
		orch:          orch,
		logger:        logger,
		secretService: secretService,
		location:      loc,
		entries:       make(map[uint]*entry),
	}
	ctx := context.Background()
	if err = cronStore.InterruptDispatches(ctx); err != nil {
		return nil, err
	}
	rows, err := cronStore.List(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	for _, c := range rows {
		schedule, e := Parse(c.Schedule, loc)
		if e != nil {
			return nil, fmt.Errorf("cron %d: %w", c.ID, e)
		}
		s.entries[c.ID] = &entry{c, schedule, schedule.Next(now)}
	}
	return s, nil
}

func (s *service) ListDeleted(ctx context.Context) ([]models.Cron, error) {
	return s.store.ListDeleted(ctx)
}

func (s *service) PermanentlyDelete(ctx context.Context, id uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.store.PermanentlyDelete(ctx, id)
}

func (s *service) Get(ctx context.Context, id uint) (*models.Cron, error) {
	return s.store.Get(ctx, id)
}

func (s *service) Timezone() string { return s.location.String() }

func (s *service) List() []Summary {
	s.mu.Lock()
	defer s.mu.Unlock()
	rows := []Summary{}
	for _, e := range s.entries {
		row := Summary{Cron: e.definition}
		next := e.next
		if !e.definition.Enabled {
			// Preview the next matching time without scheduling a paused cron.
			next = e.schedule.Next(time.Now())
		}
		row.NextRun = &next
		rows = append(rows, row)
	}
	return rows
}

func (s *service) Save(ctx context.Context, id uint, in Input) (*models.Cron, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Pipeline = strings.TrimSpace(in.Pipeline)
	if in.Name == "" || len(in.Name) > 100 {
		return nil, fmt.Errorf("%w: name must contain 1–100 bytes", ErrInvalidInput)
	}
	if len(in.Pipeline) > 50 {
		return nil, fmt.Errorf("%w: invalid pipeline name", ErrInvalidInput)
	}
	schedule, err := Parse(in.Schedule, s.location)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	found := false
	for _, p := range s.orch.ListPipelines() {
		if p.Metadata.Name == in.Pipeline {
			found = p.SupportsTrigger(pipeline.CronTriggerKind)
			break
		}
	}
	if old := s.entries[id]; !found && (old == nil || old.definition.Pipeline != in.Pipeline) {
		return nil, fmt.Errorf("%w: select an indexed pipeline declaring cron_trigger", ErrInvalidInput)
	}
	for k := range in.Payload {
		if !keyPattern.MatchString(k) {
			return nil, fmt.Errorf("%w: invalid payload key %q", ErrInvalidInput, k)
		}
	}
	if in.Payload == nil {
		in.Payload = map[string]string{}
	}
	c := models.Cron{ID: id, Name: in.Name, Pipeline: in.Pipeline, Schedule: in.Schedule, Payload: in.Payload, Enabled: in.Enabled}
	if old := s.entries[id]; old != nil {
		c.CreatedAt = old.definition.CreatedAt
	}
	if err = s.store.Save(ctx, &c); err != nil {
		return nil, err
	}
	s.entries[c.ID] = &entry{c, schedule, schedule.Next(time.Now())}
	return &c, nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.store.Delete(ctx, id); err != nil {
		return err
	}
	delete(s.entries, id)
	return nil
}

func (s *service) Runs(ctx context.Context, id, before uint) ([]models.CronRun, error) {
	return s.store.Runs(ctx, id, before)
}

func (s *service) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.tick(ctx, time.Now())
			}
		}
	}()
}

// Advance directly to the next future occurrence. An overdue minute is never backfilled.
func (s *service) tick(ctx context.Context, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.entries {
		if !e.definition.Enabled || now.Before(e.next) {
			continue
		}
		due := e.next
		e.next = e.schedule.Next(now)
		if due.Before(now.Truncate(time.Minute)) {
			continue
		}
		r := &models.CronRun{CronID: e.definition.ID, Pipeline: e.definition.Pipeline, Payload: e.definition.Payload, ScheduledAt: due, Status: "dispatching"}
		if err := s.store.CreateRun(ctx, r); err != nil {
			s.logger.LogNewError("Unable to record cron occurrence: %v", err)
			continue
		}
		go s.dispatch(ctx, r)
	}
}

func (s *service) dispatch(ctx context.Context, r *models.CronRun) {
	// Keep references in configuration and history; resolve into a separate runtime map.
	values := make(map[string]string, len(r.Payload))
	var err error
	for key, value := range r.Payload {
		var resolved string
		resolved, err = s.secretService.ResolveSecrets(ctx, value)
		if err != nil {
			// Resolver errors can contain backend details; persist only the payload key.
			err = fmt.Errorf("unable to resolve secrets in payload value %q", key)
			break
		}
		values[key] = resolved
	}
	var id uint
	if err == nil {
		id, err = s.orch.RunPipeline(context.WithValue(ctx, models.CronRunContextKey{}, r.ID), r.Pipeline, pipeline.CronTriggerKind, values)
	}
	if err != nil {
		r.Status = "failed"
		r.Error = err.Error()
	} else {
		r.Status = "started"
		r.PipelineRunID = &id
	}
	if err = s.store.FinishDispatch(ctx, r); err != nil {
		s.logger.LogNewError("Unable to save cron start outcome for run %d: %v", r.ID, err)
	}
}
