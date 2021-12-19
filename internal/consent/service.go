package consent

import (
	"context"
	"github.com/damejeras/auth/api"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"net/url"
)

type consentService struct {
	consentRepository          Repository
	consentChallengeRepository ChallengeRepository
}

func NewService(consentRepository Repository, consentChallengeRepository ChallengeRepository) api.ConsentService {
	return &consentService{
		consentRepository:          consentRepository,
		consentChallengeRepository: consentChallengeRepository,
	}
}

func (c *consentService) GrantConsent(ctx context.Context, request api.GrantConsentRequest) (*api.GrantConsentResponse, error) {
	challenge, err := c.consentChallengeRepository.FindByID(ctx, request.ChallengeID)
	if err != nil {
		return nil, errors.Wrap(err, "find consent challenge")
	}

	if challenge == nil {
		return nil, errors.Errorf("invalid consent challenge")
	}

	challenge.GrantedScopes = BuildScopes(request.Scope)

	consent, err := c.consentRepository.FindByClientAndSubject(ctx, challenge.ClientID, challenge.SubjectID)
	if err != nil {
		return nil, errors.Wrap(err, "find consent by client and subject")
	}

	if consent != nil {
		consent.Scopes = consent.Scopes.Merge(challenge.GrantedScopes)
		if err := c.consentRepository.UpdateWithScopes(ctx, consent); err != nil {
			return nil, errors.Wrap(err, "update consent with scopes")
		}
	} else {
		consent = &Consent{
			ID:        ksuid.New().String(),
			ClientID:  challenge.ClientID,
			SubjectID: challenge.SubjectID,
			Scopes:    challenge.GrantedScopes,
		}

		if err := c.consentRepository.Store(ctx, consent); err != nil {
			return nil, errors.Wrap(err, "store consent")
		}
	}

	if err := c.consentChallengeRepository.UpdateWithGrantedScopes(ctx, challenge); err != nil {
		return nil, errors.Wrap(err, "update consent challenge granted scopes")
	}

	requestURL, err := url.Parse(challenge.Footprint.RequestURL)
	if err != nil {
		return nil, errors.Wrap(err, "parse request url")
	}

	urlValues, err := url.ParseQuery(requestURL.RawQuery)
	if err != nil {
		return nil, errors.Wrap(err, "parse url values")
	}

	urlValues.Add("consent_verifier", challenge.Verifier)

	requestURL.RawQuery = urlValues.Encode()

	return &api.GrantConsentResponse{
		RedirectURL: requestURL.String(),
	}, nil
}

func (c *consentService) ShowConsentChallenge(ctx context.Context, request api.ShowConsentChallengeRequest) (*api.ShowConsentChallengeResponse, error) {
	challenge, err := c.consentChallengeRepository.FindByID(ctx, request.ConsentChallenge)
	if err != nil {
		return nil, errors.Wrap(err, "find consent challenge")
	}

	return &api.ShowConsentChallengeResponse{
		ClientID:       challenge.ID,
		SubjectID:      challenge.SubjectID,
		RequestedScope: challenge.RequestedScopes.ToSlice(),
	}, nil
}
