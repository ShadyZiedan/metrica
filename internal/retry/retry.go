package retry

import (
	"context"
	"time"
)

func WithBackoff(ctx context.Context, maxRetries int, isRetryable func(error) bool, callback func() error) error {
	var err error
	for i := 1; i <= maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * time.Duration(2*i-1)):
			err = callback()
			if err == nil {
				return nil
			}
			if !isRetryable(err) {
				return err
			}
		}
	}
	return err
}
