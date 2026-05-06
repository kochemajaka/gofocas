package focas

import (
	"context"
	"time"
)

// reconnectDo executes fn, and if it returns a transient error, re-dials and
// retries according to the client's ReconnectPolicy.
func (c *Client) reconnectDo(ctx context.Context, fn func() error) error {
	err := fn()
	if err == nil {
		return nil
	}

	p := c.cfg.reconnect
	if !p.Enabled || !IsTransient(err) {
		return err
	}

	wait := p.InitialWait
	if wait == 0 {
		wait = 200 * time.Millisecond
	}
	maxWait := p.MaxWait
	if maxWait == 0 {
		maxWait = 5 * time.Second
	}
	mult := p.Multiplier
	if mult <= 0 {
		mult = 2.0
	}

	for attempt := 1; ; attempt++ {
		if p.MaxAttempts > 0 && attempt > p.MaxAttempts {
			return err
		}

		c.cfg.logger.Warn("focas: transient error, reconnecting",
			"attempt", attempt,
			"wait", wait,
			"err", err,
		)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}

		if dialErr := c.redial(ctx); dialErr != nil {
			c.cfg.logger.Warn("focas: redial failed", "err", dialErr)
			wait = clampDuration(time.Duration(float64(wait)*mult), maxWait)
			continue
		}

		err = fn()
		if err == nil {
			return nil
		}
		if !IsTransient(err) {
			return err
		}

		wait = clampDuration(time.Duration(float64(wait)*mult), maxWait)
	}
}

func clampDuration(d, max time.Duration) time.Duration {
	if d > max {
		return max
	}
	return d
}
