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
	a := AuthHandler{
		env: env,
	}
	if env.SSODomain != "" {
		ctx := context.Background()
		authN, err := authentication.New(ctx, zitadel.New(env.SSODomain), env.SSOKey,
			openid.DefaultAuthentication(env.SSOClientId, env.PublicUrl+"/auth/callback", env.SSOKey),
		)
		if err != nil {
			slog.Error("zitadel sdk could not initialize", "error", err)
			os.Exit(1)
		}
		a.authN = authN
		a.middleware = authentication.Middleware(authN)
	}
	return &a
}

type AuthHandler struct {
	env        *env.Config
	authN      *authentication.Authenticator[*openid.UserInfoContext[*oidc.IDTokenClaims, *oidc.UserInfo]]
	middleware *authentication.Interceptor[*openid.UserInfoContext[*oidc.IDTokenClaims, *oidc.UserInfo]]
}
