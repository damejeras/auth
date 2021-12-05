package consent

import (
	"context"
	"github.com/damejeras/auth/api"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"net/url"
)

const (
	paramConsentChallenge = "consent_challenge"
)

type service struct {
	challengeRepository ChallengeRepository
	grantRepository     GrantRepository
}

func NewService(challengeRepository ChallengeRepository, grantRepository GrantRepository) api.ConsentService {
	return &service{
		challengeRepository: challengeRepository,
		grantRepository:     grantRepository,
	}
}

func (s *service) GrantConsent(ctx context.Context, request api.GrantConsentRequest) (*api.GrantConsentResponse, error) {
	challenge, err := s.challengeRepository.FindByID(ctx, request.ChallengeID)
	if err != nil {
		return nil, errors.Wrap(err, "find challenge")
	}

	grant := Grant{
		ID:          ksuid.New().String(),
		ClientID:    challenge.ID,
		SubjectID:   challenge.SubjectID,
		ChallengeID: challenge.ID,
		OriginURL:   challenge.OriginURL,
		Scope:       request.Scope,
	}

	err = s.grantRepository.Store(ctx, &grant)
	if err != nil {
		return nil, errors.Wrap(err, "store grant")
	}

	requestURI, err := url.Parse(challenge.OriginURL)
	if err != nil {
		return nil, errors.Wrap(err, "parse request url")
	}

	urlValues, err := url.ParseQuery(requestURI.RawQuery)
	if err != nil {
		return nil, errors.Wrap(err, "parse url values")
	}

	urlValues.Add(paramConsentChallenge, grant.ID)

	requestURI.RawQuery = urlValues.Encode()

	return &api.GrantConsentResponse{
		RedirectURL: requestURI.String(),
	}, nil
}
