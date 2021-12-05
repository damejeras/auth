package consent

import "context"

type Grant struct {
	ID          string
	ClientID    string
	SubjectID   string
	ChallengeID string
	OriginURL   string
	RequestID   string
	Scope       []string
}

type GrantRepository interface {
	Store(ctx context.Context, grant *Grant) error
	FindByID(ctx context.Context, id string) (*Grant, error)
	FindByClientAndSubject(ctx context.Context, client, subject string) (*Grant, error)
}
