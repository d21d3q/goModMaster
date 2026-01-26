package http

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"io/fs"
	"net/http"
	"strings"

	appembed "gomodmaster/embed"
	"gomodmaster/internal/core"
	"gomodmaster/internal/transport/ws"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const tokenCookieName = "gmm_token"

func StartServer(service *core.Service, hub *ws.Hub) (*echo.Echo, error) {
	cfg := service.Config()
	if cfg.RequireToken && cfg.Token == "" {
		cfg.Token = generateToken(16)
		service.UpdateConfig(cfg)
	}

	e := echo.New()
	e.HideBanner = true
	e.Logger.SetOutput(io.Discard)
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(cacheMiddleware())

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			config := service.Config()
			if !config.RequireToken {
				return next(c)
			}
			if shouldRedirectToken(c) && trySetTokenCookie(c, config.Token) {
				return redirectWithoutToken(c)
			}
			if isProtectedPath(c) {
				if hasValidToken(c, config.Token) {
					return next(c)
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid token")
			}
			return next(c)
		}
	})

	routes(e, service, hub)
	if err := ServeStatic(e); err != nil {
		return nil, err
	}

	return e, nil
}

func ServeStatic(e *echo.Echo) error {
	webFS, err := appembed.WebFS()
	if err != nil {
		return err
	}

	sub, err := fs.Sub(webFS, ".")
	if err != nil {
		return err
	}

	e.GET("/*", echo.WrapHandler(http.FileServer(http.FS(sub))))

	return nil
}

func hasValidToken(c echo.Context, token string) bool {
	if token == "" {
		return true
	}
	if c.QueryParam("token") == token {
		setTokenCookie(c, token)
		return true
	}
	if c.Request().Header.Get("X-GMM-Token") == token {
		setTokenCookie(c, token)
		return true
	}
	cookie, err := c.Cookie(tokenCookieName)
	if err == nil && cookie.Value == token {
		return true
	}
	return false
}

func setTokenCookie(c echo.Context, token string) {
	cookie := &http.Cookie{
		Name:     tokenCookieName,
		Value:    token,
		HttpOnly: true,
		Path:     "/",
	}
	c.SetCookie(cookie)
}

func trySetTokenCookie(c echo.Context, token string) bool {
	if token == "" {
		return false
	}
	if c.QueryParam("token") == token {
		setTokenCookie(c, token)
		return true
	}
	return false
}

func shouldRedirectToken(c echo.Context) bool {
	if c.Request().Method != http.MethodGet {
		return false
	}
	requestPath := c.Request().URL.Path
	if strings.HasPrefix(requestPath, "/api/") || requestPath == "/ws" {
		return false
	}
	return true
}

func isProtectedPath(c echo.Context) bool {
	requestPath := c.Request().URL.Path
	return strings.HasPrefix(requestPath, "/api/") || requestPath == "/ws"
}

func redirectWithoutToken(c echo.Context) error {
	req := c.Request()
	nextURL := *req.URL
	query := nextURL.Query()
	query.Del("token")
	nextURL.RawQuery = query.Encode()
	return c.Redirect(http.StatusFound, nextURL.RequestURI())
}

func generateToken(length int) string {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}

func cacheMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.HasPrefix(c.Request().URL.Path, "/assets/") {
				c.Response().Header().Set("Cache-Control", "public, max-age=31536000")
			}
			return next(c)
		}
	}
}
