package auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/flohoss/godash/config"
	"github.com/flohoss/gosso/v2"
)

const (
	SessionCookieName = "godash_session"

	loginPath    = "/auth/login"
	callbackPath = "/auth/callback"
)

type Auth struct {
	core *gosso.Auth
}

type LoginError = gosso.LoginError

type User = gosso.User

var ErrProviderUnavailable = gosso.ErrProviderUnavailable

const (
	LoginErrorRejected    = gosso.LoginErrorRejected
	LoginErrorExpired     = gosso.LoginErrorExpired
	LoginErrorUnavailable = gosso.LoginErrorUnavailable
	LoginErrorFailed      = gosso.LoginErrorFailed
)

func ClassifyLoginError(err error) LoginError {
	return gosso.ClassifyLoginError(err)
}

func New() *Auth {
	return &Auth{core: gosso.New(settings, newSessions(), callbackPath, loginPath)}
}

func (a *Auth) Enabled() bool {
	return a.core.Enabled()
}

func (a *Auth) StartLogin(c *echo.Context) error {
	return a.core.StartLogin(c.Response(), c.Request(), "")
}

func (a *Auth) HandleCallback(c *echo.Context) (*gosso.LoginResult, *gosso.LoginFailure) {
	return a.core.HandleCallback(c.Response(), c.Request())
}

func (a *Auth) Authenticate(c *echo.Context) (*gosso.User, error) {
	return a.core.Authenticate(c.Response(), c.Request())
}

func (a *Auth) Middleware() echo.MiddlewareFunc {
	echoesCore := a.core.Middleware()

	return echo.WrapMiddleware(echoesCore)
}

func (a *Auth) SessionTTL() time.Duration {
	return a.core.SessionTTL()
}

func SessionCookie(subject string) *http.Cookie {
	oidc := config.GetAuth().OIDC
	expiresAt := time.Now().Add(gosso.Config{SessionTTL: oidc.SessionTTL}.EffectiveSessionTTL())

	return &http.Cookie{
		Name:  SessionCookieName,
		Value: gosso.SignHMAC(oidc.SessionKey, subject, expiresAt),
	}
}

func (a *Auth) Shutdown() {}

func newSessions() gosso.SessionManager {
	return gosso.NewHMACSessions(gosso.HMACOptions{
		CookieName: SessionCookieName,
		Settings:   settings,
	})
}

func settings() gosso.Config {
	oidc := config.GetAuth().OIDC

	return gosso.Config{
		Enabled:      oidc.Enabled,
		IssuerURL:    oidc.IssuerURL,
		ClientID:     oidc.ClientID,
		ClientSecret: oidc.ClientSecret,
		SessionKey:   oidc.SessionKey,
		SessionTTL:   oidc.SessionTTL,
		CookieSecure: oidc.CookieSecure,
		AppURL:       config.GetAppUrl(),
	}
}
