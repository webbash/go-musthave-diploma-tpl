package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-musthave-diploma-tpl/internal/auth"
	"go-musthave-diploma-tpl/internal/domain"
	"go-musthave-diploma-tpl/internal/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, login, passwordHash string) (model.User, error)
	FindUserByLogin(ctx context.Context, login string) (model.User, error)
}

type AuthService struct {
	repository UserRepository
	jwtSecret  string
	tokenTTL   time.Duration
}

func NewAuthService(store UserRepository, jwtSecret string) *AuthService {
	return &AuthService{repository: store, jwtSecret: jwtSecret, tokenTTL: 24 * time.Hour}
}

type AuthResult struct {
	Token string
}

func (s *AuthService) Register(ctx context.Context, login, password string) (AuthResult, error) {
	if login == "" || password == "" {
		return AuthResult{}, fmt.Errorf("empty credentials: %w", domain.ErrInvalidInput)
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return AuthResult{}, err
	}

	user, err := s.repository.CreateUser(ctx, login, passwordHash)
	if err != nil {
		if err == domain.ErrAlreadyExists {
			return AuthResult{}, domain.ErrAlreadyExists
		}
		return AuthResult{}, err
	}

	token, err := auth.GenerateJWT(strconv.FormatInt(user.ID, 10), s.jwtSecret, s.tokenTTL)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{Token: token}, nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (AuthResult, error) {
	user, err := s.repository.FindUserByLogin(ctx, login)
	if err != nil {
		return AuthResult{}, domain.ErrUnauthorized
	}
	if err := auth.CheckPassword(user.PasswordHash, password); err != nil {
		return AuthResult{}, domain.ErrUnauthorized
	}

	token, err := auth.GenerateJWT(strconv.FormatInt(user.ID, 10), s.jwtSecret, s.tokenTTL)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{Token: token}, nil
}
