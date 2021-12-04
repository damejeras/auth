// Code generated by oto; DO NOT EDIT.

package api

import (
	"context"
	"net/http"

	"github.com/pacedotdev/oto/otohttp"
)

type IdentityService interface {
	Verify(context.Context, VerifyRequest) (*VerifyResponse, error)
}

type identityServiceServer struct {
	server          *otohttp.Server
	identityService IdentityService
}

// Register adds the IdentityService to the otohttp.Server.
func RegisterIdentityService(server *otohttp.Server, identityService IdentityService) {
	handler := &identityServiceServer{
		server:          server,
		identityService: identityService,
	}
	server.Register("IdentityService", "Verify", handler.handleVerify)
}

func (s *identityServiceServer) handleVerify(w http.ResponseWriter, r *http.Request) {
	var request VerifyRequest
	if err := otohttp.Decode(r, &request); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	response, err := s.identityService.Verify(r.Context(), request)
	if err != nil {
		s.server.OnErr(w, r, err)
		return
	}
	if err := otohttp.Encode(w, r, http.StatusOK, response); err != nil {
		s.server.OnErr(w, r, err)
		return
	}
}

type VerifyRequest struct {
	ChallengeID string `json:"challengeID"`
	SubjectID   string `json:"subjectID"`
}

type VerifyResponse struct {
	RedirectURL string `json:"redirectURL"`
	// Error is string explaining what went wrong. Empty if everything was fine.
	Error string `json:"error,omitempty"`
}
