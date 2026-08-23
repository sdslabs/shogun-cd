package store

import (
	"context"
	"fmt"

	"github.com/kunalvirwal/shogun-cd/internal/models"
	"gorm.io/gorm"
)

type PipelineStore interface {
	CreatePipelineRun(ctx context.Context, in *models.PipelineRun) error
	UpdatePipelineRun(ctx context.Context, in *models.PipelineRun) error
	CreatePipelineRunStep(ctx context.Context, in *models.PipelineRunStep) error
	UpdatePipelineRunStep(ctx context.Context, in *models.PipelineRunStep) error
	FetchPipelineRunDetails(ctx context.Context, pipeline string, runID uint) ([]models.PipelineRun, error)
}

type pipelineStore struct {
	db *gorm.DB
}

func newPipelineStore(db *gorm.DB) PipelineStore {
	return &pipelineStore{
		db: db,
	}
}

func (p *pipelineStore) CreatePipelineRun(ctx context.Context, in *models.PipelineRun) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return gorm.G[models.PipelineRun](p.db).
		Create(ctx, in)
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
		Order(fmt.Sprintf("%v desc", models.PipelineRunStepColStartedAt)).
		Find(ctx)

	if err != nil {
		return nil, err
	}

	if len(runs) == 0 {
		return nil, ErrRecordNotFound
	}

	return runs, nil
}
