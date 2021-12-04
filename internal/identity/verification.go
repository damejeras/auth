package identity

import "context"

type Verification struct {
	ChallengeID, LoginVerifier, RequestID string

	Data Data
}

type Data struct {
	SubjectID string
}

type VerificationRepository interface {
	Store(context.Context, *Verification) error
	RetrieveByID(context.Context, string) (*Verification, error)
}
