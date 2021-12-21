package identity

import (
	"github.com/damejeras/auth/internal/app"
	"github.com/damejeras/auth/internal/consent"
	"github.com/damejeras/auth/internal/integrity"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/server"
	pkgErrors "github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	paramLoginVerifier = "login_verifier"

	paramLoginChallenge   = "challenge"
	paramConsentChallenge = "consent_challenge"
)

type Manager struct {
	identityProviderURL        string
	consentProviderURL         string
	challengeRepository        ChallengeRepository
	consentChallengeRepository consent.ChallengeRepository
	consentRepository          consent.Repository
	logger                     *zerolog.Logger
}

func NewManager(
	challengeRepository ChallengeRepository,
	consentChallengeRepository consent.ChallengeRepository,
	consentRepository consent.Repository,
	logger *zerolog.Logger,
	cfg *app.Config,
) *Manager {
	return &Manager{
		identityProviderURL:        cfg.IdentityProviderConfig.Address,
		consentProviderURL:         cfg.ConsentProviderConfig.Address,
		challengeRepository:        challengeRepository,
		consentChallengeRepository: consentChallengeRepository,
		consentRepository:          consentRepository,
		logger:                     logger,
	}
}

func (m *Manager) UserAuthorizationHandler() server.UserAuthorizationHandler {
	return func(w http.ResponseWriter, r *http.Request) (userID string, err error) {
		loginVerifier := r.URL.Query().Get(paramLoginVerifier)
		if loginVerifier != "" {
			m.logger.Trace().Msgf("serve login verifier %q", loginVerifier)

			challenge, err := m.challengeRepository.FindByVerifier(r.Context(), loginVerifier)
			if err != nil {
				m.logger.Error().Err(err).Msg("find challenge by login verifier")

				return "", errors.ErrServerError
			}

			if challenge == nil {
				m.logger.Warn().Msgf("login verifier %q not found", loginVerifier)

				return "", errors.ErrInvalidRequest
			}

			if err := challenge.Footprint.Validate(r); err != nil {
				switch err.(type) {
				case integrity.ValidationError:
					// todo: track violation
					return "", errors.ErrAccessDenied
				default:
					m.logger.Error().Err(err).Msg("validate request")

					return "", errors.ErrServerError
				}
			}

			if challenge.Identity == nil {
				// todo: track violation
				return "", errors.ErrAccessDenied
			}

			cs, err := m.consentRepository.FindByClientAndSubject(r.Context(), challenge.ClientID, challenge.Identity.SubjectID)
			if err != nil {
				m.logger.Error().Err(err).Msgf("find client's %q consent for subject %q", challenge.ClientID, challenge.Identity.SubjectID)

				return "", errors.ErrServerError
			}

			requestedScopes := scopeParamToScopes(r.URL.Query().Get("scope"))

			if cs == nil || !cs.Scopes.HasAll(requestedScopes) {
				var consentChallenge *consent.Challenge
				if cs != nil {
					consentChallenge, err = m.createConsentChallenge(r, requestedScopes, requestedScopes.Diff(cs.Scopes), challenge.ClientID, challenge.Identity.SubjectID)
					if err != nil {
						m.logger.Error().Err(err).Msg("create consent challenge")

						return "", errors.ErrServerError
					}
				} else {
					consentChallenge, err = m.createConsentChallenge(r, requestedScopes, requestedScopes, challenge.ClientID, challenge.Identity.SubjectID)
					if err != nil {
						m.logger.Error().Err(err).Msg("create consent challenge")

						return "", errors.ErrServerError
					}
				}

				w.Header().Add("Location", consentChallenge.Footprint.RedirectURL)
				w.WriteHeader(http.StatusFound)

				return "", nil
			}

			if err := m.challengeRepository.Delete(r.Context(), challenge); err != nil {
				m.logger.Error().Err(err).Msgf("delete login challenge %q", challenge)

				return "", errors.ErrServerError
			}

			return challenge.Identity.SubjectID, nil
		}

		consentVerifier := r.URL.Query().Get("consent_verifier")
		if consentVerifier != "" {
			m.logger.Trace().Msgf("serve consent verifier %q", consentVerifier)

			consentChallenge, err := m.consentChallengeRepository.FindByVerifier(r.Context(), consentVerifier)
			if err != nil {
				m.logger.Error().Err(err).Msg("find consent challenge by verifier")

				return "", errors.ErrServerError
			}

			if consentChallenge == nil || consentChallenge.GrantedScopes == nil {
				// todo track violation
				return "", errors.ErrAccessDenied
			}

			if err := consentChallenge.Footprint.Validate(r); err != nil {
				switch err.(type) {
				case integrity.ValidationError:
					// todo: track violation
					return "", errors.ErrAccessDenied
				default:
					m.logger.Error().Err(err).Msg("validate footprint")

					return "", errors.ErrServerError
				}
			}

			consentChallenge.Used = true

			if err := m.consentChallengeRepository.Delete(r.Context(), consentChallenge); err != nil {
				m.logger.Error().Err(err).Msgf("delete consent challenge %q", consentChallenge)

				return "", errors.ErrServerError
			}

			return consentChallenge.SubjectID, nil
		}

		m.logger.Trace().Msg("creating new login challenge")

		challenge, err := m.createLoginChallenge(r)
		if err != nil {
			m.logger.Error().Err(err).Msg("create login challenge")

			return "", errors.ErrServerError
		}

		w.Header().Add("Location", challenge.Footprint.RedirectURL)
		w.WriteHeader(http.StatusFound)

		return "", nil
	}
}

