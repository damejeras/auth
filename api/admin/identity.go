package admin

type IdentityService interface {
	Verify(VerifyRequest) VerifyResponse
}

type VerifyRequest struct {
	ChallengeID string
	SubjectID   string
}

type VerifyResponse struct {
	RedirectURL string
}
