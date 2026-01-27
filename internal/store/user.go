package store

import (
	"context"
	"errors"

	"github.com/kunalvirwal/shogun-cd/api/dto"
	"github.com/kunalvirwal/shogun-cd/api/http/apiutils"
	"github.com/kunalvirwal/shogun-cd/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailTaken       = errors.New("User with this email already exists")
	ErrUserNotFound     = errors.New("User Not Found")
	ErrAccountSuspended = errors.New("Account Suspended")
)

type UserStore interface {
	Create(ctx context.Context, in *dto.LoginInput) error
	MakeAdmin(ctx context.Context, in *dto.LoginInput) error
	FindOne(ctx context.Context, filter *dto.UserFilter) (*models.User, error)
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
func (u *userStore) Create(ctx context.Context, in *dto.LoginInput) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 10)
	if err != nil {
		return err
	}
	// role is set to `user` by default
	err = gorm.G[models.User](u.db).Create(ctx, &models.User{
		Email:    in.Email,
		Password: string(pwdHash),
	})
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrDuplicatedKey):
			return ErrEmailTaken

		default:
			return err
		}
	}
	return nil
}

// escelates user's role to 'admin'
func (u *userStore) MakeAdmin(ctx context.Context, in *dto.LoginInput) error {
	_, err := gorm.G[models.User](u.db).
		Where(&models.User{
			Email: in.Email,
		}).
		Updates(ctx, models.User{
			Role: string(apiutils.RoleAdmin),
		})

	if err != nil {
		return err
	}
	return nil
}

func (u *userStore) FindOne(ctx context.Context, filter *dto.UserFilter) (*models.User, error) {
	user, err := gorm.G[*models.User](u.db).
		Where(&filter).
		First(ctx)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrUserNotFound
		default:
			return nil, err
		}
	}
	return user, nil
}

func (u *userStore) FindMany(ctx context.Context, filter *dto.UserFilter) ([]*models.User, error) {
	users, err := gorm.G[*models.User](u.db).
		Where(filter).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}
