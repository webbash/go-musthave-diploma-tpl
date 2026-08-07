package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go-musthave-diploma-tpl/internal/auth"
	"go-musthave-diploma-tpl/internal/domain"
	"go-musthave-diploma-tpl/internal/repository"
)

var ErrUnauthorized = errors.New("unauthorized")

type AuthService struct {
	repository UserRepository
	jwtSecret  string
	tokenTTL   time.Duration
}

func NewAuthService(store UserRepository, jwtSecret string, jwtTTL int) *AuthService {
	return &AuthService{repository: store, jwtSecret: jwtSecret, tokenTTL: time.Duration(jwtTTL) * time.Minute}
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
		return AuthResult{}, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.repository.CreateUser(ctx, login, passwordHash)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateUser) {
			return AuthResult{}, domain.ErrUserAlreadyExists
		}
		return AuthResult{}, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := auth.GenerateJWT(strconv.FormatInt(user.ID, 10), s.jwtSecret, s.tokenTTL)
	if err != nil {
		return AuthResult{}, fmt.Errorf("failed to generate JWT: %w", err)
	}

	return AuthResult{Token: token}, nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (AuthResult, error) {
	user, err := s.repository.FindUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return AuthResult{}, ErrUnauthorized
		}
		return AuthResult{}, fmt.Errorf("failed to find user: %w", err)
	}
	if err := auth.CheckPassword(user.PasswordHash, password); err != nil {
		return AuthResult{}, ErrUnauthorized
	}

	token, err := auth.GenerateJWT(strconv.FormatInt(user.ID, 10), s.jwtSecret, s.tokenTTL)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{Token: token}, nil
}
