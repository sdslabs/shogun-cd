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
	FindOne(ctx context.Context, filter *FilterParams) (User, error)
	FindMany(ctx context.Context, filter *FilterParams) ([]User, error)
}

var (
	ErrUserNotFound     = errors.New("User Not Found")
	ErrEmailTaken       = errors.New("User with this email already exists")
	ErrAccountSuspended = errors.New("Account Suspended")
)

type User struct {
	Email        string `json:"email"`
	Role         string `json:"role"`
	IsActive     bool   `json:"is_active"`
	PasswordHash string `json:"-"`
}

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

func (s *service) FindOne(ctx context.Context, filter *FilterParams) (User, error) {
	if filter == nil {
		filter = &FilterParams{}
	}
	var user User
	data, err := s.store.FindOne(ctx, &models.User{
		Email:    filter.Email,
		IsActive: filter.IsActive,
	})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrRecordNotFound):
			return user, ErrUserNotFound

		default:
			return user, err
		}
	}
	isActive := false
	if data.IsActive != nil {
		isActive = *data.IsActive
	}
	user = User{
		Email:        data.Email,
		Role:         data.Role,
		IsActive:     isActive,
		PasswordHash: data.PasswordHash,
	}

	return user, nil
}

func (s *service) FindMany(ctx context.Context, filter *FilterParams) ([]User, error) {
	if filter == nil {
		filter = &FilterParams{}
	}
	data, err := s.store.FindMany(ctx, &models.User{
		Email:    filter.Email,
		IsActive: filter.IsActive,
	})
	if err != nil {
		return nil, err
	}

	userArr := make([]User, len(data))
	for k, v := range data {
		isActive := false
		if v.IsActive != nil {
			isActive = *v.IsActive
		}
		userArr[k] = User{
			Email:        v.Email,
			Role:         v.Role,
			IsActive:     isActive,
			PasswordHash: v.PasswordHash,
		}
	}

	return userArr, nil
}
