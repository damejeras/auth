package login

import (
	"context"
	"time"
)

type Challenge struct {
	ID        string
	Verifier  string
	ExpiresAt time.Time
}

type ChallengeRepository interface {
	Store(context.Context, *Challenge) error
	Retrieve(context.Context, string) (*Challenge, error)
}
