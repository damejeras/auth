package integrity

import (
	"fmt"
	"github.com/damejeras/auth/internal/app"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"strings"
)

var validationParams = []string{
	"client_id",
	"response_type",
	"redirect_uri",
	"state",
	"scope",
	"code_challenge",
	"code_challenge_method",
	"prompt",
	"ui_locales",
	"nonce",
}

type ValidationError string

func (e ValidationError) Error() string {
	return string(e)
}

type Footprint struct {
	RequestID   string
	RedirectURL string
	RequestURL  string
}

func (f *Footprint) Validate(r *http.Request) error {
	previousRequestID := app.GetPreviousRequestID(r)
	if previousRequestID != f.RequestID {
		return ValidationError("request ID does not match")
	}

	redirectURL, err := url.Parse(f.RedirectURL)
	if err != nil {
		return errors.Wrap(err, "parse redirect url")
	}

	if !strings.HasPrefix(r.Header.Get("Referer"), redirectURL.Scheme+"://"+redirectURL.Host) {
		return ValidationError(fmt.Sprintf("expected referer %q, got %q", f.RedirectURL, r.Header.Get("Referer")))
	}

	footprintURL, err := url.Parse(f.RequestURL)
	if err != nil {
		return errors.Wrap(err, "parse request url")
	}

	footprintValues := footprintURL.Query()
	requestValues := r.URL.Query()

	for i := range validationParams {
		if footprintValues.Get(validationParams[i]) != requestValues.Get(validationParams[i]) {
			return ValidationError(fmt.Sprintf("request parameter %q does not match", validationParams[i]))
		}
	}

	return nil
}
