package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gomodmaster/internal/config"
	"gomodmaster/internal/core"
	"gomodmaster/internal/netutil"
	httptransport "gomodmaster/internal/transport/http"
	"gomodmaster/internal/transport/ws"
	"gomodmaster/internal/version"

	"github.com/spf13/cobra"
)

func main() {
	cfg := config.DefaultConfig()
	rootCmd := &cobra.Command{
		Use:   "gmm",
		Short: "goModMaster Modbus master",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("tui mode not implemented yet")
		},
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	addGlobalFlags(rootCmd, &cfg)
	rootCmd.AddCommand(webCommand(&cfg))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addGlobalFlags(root *cobra.Command, cfg *config.Config) {
	var (
		useRTU    bool
		serial    string
		host      string
		parity    string
		dataBits  uint
		stopBits  uint
		speed     uint
		port      int
		unitID    uint
		timeoutMs int64
		showVer   bool
	)

	root.PersistentFlags().BoolVar(&useRTU, "rtu", false, "use Modbus RTU over serial")
	root.PersistentFlags().StringVar(&serial, "serial", cfg.Serial.Device, "serial device path")
	root.PersistentFlags().UintVar(&speed, "speed", cfg.Serial.Speed, "serial baud rate")
	root.PersistentFlags().UintVar(&dataBits, "databits", cfg.Serial.DataBits, "serial data bits")
	root.PersistentFlags().UintVar(&stopBits, "stopbits", cfg.Serial.StopBits, "serial stop bits")
	root.PersistentFlags().StringVar(&parity, "parity", cfg.Serial.Parity, "serial parity (none, even, odd)")
	root.PersistentFlags().StringVar(&host, "host", cfg.TCP.Host, "tcp host")
	root.PersistentFlags().IntVar(&port, "port", cfg.TCP.Port, "tcp port")
	root.PersistentFlags().UintVar(&unitID, "unit-id", uint(cfg.UnitID), "unit id")
	root.PersistentFlags().Int64Var(&timeoutMs, "timeout", cfg.TimeoutMs, "request timeout (ms)")
	root.PersistentFlags().BoolVar(&showVer, "version", false, "print version and exit")

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if showVer {
			fmt.Println(version.Version)
			os.Exit(0)
		}
		if useRTU {
			cfg.Protocol = config.ProtocolRTU
		} else {
			cfg.Protocol = config.ProtocolTCP
		}
		cfg.Serial.Device = serial
		cfg.Serial.Speed = speed
		cfg.Serial.DataBits = dataBits
		cfg.Serial.StopBits = stopBits
		cfg.Serial.Parity = parity
		cfg.TCP.Host = host
		cfg.TCP.Port = port
		cfg.UnitID = uint8(unitID)
		cfg.TimeoutMs = timeoutMs
		return nil
	}
}

func webCommand(cfg *config.Config) *cobra.Command {
	var (
		noToken bool
		listen  string
	)

	cmd := &cobra.Command{
		Use:   "web",
		Short: "Launch local web UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.ListenAddr = listen
			cfg.RequireToken = !noToken

			service := core.NewService(*cfg)
			hub := ws.NewHub()

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			go hub.Run(ctx, service.Events())

			e, err := httptransport.StartServer(service, hub)
			if err != nil {
				return err
			}

			current := service.Config()
			announce(current)

			serverErr := make(chan error, 1)
			go func() {
				serverErr <- e.Start(current.ListenAddr)
			}()

			select {
			case err := <-serverErr:
				_ = service.Disconnect()
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					return err
				}
				return nil
			case <-ctx.Done():
			}

			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := e.Shutdown(shutdownCtx); err != nil {
				_ = service.Disconnect()
				return err
			}

			_ = service.Disconnect()
			err = <-serverErr
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&listen, "listen", cfg.ListenAddr, "listen address")
	cmd.Flags().BoolVar(&noToken, "no-token", false, "disable token requirement for web UI")

	return cmd
}

func announce(cfg config.Config) {
	portSuffix := strings.TrimPrefix(cfg.ListenAddr, "0.0.0.0")
	if portSuffix == cfg.ListenAddr {
		if idx := strings.LastIndex(cfg.ListenAddr, ":"); idx != -1 {
			portSuffix = cfg.ListenAddr[idx:]
		}
	}
	if cfg.RequireToken && cfg.Token != "" {
		fmt.Printf("Token suffix:\n/?token=%s\n", cfg.Token)
	}

	addresses := netutil.DiscoverIPv4()
	if len(addresses) == 0 {
		fmt.Println("No external IPv4 addresses detected.")
		return
	}

	for _, addr := range addresses {
		if cfg.RequireToken && cfg.Token != "" {
			fmt.Printf("http://%s%s/?token=%s\n", addr, portSuffix, cfg.Token)
		} else {
			fmt.Printf("http://%s%s/\n", addr, portSuffix)
		}
	}
	fmt.Printf("Invocation: %s\n", cfg.Invocation())
}
