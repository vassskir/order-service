package retry

import (
	"context"
	"fmt"
	"time"

	"L0/internal/logger"
)

type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

var DefaultConfig = RetryConfig{
	MaxAttempts:  3,
	InitialDelay: 1 * time.Second,
	MaxDelay:     30 * time.Second,
	Multiplier:   2.0,
}

func Retry(ctx context.Context, operationName string, config RetryConfig, operation func() error) error {
	var lastErr error
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			if attempt > 1 {
				logger.Info("Operation succeeded after retry", map[string]interface{}{
					"operation": operationName,
					"attempt":   attempt,
				})
			}
			return nil
		}

		lastErr = err

		if attempt == config.MaxAttempts {
			break
		}

		delay := calculateDelay(config, attempt)

		logger.Warn("Operation failed, retrying", map[string]interface{}{
			"operation":    operationName,
			"attempt":      attempt,
			"max_attempts": config.MaxAttempts,
			"delay":        delay.String(),
			"error":        err.Error(),
		})

		if !sleepWithContext(ctx, delay) {
			return fmt.Errorf("operation cancelled: %w", lastErr)
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

func calculateDelay(config RetryConfig, attempt int) time.Duration {
	delay := config.InitialDelay
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}
	return delay
}

func sleepWithContext(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}
