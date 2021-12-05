package identity

import (
	"context"
	"github.com/damejeras/auth/api"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"net/url"
)

type service struct {
	challengeRepository    ChallengeRepository
	verificationRepository VerificationRepository
}

func NewService(challengeRepository ChallengeRepository, verificationRepository VerificationRepository) api.IdentityService {
	return &service{
		challengeRepository:    challengeRepository,
		verificationRepository: verificationRepository,
	}
}

func (s *service) Authenticate(ctx context.Context, request api.AuthenticateRequest) (*api.AuthenticateResponse, error) {
	challenge, err := s.challengeRepository.RetrievePendingByID(ctx, request.ChallengeID)
	if err != nil {
		return nil, errors.Wrap(err, "find challenge")
	}

	if challenge == nil {
		return nil, errors.New("challenge doesn't exist")
	}

	originalURL, err := url.Parse(challenge.OriginURL)
	if err != nil {
		return nil, errors.Wrap(err, "parse original uri")
	}

	verification := Verification{
		ChallengeID:   challenge.ID,
		LoginVerifier: ksuid.New().String(),
		RequestID:     challenge.RequestID,
		Data: Data{
			ClientID:  originalURL.Query().Get("client_id"),
			SubjectID: request.SubjectID,
			OriginURL: challenge.OriginURL,
		},
	}

	if err := s.verificationRepository.Store(ctx, &verification); err != nil {
		return nil, errors.Wrap(err, "store verification")
	}

	requestURI, err := url.Parse(challenge.OriginURL)
	if err != nil {
		return nil, errors.Wrap(err, "parse request url")
	}

	urlValues, err := url.ParseQuery(requestURI.RawQuery)
	if err != nil {
		return nil, errors.Wrap(err, "parse url values")
	}

	urlValues.Add(paramLoginVerifier, verification.ChallengeID)

	requestURI.RawQuery = urlValues.Encode()

	return &api.AuthenticateResponse{
		RedirectURL: requestURI.String(),
	}, nil
}
