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
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/thanhpk/randstr"
	"golang.org/x/oauth2"

	"gitlab.unjx.de/flohoss/godash/internal/env"
	"gitlab.unjx.de/flohoss/godash/services"
)

func setCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}
	http.SetCookie(w, c)
}

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

func (ah *AuthHandler) saveTokenToSession(r *http.Request, oauth2Token *oauth2.Token) {
	ah.SessionManager.Put(r.Context(), "access_token", oauth2Token.AccessToken)
	ah.SessionManager.Put(r.Context(), "refresh_token", oauth2Token.RefreshToken)
	ah.SessionManager.Put(r.Context(), "token_type", oauth2Token.TokenType)
	ah.SessionManager.Put(r.Context(), "expiry", oauth2Token.Expiry.Unix())
}

func (ah *AuthHandler) loadTokenFromSession(r *http.Request) *oauth2.Token {
	ex := ah.SessionManager.GetInt64(r.Context(), "expiry")
	return &oauth2.Token{
		AccessToken:  ah.SessionManager.GetString(r.Context(), "access_token"),
		RefreshToken: ah.SessionManager.GetString(r.Context(), "refresh_token"),
		TokenType:    ah.SessionManager.GetString(r.Context(), "token_type"),
		Expiry:       time.Unix(ex, 0),
	}
}

func NewAuthHandler(env *env.Config) *AuthHandler {
	ctx := context.Background()

	oidcProvider, err := oidc.NewProvider(ctx, env.OIDCIssuer)
	if err != nil {
		slog.Error("Failed to get oidc provider", "err", err.Error())
		os.Exit(1)
	}

	oauth2Config := &oauth2.Config{
		ClientID:     env.OIDCClientID,
		ClientSecret: env.OIDCClientSecret,
		Endpoint:     oidcProvider.Endpoint(),
		RedirectURL:  env.OIDCRedirectURI,
		Scopes:       env.OIDCScopes,
	}

	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		slog.Error("Error generating code verifier", "err", err.Error())
		os.Exit(1)
	}
	codeChallenge := generateCodeChallenge(codeVerifier)
	authCodeOptions := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("redirect_uri", env.OIDCRedirectURI),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * 31 * time.Hour

	return &AuthHandler{
		ctx:             ctx,
		oidcProvider:    oidcProvider,
		oauth2Config:    oauth2Config,
		authCodeOptions: authCodeOptions,
		SessionManager:  sessionManager,
	}
}

type AuthHandler struct {
	ctx             context.Context
	oidcProvider    *oidc.Provider
	oauth2Config    *oauth2.Config
	authCodeOptions []oauth2.AuthCodeOption
	SessionManager  *scs.SessionManager
}

func (ah *AuthHandler) handleCallback(w http.ResponseWriter, r *http.Request) {
	state, err := r.Cookie("state")
	if err != nil {
		http.Error(w, "state not found", http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("state") != state.Value {
		http.Error(w, "state did not match", http.StatusBadRequest)
		return
	}

	oauth2Token, err := ah.oauth2Config.Exchange(ah.ctx, r.URL.Query().Get("code"), ah.authCodeOptions...)
	if err != nil {
		http.Error(w, "failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ah.saveTokenToSession(r, oauth2Token)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (ah *AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	ah.SessionManager.Clear(r.Context())
	http.Redirect(w, r, "/", http.StatusFound)
}

func (ah *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	state := randstr.String(16)
	setCallbackCookie(w, r, "state", state)
	http.Redirect(w, r, ah.oauth2Config.AuthCodeURL(state, ah.authCodeOptions...), http.StatusFound)
}

func (ah *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		exists := ah.SessionManager.Exists(r.Context(), "access_token")
		if !exists {
			ah.handleLogin(w, r)
			return
		}

		token := ah.loadTokenFromSession(r)
		ah.oauth2Config.Client(ah.ctx, token)

		tokenInfo, err := ah.oidcProvider.Verifier(&oidc.Config{ClientID: ah.oauth2Config.ClientID}).Verify(ah.ctx, token.AccessToken)
		if err != nil {
			ah.handleLogin(w, r)
			return
		}

		ah.saveTokenToSession(r, token)

		var userClaims services.User
		tokenInfo.Claims(&userClaims)
		w.Header().Set("X-User-Name", userClaims.Name)
		w.Header().Set("X-User-Email", userClaims.Email)

		next.ServeHTTP(w, r)
	})
}
