package model

import "time"

type Withdrawal struct {
	Order       string
	UserID      int64
	Sum         float64
	ProcessedAt time.Time
}
