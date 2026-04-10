package store

import (
	"context"

	"github.com/kunalvirwal/shogun-cd/internal/models"
	"gorm.io/gorm"
)

type PipelineStore interface {
	SavePipelineRunWithSteps(ctx context.Context, in *models.PipelineRunDetails) error
}

type pipelineStore struct {
	db *gorm.DB
}

func newPipelineStore(db *gorm.DB) PipelineStore {
	return &pipelineStore{
		db: db,
	}
}

func (p *pipelineStore) SavePipelineRunWithSteps(ctx context.Context, in *models.PipelineRunDetails) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := gorm.G[models.PipelineRun](tx).
			Create(ctx, &in.Run)
		if err != nil {
			return err
		}

		for i := range in.Steps {
			in.Steps[i].RunID = in.Run.ID
		}

		err = gorm.G[[]models.PipelineRunStep](tx).
			Create(ctx, &in.Steps)
		if err != nil {
			return err
		}

		return nil
	})
}
