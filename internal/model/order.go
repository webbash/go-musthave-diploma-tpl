package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	Number     string          `json:"order"`
	UserID     int64           `json:"userId,omitempty"`
	Status     OrderStatus     `json:"status"`
	Accrual    decimal.Decimal `json:"accrual"`
	UploadedAt time.Time       `json:"uploaded_at,omitempty"`
	UpdatedAt  time.Time       `json:"updated_at,omitempty"`
}
