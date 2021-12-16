package identity

import (
	"context"
	"github.com/damejeras/auth/api"
	"github.com/pkg/errors"
	"net/url"
)

type service struct {
	challengeRepository ChallengeRepository
}

func NewService(challengeRepository ChallengeRepository) api.IdentityService {
	return &service{
		challengeRepository: challengeRepository,
	}
}

func (s *service) Authenticate(ctx context.Context, request api.AuthenticateRequest) (*api.AuthenticateResponse, error) {
	challenge, err := s.challengeRepository.FindByID(ctx, request.ChallengeID)
	if err != nil {
		return nil, errors.Wrap(err, "find challenge")
	}

	if challenge == nil || challenge.Identity.SubjectID != "" {
		return nil, errors.New("invalid challenge")
	}

	challenge.Identity = &Identity{SubjectID: request.SubjectID}

	if err := s.challengeRepository.UpdateWithAuthorization(ctx, challenge); err != nil {
		return nil, errors.Wrap(err, "update challenge")
	}

	requestURL, err := url.Parse(challenge.Footprint.RequestURL)
	if err != nil {
		return nil, errors.Wrap(err, "parse request url")
	}

	urlValues, err := url.ParseQuery(requestURL.RawQuery)
	if err != nil {
		return nil, errors.Wrap(err, "parse url values")
	}

	urlValues.Add(paramLoginVerifier, challenge.Verifier)

	requestURL.RawQuery = urlValues.Encode()

	return &api.AuthenticateResponse{
		RedirectURL: requestURL.String(),
	}, nil
}
