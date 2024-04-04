package handlers

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/logto-io/go/client"
	"github.com/logto-io/go/core"

	"gitlab.unjx.de/flohoss/godash/internal/env"
)

func NewAuthHandler(env *env.Config, sessionManager *scs.SessionManager) *AuthHandler {
	return &AuthHandler{
		logtoConfig: &client.LogtoConfig{
			Endpoint:  env.OIDCIssuerUrl,
			AppId:     env.OIDCClientId,
			AppSecret: env.OIDCClientSecret,
			Scopes: []string{
				core.UserScopeProfile,
				core.UserScopeEmail,
				core.UserScopeCustomData,
				core.UserScopeIdentities,
				core.UserScopeRoles,
			},
		},
		sessionManager:         sessionManager,
		redirectUri:            env.OIDCRedirectUri,
		postSignOutRedirectUri: env.OIDCPostSignOutRedirectUri,
	}
}

type AuthHandler struct {
	logtoConfig            *client.LogtoConfig
	sessionManager         *scs.SessionManager
	redirectUri            string
	postSignOutRedirectUri string
}

func (ah *AuthHandler) authRequired(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ah.sessionManager == nil {
			handler.ServeHTTP(w, r)
			return
		}
		logtoClient := client.NewLogtoClient(
			ah.logtoConfig,
			&SessionStorage{
				sessionManager: ah.sessionManager,
				write:          w,
				request:        r,
			},
		)
		if !logtoClient.IsAuthenticated() {
			http.Redirect(w, r, "/sign-in", http.StatusTemporaryRedirect)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func (ah *AuthHandler) signInHandler(w http.ResponseWriter, r *http.Request) {
	logtoClient := client.NewLogtoClient(
		ah.logtoConfig,
		&SessionStorage{
			sessionManager: ah.sessionManager,
			write:          w,
			request:        r,
		},
	)
	signInUri, err := logtoClient.SignIn(ah.redirectUri)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, signInUri, http.StatusTemporaryRedirect)
}

func (ah *AuthHandler) signInCallbackHandler(w http.ResponseWriter, r *http.Request) {
	logtoClient := client.NewLogtoClient(
		ah.logtoConfig,
		&SessionStorage{
			sessionManager: ah.sessionManager,
			write:          w,
			request:        r,
		},
	)
	err := logtoClient.HandleSignInCallback(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (ah *AuthHandler) signOutHandler(w http.ResponseWriter, r *http.Request) {
	logtoClient := client.NewLogtoClient(
		ah.logtoConfig,
		&SessionStorage{
			sessionManager: ah.sessionManager,
			write:          w,
			request:        r,
		},
	)
	signOutUri, err := logtoClient.SignOut(ah.postSignOutRedirectUri)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, signOutUri, http.StatusTemporaryRedirect)
}
