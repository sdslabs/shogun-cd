package users

import (
	"context"
	"testing"

	"github.com/kunalvirwal/shogun-cd/internal/models"
	"github.com/kunalvirwal/shogun-cd/internal/store"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

func TestCreateAssignsUserRole(t *testing.T) {
	userStore := &createUserStoreStub{}
	service := NewUserService(
		&store.Store{User: userStore},
		utils.NewLogger(utils.ProdLevel, false),
	)

	err := service.Create(context.Background(), &CreateParams{
		Email:    "new-user@shogun.dev",
		Password: "correct-horse-battery-staple",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if userStore.created == nil {
		t.Fatal("expected user to be persisted")
	}
	if userStore.created.Role != models.RoleUser {
		t.Fatalf("expected role %q, got %q", models.RoleUser, userStore.created.Role)
	}
}

type createUserStoreStub struct {
	created *models.User
}

var _ store.UserStore = (*createUserStoreStub)(nil)

func (s *createUserStoreStub) Create(_ context.Context, user *models.User) error {
	copy := *user
	s.created = &copy
	return nil
}

func (*createUserStoreStub) MakeAdmin(context.Context, string) error {
	return nil
}

func (*createUserStoreStub) FindOne(context.Context, *models.User) (*models.User, error) {
	return nil, store.ErrRecordNotFound
}

func (*createUserStoreStub) FindMany(context.Context, *models.User) ([]*models.User, error) {
	return nil, nil
}
