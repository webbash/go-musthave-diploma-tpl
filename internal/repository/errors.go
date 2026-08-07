package repository

import "errors"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrOrderNotFound  = errors.New("order not found")
	ErrDuplicateUser  = errors.New("duplicate user")
	ErrDuplicateOrder = errors.New("duplicate order")
)
