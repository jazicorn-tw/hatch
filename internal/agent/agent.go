package agent

import "context"

// Agent orchestrates the quiz or kata session lifecycle.
type Agent interface {
	Run(ctx context.Context) error
}
