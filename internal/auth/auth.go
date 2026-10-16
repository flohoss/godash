package auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/flohoss/godash/config"
	"github.com/flohoss/gosso"
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
	return a.core.StartLogin(c, "")
}

func (a *Auth) ConsumeLoginState(c *echo.Context) (string, string, error) {
	state, verifier, _, err := a.core.ConsumeLoginState(c)

	return state, verifier, err
}

func (a *Auth) Exchange(c *echo.Context, code, verifier string) (string, error) {
	return a.core.Exchange(c, code, verifier)
}

func (a *Auth) StartSession(c *echo.Context, subject string) {
	_ = a.core.StartSession(c, subject)
}

func (a *Auth) Authenticate(c *echo.Context) (*gosso.User, error) {
	return a.core.Authenticate(c)
}

func (a *Auth) Middleware() echo.MiddlewareFunc {
	return a.core.Middleware()
}

func (a *Auth) SessionTTL() time.Duration {
	return a.core.SessionTTL()
}

func SessionCookie(subject string) *http.Cookie {
	oidc := config.GetAuth().OIDC
	expiresAt := time.Now().Add(oidc.SessionTTL)
	if oidc.SessionTTL <= 0 {
		expiresAt = time.Now().Add(gosso.DefaultSessionTTL)
	}

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
