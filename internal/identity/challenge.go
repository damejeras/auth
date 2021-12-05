package identity

import (
	"context"
	"github.com/damejeras/auth/internal/application"
	"github.com/segmentio/ksuid"
	"net/http"
)

type Challenge struct {
	ID        string
	RequestID string
	OriginURL string
}

func BuildChallenge(r *http.Request) *Challenge {
	return &Challenge{
		ID:        ksuid.New().String(),
		RequestID: r.Context().Value(application.ContextRequestID).(string),
		// TODO: use https
		OriginURL: "http://" + r.Host + r.URL.RequestURI(),
	}
}

type ChallengeRepository interface {
	Store(context.Context, *Challenge) error
	RetrievePendingByID(context.Context, string) (*Challenge, error)
}
