package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type User struct {
	ID           int64
	Login        string
	PasswordHash string
	Balance      decimal.Decimal
	CreatedAt    time.Time
}
