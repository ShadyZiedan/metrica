// Package retry provides utility functions for executing functions with exponential backoff and retry logic.
package retry

import (
	"context"
	"time"
)

// WithBackoff is a utility function that executes a callback function with exponential backoff and retry logic.
// It takes a context, maximum number of retries, a function to determine if an error is retryable, and a callback function.
// The function will execute the callback function and retry it if an error is returned and the error is retryable.
// The backoff time between retries increases exponentially with each retry attempt.
//
// Parameters:
// ctx (context.Context): The context for the operation. If the context is canceled or timed out, the function will return the context's error.
// maxRetries (int): The maximum number of retry attempts. If maxRetries is less than or equal to 0, the function will not retry.
// isRetryable (func(error) bool): A function that determines if an error is retryable. If the function returns false, the error will not be retried.
// callback (func() error): The callback function to be executed.
//
// error: If the callback function returns an error and it is retryable, the function will return the error after all retry attempts.
// If the context is canceled or timed out, the function will return the context's error.
// If the callback function returns nil or no error, the function will return nil.
func WithBackoff(ctx context.Context, maxRetries int, isRetryable func(error) bool, callback func() error) error {
	var err error
	err = callback()
	if err == nil {
		return nil
	}
	if maxRetries <= 0 || (maxRetries > 0 && !isRetryable(err)) {
		return err
	}
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
