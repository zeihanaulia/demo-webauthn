package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/duo-labs/webauthn.io/models"
	"github.com/duo-labs/webauthn.io/session"
)

// LoginRequired sets a context variable with the user loaded from the user ID
// stored in the session cookie
func (ws *Server) LoginRequired(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := ws.store.Get(r, session.WebauthnSession)
		// Load the user from the database and store it in the request context
		if id, ok := session.Values["user_id"]; ok {
			fmt.Println("LoginRequired id, ok := session.Values[user_id]")
			u, err := models.GetUser(id.(uint))
			if err != nil {
				fmt.Println("LoginRequired id, ok := session.Values[user_id] if err != nil")
				r = r.WithContext(context.WithValue(r.Context(), "user", nil))
			} else {
				fmt.Println("id, ok := session.Values[user_id] else")
				r = r.WithContext(context.WithValue(r.Context(), "user", u))
			}
		} else {
			fmt.Println("LoginRequired else")
			r = r.WithContext(context.WithValue(r.Context(), "user", nil))
		}

		// If we have a valid user, allow access to the handler. Otherwise,
		// redirect to the main login page.
		if u := r.Context().Value("user"); u != nil {
			next.ServeHTTP(w, r)
			return
		}
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	})
}
