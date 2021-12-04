package identity

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/damejeras/auth/internal/application"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/server"
)

const (
	paramLoginVerifier = "login_verifier"

	paramLoginChallenge = "challenge"
)

type Manager struct {
	identityProviderURL    string
	challengeRepository    ChallengeRepository
	verificationRepository VerificationRepository
}

func NewManager(challengeRepository ChallengeRepository, verificationRepository VerificationRepository) *Manager {
	return &Manager{identityProviderURL: "http://localhost:8888/auth", challengeRepository: challengeRepository, verificationRepository: verificationRepository}
}

func (m *Manager) UserAuthorizationHandler() server.UserAuthorizationHandler {
	return func(w http.ResponseWriter, r *http.Request) (userID string, err error) {
		loginVerifier := r.URL.Query().Get(paramLoginVerifier)
		if loginVerifier != "" {
			// TODO: check identity provider from configuration
			if !strings.HasPrefix(r.Header.Get("Referer"), "http://localhost:8888/") {
				return "", errors.ErrAccessDenied
			}

			verification, err := m.verificationRepository.RetrieveByID(r.Context(), loginVerifier)
			if err != nil {
				// TODO: log error
				return "", errors.ErrServerError
			}

			if verification == nil {
				return "", errors.ErrAccessDenied
			}

			previousRequestID := application.GetRequestID(r)
			if previousRequestID != verification.RequestID {
				return "", errors.ErrAccessDenied
			}

			return verification.Data.SubjectID, nil
		}

		challenge := BuildChallenge(r)
		if err := m.challengeRepository.Store(r.Context(), challenge); err != nil {
			// TODO: log error
			return "", errors.ErrServerError
		}

		queryValues := make(url.Values)
		queryValues.Add(paramLoginChallenge, challenge.ID)

		idpURL, err := url.Parse(m.identityProviderURL)
		if err != nil {
			// TODO: log error
			return "", errors.ErrServerError
		}

		idpURL.RawQuery = queryValues.Encode()

		w.Header().Add("Location", idpURL.String())
		w.WriteHeader(http.StatusFound)

		return "", nil
	}
}
