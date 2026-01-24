package http

import (
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"

	"gomodmaster/internal/config"
	"gomodmaster/internal/core"
	"gomodmaster/internal/transport/ws"
	"gomodmaster/internal/version"

	"github.com/labstack/echo/v4"
)

type configResponse struct {
	Config     config.Config `json:"config"`
	Invocation string        `json:"invocation"`
}

type serialDevicesResponse struct {
	Devices []string `json:"devices"`
}

func routes(e *echo.Echo, service *core.Service, hub *ws.Hub) {
	e.GET("/api/config", func(c echo.Context) error {
		cfg := service.Config()
		return c.JSON(http.StatusOK, configResponse{Config: cfg, Invocation: cfg.Invocation()})
	})

	e.POST("/api/config", func(c echo.Context) error {
		var cfg config.Config
		if err := c.Bind(&cfg); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		current := service.Config()
		if cfg.RequireToken && cfg.Token == "" {
			cfg.Token = current.Token
		}
		service.UpdateConfig(cfg)
		return c.JSON(http.StatusOK, configResponse{Config: cfg, Invocation: cfg.Invocation()})
	})

	e.POST("/api/connect", func(c echo.Context) error {
		if err := service.Connect(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]any{
			"connected":  service.IsConnected(),
			"connecting": service.IsConnecting(),
		})
	})

	e.POST("/api/disconnect", func(c echo.Context) error {
		if err := service.Disconnect(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]any{
			"connected":  service.IsConnected(),
			"connecting": service.IsConnecting(),
		})
	})

	e.POST("/api/read", func(c echo.Context) error {
		var req core.ReadRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		result, err := service.Read(req)
		if err != nil {
			return c.JSON(http.StatusBadRequest, result)
		}
		return c.JSON(http.StatusOK, result)
	})

	e.GET("/api/stats", func(c echo.Context) error {
		return c.JSON(http.StatusOK, service.Stats())
	})

	e.GET("/api/status", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{
			"connected":  service.IsConnected(),
			"connecting": service.IsConnecting(),
			"lastError":  service.LastConnectError(),
		})
	})

	e.GET("/api/version", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"version": version.Version})
	})

	e.GET("/api/serial-devices", func(c echo.Context) error {
		return c.JSON(http.StatusOK, serialDevicesResponse{Devices: serialDeviceOptions()})
	})

	e.GET("/ws", func(c echo.Context) error {
		return hub.Handle(c)
	})
}

func serialDeviceOptions() []string {
	switch runtime.GOOS {
	case "darwin":
		return globDevices([]string{"/dev/tty.*"})
	case "linux":
		return globDevices([]string{"/dev/serial/by-id/*"})
	case "windows":
		return windowsComPorts(32)
	default:
		return []string{}
	}
}

func globDevices(patterns []string) []string {
	devices := []string{}
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		devices = append(devices, matches...)
	}
	return devices
}

func windowsComPorts(max int) []string {
	devices := make([]string, 0, max)
	for i := 1; i <= max; i++ {
		devices = append(devices, fmt.Sprintf("COM%d", i))
	}
	return devices
}
