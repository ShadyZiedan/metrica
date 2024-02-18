package utils

import "time"

func RetryWithBackoff(maxRetries int, isRetryable func(error) bool, callback func() error) error {
	var err error
	for i := 1; i <= maxRetries; i++ {
		err = callback()
		if err == nil {
			return nil
		}
		if !isRetryable(err) {
			return err
		}
		time.Sleep(time.Duration(2*i-1) * time.Second)
	}
	return err
}
