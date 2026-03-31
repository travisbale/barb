package api

import (
	"context"
	"net/http"

	"github.com/travisbale/barb/internal/phishing"
)

type userContextKey struct{}

// requireAuth wraps a handler with session validation.
func (r *Router) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cookie, err := req.Cookie("session")
		if err != nil {
			r.writeError(w, http.StatusUnauthorized, "authentication required", nil)
			return
		}

		user, err := r.Auth.CurrentUser(cookie.Value)
		if err != nil {
			r.writeError(w, http.StatusUnauthorized, "session expired", nil)
			return
		}

		ctx := context.WithValue(req.Context(), userContextKey{}, user)
		next(w, req.WithContext(ctx))
	}
}

func userFromContext(ctx context.Context) *phishing.User {
	user, _ := ctx.Value(userContextKey{}).(*phishing.User)
	return user
}
