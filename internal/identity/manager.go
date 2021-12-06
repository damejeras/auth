package identity

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/damejeras/auth/internal/consent"
	"github.com/segmentio/ksuid"

	"github.com/damejeras/auth/internal/application"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/server"
)

const (
	paramLoginVerifier   = "login_verifier"
	paramConsentVerifier = "consent_verifier"

	paramLoginChallenge   = "challenge"
	paramConsentChallenge = "consent_challenge"
)

type Manager struct {
	identityProviderURL        string
	consentProviderURL         string
	challengeRepository        ChallengeRepository
	verificationRepository     VerificationRepository
	consentChallengeRepository consent.ChallengeRepository
	consentGrantRepository     consent.GrantRepository
}

func NewManager(
	challengeRepository ChallengeRepository,
	verificationRepository VerificationRepository,
	consentChallengeRepository consent.ChallengeRepository,
	consentGrantRepository consent.GrantRepository,
) *Manager {
	return &Manager{
		identityProviderURL:        "http://localhost:8888/auth",
		consentProviderURL:         "http://localhost:8888/consent",
		challengeRepository:        challengeRepository,
		verificationRepository:     verificationRepository,
		consentChallengeRepository: consentChallengeRepository,
		consentGrantRepository:     consentGrantRepository,
	}
}

func (m *Manager) UserAuthorizationHandler() server.UserAuthorizationHandler {
	return func(w http.ResponseWriter, r *http.Request) (userID string, err error) {
		loginVerifier := r.URL.Query().Get(paramLoginVerifier)
		if loginVerifier != "" {
			// TODO: check identity provider host from configuration
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

			originalURL, err := url.Parse(verification.Data.OriginURL)
			if err != nil {
				// TODO: log error
				return "", errors.ErrServerError
			}

			previousRequestID := application.GetRequestID(r)
			if previousRequestID != verification.RequestID {
				return "", errors.ErrAccessDenied
			}

			originalValues := originalURL.Query()
			currentValues := r.URL.Query()

			if len(originalValues)+1 != len(currentValues) {
				return "", errors.ErrAccessDenied
			}

			for k := range currentValues {
				if k == paramLoginVerifier {
					continue
				}

				if currentValues.Get(k) != originalValues.Get(k) {
					return "", errors.ErrAccessDenied
				}
			}

			// TODO: implement consent and OIDC

			allScopes := strings.Split(originalValues.Get("scope"), " ")
			requestedScopes := make(map[string]struct{})
			for i := range allScopes {
				switch allScopes[i] {
				case "openid", "offline":
					continue
				default:
					requestedScopes[allScopes[i]] = struct{}{}
				}
			}

			grant, err := m.consentGrantRepository.FindByClientAndSubject(
				r.Context(),
				verification.Data.ClientID,
				verification.Data.SubjectID,
			)
			if err != nil {
				// TODO: log error
				return "", errors.ErrServerError
			}

			if grant == nil {
				consentChallenge := &consent.Challenge{
					ID:        ksuid.New().String(),
					ClientID:  verification.Data.ClientID,
					SubjectID: verification.Data.SubjectID,
					RequestID: r.Context().Value(application.ContextRequestID).(string),
					OriginURL: verification.Data.OriginURL,
					Data: consent.ChallengeData{
						Scope: allScopes,
					},
				}

				if err := m.consentChallengeRepository.Store(r.Context(), consentChallenge); err != nil {
					// TODO: log error
					return "", errors.ErrServerError
				}

				queryValues := make(url.Values)
				queryValues.Add(paramConsentChallenge, consentChallenge.ID)

				consentProviderURL, err := url.Parse(m.consentProviderURL)
				if err != nil {
					// TODO: log error
					return "", errors.ErrServerError
				}

				consentProviderURL.RawQuery = queryValues.Encode()

				w.Header().Add("Location", consentProviderURL.String())
				w.WriteHeader(http.StatusFound)

				return "", nil
			}

			consentScopes := make(map[string]struct{})
			for i := range grant.Scope {
				consentScopes[grant.Scope[i]] = struct{}{}
			}

			for k := range requestedScopes {
				if _, ok := consentScopes[k]; !ok {
					consentChallenge := &consent.Challenge{
						ID:        ksuid.New().String(),
						ClientID:  verification.Data.ClientID,
						SubjectID: verification.Data.SubjectID,
						RequestID: r.Context().Value(application.ContextRequestID).(string),
						OriginURL: verification.Data.OriginURL,
						Data: consent.ChallengeData{
							Scope: allScopes,
						},
					}

					if err := m.consentChallengeRepository.Store(r.Context(), consentChallenge); err != nil {
						// TODO: log error
						return "", errors.ErrServerError
					}

					consentProviderURL, err := url.Parse(m.consentProviderURL)
					if err != nil {
						// TODO: log error
						return "", errors.ErrServerError
					}
					queryValues := make(url.Values)
					queryValues.Add(paramConsentChallenge, consentChallenge.ID)

					consentProviderURL.RawQuery = queryValues.Encode()

					w.Header().Add("Location", consentProviderURL.String())
					w.WriteHeader(http.StatusFound)

					return "", nil
				}
			}

			return verification.Data.SubjectID, nil
		}

		consentVerifier := r.URL.Query().Get(paramConsentVerifier)
		if consentVerifier != "" {
			// TODO: check identity provider host from configuration
			if !strings.HasPrefix(r.Header.Get("Referer"), "http://localhost:8888/") {
				return "", errors.ErrAccessDenied
			}

			grant, err := m.consentGrantRepository.FindByID(r.Context(), consentVerifier)
			if err != nil {
				// TODO: log error
				return "", errors.ErrServerError
			}

			if grant == nil {
				return "", errors.ErrAccessDenied
			}

			originalURL, err := url.Parse(grant.OriginURL)
			if err != nil {
				// TODO: log error
				return "", errors.ErrServerError
			}

			previousRequestID := application.GetRequestID(r)
			if previousRequestID != grant.RequestID {
				return "", errors.ErrAccessDenied
			}

			originalValues := originalURL.Query()
			currentValues := r.URL.Query()

			if len(originalValues)+1 != len(currentValues) {
				return "", errors.ErrAccessDenied
			}

			for k := range currentValues {
				if k == paramConsentVerifier {
					continue
				}

				if currentValues.Get(k) != originalValues.Get(k) {
					return "", errors.ErrAccessDenied
				}
			}

			return grant.SubjectID, nil
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
