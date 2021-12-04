package identity

import (
	"context"
	"github.com/damejeras/auth/api"
	"github.com/google/uuid"
	"github.com/pkg/errors"
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

func (s *service) Verify(ctx context.Context, request api.VerifyRequest) (*api.VerifyResponse, error) {
	challenge, err := s.challengeRepository.RetrievePendingByID(ctx, request.ChallengeID)
	if err != nil {
		return nil, errors.Wrap(err, "find challenge")
	}

	if challenge == nil {
		return nil, errors.New("challenge doesn't exist")
	}

	verification := Verification{
		ChallengeID:   challenge.ID,
		LoginVerifier: uuid.New().String(),
		RequestID:     challenge.RequestID,
		Data: Data{
			SubjectID: request.SubjectID,
		},
	}

	if err := s.verificationRepository.Store(ctx, &verification); err != nil {
		return nil, errors.Wrap(err, "store verification")
	}

	requestURI, err := url.Parse(challenge.RequestURL)
	if err != nil {
		return nil, errors.Wrap(err, "parse request url")
	}

	urlValues, err := url.ParseQuery(requestURI.RawQuery)
	if err != nil {
		return nil, errors.Wrap(err, "parse url values")
	}

	urlValues.Add(paramLoginVerifier, verification.ChallengeID)

	requestURI.RawQuery = urlValues.Encode()

	return &api.VerifyResponse{
		RedirectURL: requestURI.String(),
	}, nil
}
