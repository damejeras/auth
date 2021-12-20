package app

import (
	"context"
	"github.com/segmentio/ksuid"
	"net/http"
)

type contextKey int

const (
	requestIDContextKey contextKey = iota

	cookieRequestID = "r"
)

func RequestMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := ksuid.New().String()

		http.SetCookie(w, &http.Cookie{
			Name:     cookieRequestID,
			Value:    requestID,
			Secure:   false, // TODO: use true in production
			HttpOnly: true,
		})

		next(w, r.Clone(context.WithValue(r.Context(), requestIDContextKey, requestID)))
	}
}

func GetCurrentRequestID(r *http.Request) string {
	return r.Context().Value(requestIDContextKey).(string)
}

func GetPreviousRequestID(r *http.Request) string {
	cookie, err := r.Cookie(cookieRequestID)
	if err != nil {
		return ""
	}

	return cookie.Value
}
