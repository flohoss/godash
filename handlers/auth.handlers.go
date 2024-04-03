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
	ah := AuthHandler{
		redirectUri: env.OIDCRedirectUri,
	}

	if env.OIDCIssuerUrl != "" {
		ctx := context.Background()
		var err error
		ah.oidc, err = authentication.New(ctx, zitadel.New(env.OIDCIssuerUrl), env.OIDCKey,
			openid.DefaultAuthentication(env.OIDCClientId, env.OIDCRedirectUri, env.OIDCKey),
		)
		if err != nil {
			slog.Error("zitadel sdk could not initialize", "error", err)
			os.Exit(1)
		}
		ah.mw = authentication.Middleware(ah.oidc)
	}

	return &ah
}

type AuthHandler struct {
	oidc        *authentication.Authenticator[*openid.UserInfoContext[*oidc.IDTokenClaims, *oidc.UserInfo]]
	mw          *authentication.Interceptor[*openid.UserInfoContext[*oidc.IDTokenClaims, *oidc.UserInfo]]
	redirectUri string
}
