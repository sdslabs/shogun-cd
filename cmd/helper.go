package main

import (
	"context"
	"errors"

	"github.com/kunalvirwal/shogun-cd/internal/config"
	"github.com/kunalvirwal/shogun-cd/internal/users"
)

// makes the admin user upon startup (ignores if admin exists)
func seedAdmin(u users.Service, cfg config.Admin) error {
	ctx := context.Background()
	if err := u.Create(ctx, &users.CreateParams{
		Email:    cfg.Email,
		Password: cfg.Password,
	}); err != nil {
		switch {
		case errors.Is(err, users.ErrEmailTaken):
			return nil

		default:
			return err
		}
	}

	if err := u.MakeAdmin(ctx, cfg.Email); err != nil {
		return err
	}

	return nil
}
