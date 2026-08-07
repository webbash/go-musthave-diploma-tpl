package service_test

import (
	"context"
	"errors"
	"testing"

	"go-musthave-diploma-tpl/internal/auth"
	"go-musthave-diploma-tpl/internal/domain"
	"go-musthave-diploma-tpl/internal/model"
	"go-musthave-diploma-tpl/internal/repository"
	"go-musthave-diploma-tpl/internal/service"
	"go-musthave-diploma-tpl/internal/service/mocks"
	"go.uber.org/mock/gomock"
)

func TestAuthServiceRegister(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewAuthService(repo, "secret", 10)

		repo.EXPECT().
			CreateUser(gomock.Any(), "alice", gomock.Any()).
			DoAndReturn(func(ctx context.Context, login, passwordHash string) (model.User, error) {
				if passwordHash == "password" {
					t.Fatal("password must be hashed")
				}
				if err := auth.CheckPassword(passwordHash, "password"); err != nil {
					t.Fatalf("hashed password should be valid: %v", err)
				}
				return model.User{ID: 42}, nil
			})

		res, err := svc.Register(context.Background(), "alice", "password")
		if err != nil {
			t.Fatalf("Register() error = %v", err)
		}
		if res.Token == "" {
			t.Fatal("Register() token is empty")
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		svc := service.NewAuthService(nil, "secret", 10)
		_, err := svc.Register(context.Background(), "", "password")
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("Register() error = %v, want invalid input", err)
		}
	})

	t.Run("duplicate user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewAuthService(repo, "secret", 10)

		repo.EXPECT().
			CreateUser(gomock.Any(), "alice", gomock.Any()).
			Return(model.User{}, repository.ErrDuplicateUser)

		_, err := svc.Register(context.Background(), "alice", "password")
		if !errors.Is(err, domain.ErrUserAlreadyExists) {
			t.Fatalf("Register() error = %v, want already exists", err)
		}
	})
}

func TestAuthServiceLogin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewAuthService(repo, "secret", 10)

		hash, err := auth.HashPassword("password")
		if err != nil {
			t.Fatalf("HashPassword() error = %v", err)
		}

		repo.EXPECT().
			FindUserByLogin(gomock.Any(), "alice").
			Return(model.User{ID: 42, PasswordHash: hash}, nil)

		res, err := svc.Login(context.Background(), "alice", "password")
		if err != nil {
			t.Fatalf("Login() error = %v", err)
		}
		if res.Token == "" {
			t.Fatal("Login() token is empty")
		}
	})

	t.Run("unauthorized on missing user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewAuthService(repo, "secret", 10)

		repo.EXPECT().
			FindUserByLogin(gomock.Any(), "alice").
			Return(model.User{}, repository.ErrUserNotFound)

		_, err := svc.Login(context.Background(), "alice", "password")
		if !errors.Is(err, service.ErrUnauthorized) {
			t.Fatalf("Login() error = %v, want unauthorized", err)
		}
	})

	t.Run("unauthorized on password mismatch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockUserRepository(ctrl)
		svc := service.NewAuthService(repo, "secret", 10)

		hash, err := auth.HashPassword("password")
		if err != nil {
			t.Fatalf("HashPassword() error = %v", err)
		}

		repo.EXPECT().
			FindUserByLogin(gomock.Any(), "alice").
			Return(model.User{ID: 42, PasswordHash: hash}, nil)

		_, err = svc.Login(context.Background(), "alice", "other")
		if !errors.Is(err, service.ErrUnauthorized) {
			t.Fatalf("Login() error = %v, want unauthorized", err)
		}
	})
}
