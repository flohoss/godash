package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/sessions"
	"github.com/thanhpk/randstr"
	"golang.org/x/oauth2"

	"gitlab.unjx.de/flohoss/godash/internal/env"
	"gitlab.unjx.de/flohoss/godash/services"
)

type contextKey string

const (
	NameKey     contextKey = "name"
	EmailKey    contextKey = "email"
	GravatarKey contextKey = "gravatar"

	StoreSessionKey = "godash_session"
)

func generateCodeVerifier() (string, error) {
	verifierLength := 64
	verifier := make([]byte, verifierLength)

	_, err := rand.Read(verifier)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(verifier), nil
}

func generateCodeChallenge(verifier string) string {
	hash := sha256.New()
	_, _ = io.WriteString(hash, verifier)
	sha := hash.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(sha)
}

func NewAuthHandler(env *env.Config, store *sessions.CookieStore) *AuthHandler {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, env.AuthIssuer)
	if err != nil {
		slog.Error("Failed to get oidc provider", "err", err.Error())
		os.Exit(1)
	}

	config := &oauth2.Config{
		ClientID:     env.AuthClientID,
		ClientSecret: env.AuthClientSecret,
		RedirectURL:  env.PublicUrl + "/callback",
		Scopes:       env.AuthScopes,
		Endpoint:     provider.Endpoint(),
	}

	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		slog.Error("Error generating code verifier", "err", err.Error())
		os.Exit(1)
	}
	codeChallenge := generateCodeChallenge(codeVerifier)
	authCodeOptions := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	}

	return &AuthHandler{
		provider:        provider,
		config:          config,
		authCodeOptions: authCodeOptions,
		store:           store,
	}
}

type AuthHandler struct {
	provider        *oidc.Provider
	config          *oauth2.Config
	authCodeOptions []oauth2.AuthCodeOption
	store           *sessions.CookieStore
}

func (ah *AuthHandler) handleCallback(w http.ResponseWriter, r *http.Request) {
	session, _ := ah.store.Get(r, StoreSessionKey)
	state, ok := session.Values["state"].(string)
	if !ok || state == "" {
		http.Error(w, "state not found", http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("state") != state {
		http.Error(w, "state did not match", http.StatusBadRequest)
		return
	}

	oauth2Token, err := ah.config.Exchange(r.Context(), r.URL.Query().Get("code"), ah.authCodeOptions...)
	if err != nil {
		http.Error(w, "failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo, err := ah.provider.UserInfo(r.Context(), oauth2.StaticTokenSource(oauth2Token))
	if err != nil {
		http.Error(w, "failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	user := &services.User{}
	userInfo.Claims(user)

	session.Values[string(NameKey)] = user.Name
	session.Values[string(EmailKey)] = user.Email
	session.Values[string(GravatarKey)] = services.NewGravatarFromEmail(user.Email).GetURL()
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (ah *AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := ah.store.Get(r, StoreSessionKey)
	session.Values = make(map[interface{}]interface{})
	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (ah *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := ah.store.Get(r, StoreSessionKey)
		name, ok := session.Values[string(NameKey)].(string)
		if !ok || name == "" {
			state := randstr.String(16)
			session.Values["state"] = state
			err := session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, ah.config.AuthCodeURL(state, ah.authCodeOptions...), http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
