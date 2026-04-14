package store

import (
	"context"
	"fmt"

	"github.com/kunalvirwal/shogun-cd/internal/models"
	"gorm.io/gorm"
)

type PipelineStore interface {
	SavePipelineRunWithSteps(ctx context.Context, in *models.PipelineRun) error
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

func (p *pipelineStore) SavePipelineRunWithSteps(ctx context.Context, in *models.PipelineRun) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return gorm.G[models.PipelineRun](p.db).
		Create(ctx, in)
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
