package domain

import "errors"

var (
	ErrAlreadyExists   = errors.New("already exists")
	ErrNotFound        = errors.New("not found")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrConflict        = errors.New("conflict")
	ErrInvalidInput    = errors.New("invalid input")
	ErrInvalidOrder    = errors.New("invalid order")
	ErrInsufficientSum = errors.New("insufficient balance")
)
