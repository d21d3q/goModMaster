package http

import (
	"crypto/rand"
	"encoding/hex"
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
	e.Use(middleware.Recover())

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			config := service.Config()
			if !config.RequireToken {
				return next(c)
			}
			if hasValidToken(c, config.Token) {
				return next(c)
			}
			return echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid token")
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

	fileServer := http.FileServer(http.FS(sub))
	e.GET("/*", func(c echo.Context) error {
		path := strings.TrimPrefix(c.Request().URL.Path, "/")
		if path == "" || path == "/" {
			path = "index.html"
		}
		if exists(sub, path) {
			c.Request().URL.Path = "/" + path
			fileServer.ServeHTTP(c.Response(), c.Request())
			return nil
		}

		c.Request().URL.Path = "/index.html"
		fileServer.ServeHTTP(c.Response(), c.Request())
		return nil
	})

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

func exists(files fs.FS, path string) bool {
	file, err := files.Open(path)
	if err != nil {
		return false
	}
	_ = file.Close()
	return true
}

func generateToken(length int) string {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}
