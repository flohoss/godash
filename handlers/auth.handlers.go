package handlers

import (
	"context"
	"log/slog"
	"os"

	"gitlab.unjx.de/flohoss/godash/internal/env"

	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/zitadel-go/v3/pkg/authentication"
	openid "github.com/zitadel/zitadel-go/v3/pkg/authentication/oidc"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
)

func NewAuthHandler(env *env.Config) *AuthHandler {
	ctx := context.Background()
	authN, err := authentication.New(ctx, zitadel.New(env.OIDCIssuerUrl), env.OIDCClientSecret,
		openid.DefaultAuthentication(env.OIDCClientId, env.OIDCRedirectUri, env.OIDCClientSecret, oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeEmail, oidc.ScopeOfflineAccess),
	)
	if err != nil {
		slog.Error("zitadel sdk could not initialize", "error", err)
		os.Exit(1)
	}
	mw := authentication.Middleware(authN)

	return &AuthHandler{
		authenticator: authN,
		middleware:    mw,
	}
}

type AuthHandler struct {
	authenticator *authentication.Authenticator[*openid.UserInfoContext[*oidc.IDTokenClaims, *oidc.UserInfo]]
	middleware    *authentication.Interceptor[*openid.UserInfoContext[*oidc.IDTokenClaims, *oidc.UserInfo]]
}
