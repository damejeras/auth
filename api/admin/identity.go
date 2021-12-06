package admin

type IdentityService interface {
	Authenticate(AuthenticateRequest) AuthenticateResponse
}

type AuthenticateRequest struct {
	ChallengeID string
	SubjectID   string
}

type AuthenticateResponse struct {
	RedirectURL string
}

type ConsentService interface {
	ShowConsentChallenge(ShowConsentChallengeRequest) ShowConsentChallengeResponse
	GrantConsent(GrantConsentRequest) GrantConsentResponse
}

type ShowConsentChallengeRequest struct {
	ConsentChallenge string
}

type ShowConsentChallengeResponse struct {
	ClientID       string
	SubjectID      string
	RequestedScope []string
}

type GrantConsentRequest struct {
	ChallengeID string
	Scope       []string
}

type GrantConsentResponse struct {
	RedirectURL string
}