func (m *Manager) createConsentChallenge(r *http.Request, requested, missing consent.Scopes, clientID, subjectID string) (*consent.Challenge, error) {
	challengeID := ksuid.New().String()

	cpURL, err := url.Parse(m.consentProviderURL)
	if err != nil {
		return nil, pkgErrors.Wrap(err, "parse idp url")
	}

	queryValues := make(url.Values)
	queryValues.Add(paramConsentChallenge, challengeID)

	cpURL.RawQuery = queryValues.Encode()

	challenge := consent.Challenge{
		ID:              challengeID,
		Verifier:        ksuid.New().String(),
		ClientID:        clientID,
		SubjectID:       subjectID,
		RequestedScopes: requested,
		MissingScopes:   missing,
		GrantedScopes:   nil,
		Footprint: &integrity.Footprint{
			RequestID:   app.GetCurrentRequestID(r),
			RedirectURL: cpURL.String(),
			// TODO: use r.URL.Scheme ?
			RequestURL: "http" + "://" + r.Host + r.URL.RequestURI(),
		},
		CreatedAt: time.Now(),
	}

	if err := m.consentChallengeRepository.Store(r.Context(), &challenge); err != nil {
		return nil, pkgErrors.Wrap(err, "store challenge")
	}

	return &challenge, nil
}

func (m *Manager) createLoginChallenge(r *http.Request) (*Challenge, error) {
	challengeID := ksuid.New().String()

	idpURL, err := url.Parse(m.identityProviderURL)
	if err != nil {
		return nil, pkgErrors.Wrap(err, "parse idp url")
	}

	queryValues := make(url.Values)
	queryValues.Add(paramLoginChallenge, challengeID)

	idpURL.RawQuery = queryValues.Encode()

	challenge := Challenge{
		ID:       challengeID,
		ClientID: r.URL.Query().Get("client_id"),
		Verifier: ksuid.New().String(),
		Footprint: &integrity.Footprint{
			RequestID:   app.GetCurrentRequestID(r),
			RedirectURL: idpURL.String(),
			// TODO: use r.URL.Scheme ???
			RequestURL: "http" + "://" + r.Host + r.URL.RequestURI(),
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}

	if err := m.challengeRepository.Store(r.Context(), &challenge); err != nil {
		return nil, pkgErrors.Wrap(err, "store challenge")
	}

	return &challenge, nil
}

func stringSliceDifference(i1, i2 []string) []string {
	var first, second []string

	if len(i1) < len(i2) {
		first = i2
		second = i1
	} else {
		first = i1
		second = i2
	}

	vmap := make(map[string]struct{})
	for i := range second {
		vmap[second[i]] = struct{}{}
	}

	result := make([]string, 0)
	for i := range first {
		if _, ok := vmap[first[i]]; !ok {
			result = append(result, first[i])
		}
	}

	return result
}

func scopeParamToScopes(input string) consent.Scopes {
	result := make(map[string]struct{})
	split := strings.Split(input, " ")
	for i := range split {
		result[split[i]] = struct{}{}
	}

	return result
}
