package domain

import "errors"

var (
	ErrUserAlreadyExists                = errors.New("already exists")
	ErrOrderAlreadyCreatedByAnotherUser = errors.New("another user have order")
	ErrOrderAlreadyCreatedByCurrentUser = errors.New("order already created by current user")
	ErrInvalidOrderNumber               = errors.New("invalid order number")
	ErrNotEnoughBalance                 = errors.New("not enough money in balance")

	ErrInvalidInput = errors.New("invalid input")
)
