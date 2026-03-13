package agent

import "context"

// Runner orchestrates the quiz or kata session lifecycle.
type Runner interface {
	Run(ctx context.Context) error
}
