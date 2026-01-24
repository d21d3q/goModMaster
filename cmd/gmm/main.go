package main

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strings"

	"gomodmaster/internal/config"
	"gomodmaster/internal/core"
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
			go hub.Run(service.Events())

			e, err := httptransport.StartServer(service, hub)
			if err != nil {
				return err
			}
			defer func() {
				_ = service.Disconnect()
			}()

			current := service.Config()
			announce(current)

			return e.Start(current.ListenAddr)
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

	addresses := discoverIPv4()
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

func discoverIPv4() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	addresses := []string{}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagPointToPoint != 0 {
			continue
		}
		name := strings.ToLower(iface.Name)
		if strings.HasPrefix(name, "utun") || strings.HasPrefix(name, "docker") || strings.HasPrefix(name, "lo") {
			continue
		}
		if !isPhysicalInterface(name) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip := extractIPv4(addr)
			if ip == "" {
				continue
			}
			addresses = append(addresses, ip)
		}
	}

	sort.Strings(addresses)
	return uniqueStrings(addresses)
}

func isPhysicalInterface(name string) bool {
	for _, prefix := range []string{"en", "eth", "wlan", "wl"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func extractIPv4(addr net.Addr) string {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil {
		return ""
	}
	ip = ip.To4()
	if ip == nil {
		return ""
	}
	return ip.String()
}

func uniqueStrings(values []string) []string {
	out := []string{}
	seen := map[string]struct{}{}
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
