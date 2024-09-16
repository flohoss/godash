package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"gitlab.unjx.de/flohoss/godash/internal/env"
)

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

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
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	}

	return &AuthHandler{
		ctx:             ctx,
		oidcProvider:    oidcProvider,
		oauth2Config:    oauth2Config,
		authCodeOptions: authCodeOptions,
	}
}

type AuthHandler struct {
	ctx             context.Context
	oidcProvider    *oidc.Provider
	oauth2Config    *oauth2.Config
	authCodeOptions []oauth2.AuthCodeOption
}

func (ah *AuthHandler) handleAuth(w http.ResponseWriter, r *http.Request) {
	state, err := randString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	setCallbackCookie(w, r, "state", state)

	http.Redirect(w, r, ah.oauth2Config.AuthCodeURL(state, ah.authCodeOptions...), http.StatusFound)
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

	oauth2Token, err := ah.oauth2Config.Exchange(ah.ctx, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo, err := ah.oidcProvider.UserInfo(ah.ctx, oauth2.StaticTokenSource(oauth2Token))
	if err != nil {
		http.Error(w, "failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		OAuth2Token *oauth2.Token
		UserInfo    *oidc.UserInfo
	}{oauth2Token, userInfo}
	data, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func (ah *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state, err := r.Cookie("state")
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		oauth2Token, err := ah.oauth2Config.Exchange(ah.ctx, r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userInfo, err := ah.oidcProvider.UserInfo(ah.ctx, oauth2.StaticTokenSource(oauth2Token))
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		fmt.Println(userInfo)
		next.ServeHTTP(w, r)
	})
}
