package integrity

import (
	"context"
	"github.com/segmentio/ksuid"
	"net/http"
)

type contextKey int

const (
	ContextRequestID contextKey = iota

	cookieRequestID = "r"
)

func ContextMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := ksuid.New().String()

		http.SetCookie(w, &http.Cookie{
			Name:     cookieRequestID,
			Value:    requestID,
			Secure:   false, // TODO: use true in production
			HttpOnly: true,
		})

		next(w, r.Clone(context.WithValue(r.Context(), ContextRequestID, requestID)))
	}
}

func GetRequestID(r *http.Request) string {
	cookie, err := r.Cookie(cookieRequestID)
	if err != nil {
		return ""
	}

	return cookie.Value
}
