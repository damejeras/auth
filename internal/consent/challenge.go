package consent

import (
	"context"
	"github.com/damejeras/auth/internal/application"
	"github.com/segmentio/ksuid"
	"net/http"
)

type Challenge struct {
	ID, ClientID, SubjectID, RequestID, OriginURL string
	Data                                          ChallengeData
}

type ChallengeData struct {
	Scope []string
}

func BuildChallenge(r *http.Request, subject string, scope []string) *Challenge {
	return &Challenge{
		ID:        ksuid.New().String(),
		ClientID:  r.URL.Query().Get("client_id"),
		SubjectID: subject,
		RequestID: application.GetRequestID(r),
		OriginURL: r.URL.String(),
		Data: ChallengeData{
			Scope: scope,
		},
	}
}

type ChallengeRepository interface {
	Store(ctx context.Context, consent *Challenge) error
	FindByID(ctx context.Context, id string) (*Challenge, error)
}
