package store

import (
	"context"
	"time"

	"github.com/kunalvirwal/shogun-cd/internal/models"
	"gorm.io/gorm"
)

const CronRunPageSize = 10

type CronStore interface {
	List(ctx context.Context) ([]models.Cron, error)
	ListDeleted(ctx context.Context) ([]models.Cron, error)
	Get(ctx context.Context, id uint) (*models.Cron, error)
	Save(ctx context.Context, in *models.Cron) error
	Delete(ctx context.Context, id uint) error
	PermanentlyDelete(ctx context.Context, id uint) error
	CreateRun(ctx context.Context, in *models.CronRun) error
	FinishDispatch(ctx context.Context, in *models.CronRun) error
	Runs(ctx context.Context, id, before uint) ([]models.CronRun, error)
	InterruptDispatches(ctx context.Context) error
}

type cronStore struct {
	db *gorm.DB
}

func newCronStore(db *gorm.DB) CronStore {
	return &cronStore{db: db}
}

func (s *cronStore) List(ctx context.Context) ([]models.Cron, error) {
	rows := []models.Cron{}
	err := s.db.WithContext(ctx).Where("deleted_at IS NULL").Order("id DESC").Find(&rows).Error
	return rows, err
}

func (s *cronStore) ListDeleted(ctx context.Context) ([]models.Cron, error) {
	rows := []models.Cron{}
	err := s.db.WithContext(ctx).Where("deleted_at IS NOT NULL").Order("deleted_at DESC, id DESC").Find(&rows).Error
	return rows, err
}

// Get returns active and soft-deleted definitions for history pages.
func (s *cronStore) Get(ctx context.Context, id uint) (*models.Cron, error) {
	var cron models.Cron
	if err := s.db.WithContext(ctx).First(&cron, id).Error; err != nil {
		return nil, err
	}
	return &cron, nil
}
func (s *cronStore) Save(ctx context.Context, c *models.Cron) error {
	if c.ID == 0 {
		return s.db.WithContext(ctx).Create(c).Error
	}
	result := s.db.WithContext(ctx).Model(&models.Cron{}).Where("id = ? AND deleted_at IS NULL", c.ID).Select("Name", "Pipeline", "Schedule", "Payload", "Enabled", "UpdatedAt").Updates(c)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (s *cronStore) Delete(ctx context.Context, id uint) error {
	r := s.db.WithContext(ctx).Model(&models.Cron{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]any{"deleted_at": time.Now(), "enabled": false})
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// PermanentlyDelete removes only deleted cron definitions and their cron runs.
// Pipeline runs remain available in pipeline history.
func (s *cronStore) PermanentlyDelete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id = ? AND deleted_at IS NOT NULL", id).Delete(&models.Cron{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Where("cron_id = ?", id).Delete(&models.CronRun{}).Error
	})
}
func (s *cronStore) CreateRun(ctx context.Context, r *models.CronRun) error {
	return s.db.WithContext(ctx).Create(r).Error
}
func (s *cronStore) FinishDispatch(ctx context.Context, r *models.CronRun) error {
	return s.db.WithContext(ctx).Model(r).Select("Status", "Error", "PipelineRunID").Updates(r).Error
}
func (s *cronStore) Runs(ctx context.Context, id uint, before uint) ([]models.CronRun, error) {
	rows := []models.CronRun{}
	q := s.db.WithContext(ctx).Where("cron_id = ?", id)
	if before > 0 {
		q = q.Where("id < ?", before)
	}
	err := q.Preload("PipelineRun").Order("id DESC").Limit(CronRunPageSize + 1).Find(&rows).Error
	return rows, err
}
func (s *cronStore) InterruptDispatches(ctx context.Context) error {
	return s.db.WithContext(ctx).Model(&models.CronRun{}).Where("status = ?", "dispatching").Updates(map[string]any{"status": "interrupted", "error": "Shogun stopped during the start attempt; no retry will be made"}).Error
}
