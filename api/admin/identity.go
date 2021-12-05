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
	GrantConsent(GrantConsentRequest) GrantConsentResponse
}

type GrantConsentRequest struct {
	ChallengeID string
	Scope       []string
}

type GrantConsentResponse struct {
	RedirectURL string
}
