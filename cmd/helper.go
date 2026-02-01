package main

import (
	"context"
	"errors"

	"github.com/kunalvirwal/shogun-cd/api/dto"
	"github.com/kunalvirwal/shogun-cd/internal/config"
	"github.com/kunalvirwal/shogun-cd/internal/store"
)

// makes the admin user upon startup (ignores if admin exists)
func seedAdmin(ctx context.Context, s *store.Store, cfg *config.Config) error {
	if err := s.User.Create(ctx, &dto.LoginInput{
		Email:    cfg.ApiConfig.Admin.Email,
		Password: cfg.ApiConfig.Admin.Password,
	}); err != nil {
		switch {
		case errors.Is(err, store.ErrEmailTaken):
			return nil

		default:
			return err
		}
	}

	if err := s.User.MakeAdmin(ctx, &dto.LoginInput{
		Email: cfg.ApiConfig.Admin.Email,
	}); err != nil {
		return err
	}

	return nil
}
