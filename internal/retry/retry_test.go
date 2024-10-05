package retry

import (
	"context"
	"errors"
	"fmt"
)

// ExampleWithBackoff demonstrates how to use the WithBackoff function.
func ExampleWithBackoff() {
	// Simulate a callback that will fail a few times before succeeding
	var attempts int
	callback := func() error {
		attempts++
		if attempts < 3 {
			return errors.New("retryable error") // Fail for the first two attempts
		}
		return nil // Succeed on the third attempt
	}

	// Define a function to check if an error is retryable
	isRetryable := func(err error) bool {
		return err.Error() == "retryable error"
	}

	// Create a context that will not timeout
	ctx := context.Background()

	// Call WithBackoff
	err := WithBackoff(ctx, 5, isRetryable, callback)

	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Success after", attempts, "attempts")
	}

	// Output:
	// Success after 3 attempts
}
