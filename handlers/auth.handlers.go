package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/zitadel/oidc/v3/pkg/client/rp"
	httphelper "github.com/zitadel/oidc/v3/pkg/http"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"gitlab.unjx.de/flohoss/godash/internal/env"
)

func NewAuthHandler(env *env.Config) *AuthHandler {
	key := []byte(env.SessionKey)
	cookieHandler := httphelper.NewCookieHandler(key, key, httphelper.WithUnsecure())
	client := &http.Client{
		Timeout: time.Minute,
	}

	options := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
		rp.WithHTTPClient(client),
		rp.WithSigningAlgsFromDiscovery(),
		rp.WithPKCE(cookieHandler),
	}

	ctx := context.Background()
	provider, err := rp.NewRelyingPartyOIDC(ctx, env.OIDCIssuer, env.OIDCClientID, env.OIDCClientSecret, env.OIDCRedirectURI, env.OIDCScopes, options...)
	if err != nil {
		slog.Error("error creating provider", "err", err.Error())
		os.Exit(1)
	}

	urlOptions := []rp.URLParamOpt{}
	if env.OIDCResponseMode != "" {
		urlOptions = append(urlOptions, rp.WithResponseModeURLParam(oidc.ResponseMode(env.OIDCResponseMode)))
	}

	marshalToken := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty) {
		data, err := json.Marshal(tokens)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(data)
	}

	return &AuthHandler{
		env:          env,
		provider:     provider,
		options:      options,
		urlOptions:   urlOptions,
		marshalToken: marshalToken,
	}
}

type AuthHandler struct {
	env          *env.Config
	provider     rp.RelyingParty
	options      []rp.Option
	urlOptions   []rp.URLParamOpt
	marshalToken func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty)
}
