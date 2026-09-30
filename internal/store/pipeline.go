package store

import (
	"context"
	"fmt"

	"github.com/kunalvirwal/shogun-cd/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PipelineStore interface {
	CreatePipelineRun(ctx context.Context, in *models.PipelineRun) error
	UpdatePipelineRun(ctx context.Context, in *models.PipelineRun) error
	CreatePipelineRunStep(ctx context.Context, in *models.PipelineRunStep) error
	UpdatePipelineRunStep(ctx context.Context, in *models.PipelineRunStep) error
	FetchPipelineRunDetails(ctx context.Context, pipeline string, runID uint) ([]models.PipelineRun, error)
	FetchLatestPipelineRuns(ctx context.Context, pipelines []string) (map[string]*models.PipelineRun, error)
}

// FetchLatestPipelineRuns returns at most one run per pipeline. Run IDs are
// monotonically increasing, so selecting the greatest ID per pipeline gives
// the latest execution without loading complete run histories.
func (p *pipelineStore) FetchLatestPipelineRuns(ctx context.Context, pipelines []string) (map[string]*models.PipelineRun, error) {
	result := make(map[string]*models.PipelineRun, len(pipelines))
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if len(pipelines) == 0 {
		return result, nil
	}

	subquery := p.db.
		Model(&models.PipelineRun{}).
		Select(fmt.Sprintf("MAX(%s)", models.PipelineRunColID)).
		Where(fmt.Sprintf("%s IN ?", models.PipelineRunColPipeline), pipelines).
		Group(models.PipelineRunColPipeline)

	runs, err := gorm.G[models.PipelineRun](p.db).
		Where(fmt.Sprintf("%s IN (?)", models.PipelineRunColID), subquery).
		Find(ctx)
	if err != nil {
		return nil, err
	}

	for i := range runs {
		result[runs[i].Pipeline] = &runs[i]
	}

	return result, nil
}

type pipelineStore struct {
	db *gorm.DB
}

func newPipelineStore(db *gorm.DB) PipelineStore {
	return &pipelineStore{
		db: db,
	}
}

// CreatePipelineRun persists a new run for the given pipeline. The pipeline row
// is created on demand, then locked by the counter update so concurrent runs of
// the same pipeline cannot allocate the same run number. Both statements share
// one transaction with the insert, so a failed insert cannot consume a number
// and the pipeline_runs.pipeline foreign key always finds its parent row.
func (p *pipelineStore) CreatePipelineRun(ctx context.Context, in *models.PipelineRun) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Ensure the pipeline exists.
		if err := gorm.G[models.Pipeline](tx, clause.OnConflict{DoNothing: true}).
			Create(ctx, &models.Pipeline{
				Name: in.Pipeline,
			}); err != nil {
			return err
		}

		// Lock this pipeline row so concurrent runs cannot
		// allocate the same run number.
		pipeline, err := gorm.G[models.Pipeline](tx, clause.Locking{Strength: "UPDATE"}).
			Where(&models.Pipeline{Name: in.Pipeline}).
			First(ctx)
		if err != nil {
			return err
		}

		// Allocate the current number, then advance the counter.
		in.RunNumber = pipeline.NextRunNumber

		if _, err := gorm.G[models.Pipeline](tx).
			Where(&models.Pipeline{Name: in.Pipeline}).
			Update(ctx, models.PipelineColNextRunNumber, pipeline.NextRunNumber+1); err != nil {
			return err
		}

		return gorm.G[models.PipelineRun](tx).
			Create(ctx, in)
	})
}

func (p *pipelineStore) UpdatePipelineRun(ctx context.Context, in *models.PipelineRun) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	rows, err := gorm.G[models.PipelineRun](p.db).
		Where(&models.PipelineRun{ID: in.ID}).
		Updates(ctx, *in)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (p *pipelineStore) CreatePipelineRunStep(ctx context.Context, in *models.PipelineRunStep) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return gorm.G[models.PipelineRunStep](p.db).
		Create(ctx, in)
}

func (p *pipelineStore) UpdatePipelineRunStep(ctx context.Context, in *models.PipelineRunStep) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	rows, err := gorm.G[models.PipelineRunStep](p.db).
		Where(
			fmt.Sprintf("%s = ? AND %s = ?", models.PipelineRunStepColRunID, models.PipelineRunStepColStepIndex),
			in.RunID,
			in.StepIndex,
		).
		Updates(ctx, *in)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (p *pipelineStore) FetchPipelineRunDetails(ctx context.Context, pipeline string, runID uint) ([]models.PipelineRun, error) {
	filter := models.PipelineRun{
		Pipeline: pipeline,
		ID:       runID,
	}

	runs, err := gorm.G[models.PipelineRun](p.db).
		Preload("Steps", func(db gorm.PreloadBuilder) error {
			db.Order(fmt.Sprintf("%v asc", models.PipelineRunStepColStepIndex))
			return nil
		}).
		Where(&filter).
		Order(fmt.Sprintf("%v desc", models.PipelineRunColStartedAt)).
		Find(ctx)

	if err != nil {
		return nil, err
	}

	if len(runs) == 0 {
		return nil, ErrRecordNotFound
	}

	return runs, nil
}
