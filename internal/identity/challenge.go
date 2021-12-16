package identity

import (
	"context"
	"github.com/damejeras/auth/internal/integrity"
	"time"
)

type Challenge struct {
	ID        string
	ClientID  string
	Verifier  string
	Identity  *Identity
	Footprint *integrity.Footprint

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Identity struct {
	SubjectID string
}

type ChallengeRepository interface {
	Store(context.Context, *Challenge) error
	UpdateWithAuthorization(context.Context, *Challenge) error
	Delete(context.Context, *Challenge) error
	FindByID(context.Context, string) (*Challenge, error)
	FindByVerifier(context.Context, string) (*Challenge, error)
}
