package model

import "time"

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	Number     string      `json:"order"`
	UserID     int64       `json:"user_id,omitempty"`
	Status     OrderStatus `json:"status"`
	Accrual    float64     `json:"accrual"`
	UploadedAt time.Time   `json:"uploaded_at,omitempty"`
	UpdatedAt  time.Time   `json:"updated_at,omitempty"`
}
