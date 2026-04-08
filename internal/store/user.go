package store

import (
	"context"
	"errors"

	"github.com/kunalvirwal/shogun-cd/internal/models"
	"gorm.io/gorm"
)

type UserStore interface {
	Create(ctx context.Context, in *models.User) error
	MakeAdmin(ctx context.Context, email string) error
	FindOne(ctx context.Context, filter *models.User) (*models.User, error)
	FindMany(ctx context.Context, filter *models.User) ([]*models.User, error)
}

type userStore struct {
	db *gorm.DB
}

func newUserStore(db *gorm.DB) UserStore {
	return &userStore{
		db: db,
	}
}

// create a new user with the role 'admin'
func (u *userStore) Create(ctx context.Context, in *models.User) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// role is set to `user` by default
	return gorm.G[models.User](u.db).
		Create(ctx, in)
}

// escelates user's role to 'admin'
func (u *userStore) MakeAdmin(ctx context.Context, email string) error {
	rows, err := gorm.G[models.User](u.db).
		Where(&models.User{
			Email: email,
		}).
		Updates(ctx, models.User{
			Role: models.RoleAdmin,
		})
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (u *userStore) FindOne(ctx context.Context, filter *models.User) (*models.User, error) {
	user, err := gorm.G[*models.User](u.db).
		Where(filter).
		First(ctx)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return user, nil
}

func (u *userStore) FindMany(ctx context.Context, filter *models.User) ([]*models.User, error) {
	users, err := gorm.G[*models.User](u.db).
		Where(filter).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}
