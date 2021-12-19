package consent

import (
	"context"
	"time"
)

type Consent struct {
	ID        string
	ClientID  string
	SubjectID string
	Scopes    Scopes

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Repository interface {
	Store(ctx context.Context, consent *Consent) error
	UpdateWithScopes(ctx context.Context, consent *Consent) error
	FindByClientAndSubject(ctx context.Context, clientID, subjectID string) (*Consent, error)
}
