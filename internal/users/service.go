package users

import (
	"context"
	"errors"

	"github.com/kunalvirwal/shogun-cd/internal/models"
	"github.com/kunalvirwal/shogun-cd/internal/store"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service interface {
	Create(ctx context.Context, in *CreateParams) error
	MakeAdmin(ctx context.Context, email string) error
	FindOne(ctx context.Context, filter *FilterParams) (*models.User, error)
	FindMany(ctx context.Context, filter *FilterParams) ([]*models.User, error)
}

var (
	ErrUserNotFound     = errors.New("User Not Found")
	ErrEmailTaken       = errors.New("User with this email already exists")
	ErrAccountSuspended = errors.New("Account Suspended")
)

type service struct {
	store  store.UserStore
	logger utils.Logger
}

func NewUserService(s *store.Store, l utils.Logger) Service {
	return &service{
		store:  s.User,
		logger: l,
	}
}

func (s *service) Create(ctx context.Context, in *CreateParams) error {
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 10)
	if err != nil {
		return err
	}

	if err := s.store.Create(ctx, &models.User{
		Email:        in.Email,
		PasswordHash: string(pwdHash),
	}); err != nil {
		switch {
		case errors.Is(err, gorm.ErrDuplicatedKey):
			return ErrEmailTaken

		default:
			return err
		}
	}
	return nil
}

func (s *service) MakeAdmin(ctx context.Context, email string) error {
	return s.store.MakeAdmin(ctx, email)
}

func (s *service) FindOne(ctx context.Context, filter *FilterParams) (*models.User, error) {
	if filter == nil {
		filter = &FilterParams{}
	}
	user, err := s.store.FindOne(ctx, &models.User{
		Email:    filter.Email,
		IsActive: filter.IsActive,
	})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrRecordNotFound):
			return nil, ErrUserNotFound

		default:
			return nil, err
		}
	}
	return user, nil
}

func (s *service) FindMany(ctx context.Context, filter *FilterParams) ([]*models.User, error) {
	if filter == nil {
		filter = &FilterParams{}
	}
	return s.store.FindMany(ctx, &models.User{
		Email:    filter.Email,
		IsActive: filter.IsActive,
	})
}
