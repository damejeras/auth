package consent

import (
	"context"
	"github.com/damejeras/auth/internal/integrity"
	"time"
)

type Challenge struct {
	ID              string
	Verifier        string
	ClientID        string
	SubjectID       string
	RequestedScopes Scopes
	MissingScopes   Scopes
	GrantedScopes   Scopes
	Footprint       *integrity.Footprint
	Used            bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Scopes map[string]struct{}

func BuildScopes(scopes []string) Scopes {
	result := make(Scopes)

	for i := range scopes {
		result[scopes[i]] = struct{}{}
	}

	return Scopes(result)
}

func (s Scopes) HasAll(scopes Scopes) bool {
	for v := range scopes {
		if _, ok := s[v]; !ok {
			return false
		}
	}

	return true
}

func (s Scopes) ToSlice() []string {
	result := make([]string, len(s))

	i := 0
	for k := range s {
		result[i] = k
		i++
	}

	return result
}

func (s Scopes) Diff(scopes Scopes) Scopes {
	result := make(map[string]struct{})
	for v := range s {
		if _, ok := scopes[v]; !ok {
			result[v] = struct{}{}
		}
	}

	return result
}

func (s Scopes) Merge(scopes Scopes) Scopes {
	result := make(Scopes)

	for k := range s {
		result[k] = struct{}{}
	}

	for k := range scopes {
		if _, ok := result[k]; !ok {
			result[k] = struct{}{}
		}
	}

	return result
}

type ChallengeRepository interface {
	Store(context.Context, *Challenge) error
	UpdateWithGrantedScopes(context.Context, *Challenge) error
	FindByID(context.Context, string) (*Challenge, error)
	FindByVerifier(context.Context, string) (*Challenge, error)
	Delete(context.Context, *Challenge) error
}
