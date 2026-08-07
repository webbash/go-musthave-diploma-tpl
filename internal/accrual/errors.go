package accrual

import (
	"errors"
	"fmt"
	"time"
)

var ErrOrderNotFound = errors.New("order not found")

type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, retry after %s", e.RetryAfter)
}
