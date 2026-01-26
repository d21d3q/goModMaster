package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gomodmaster/internal/config"
	"gomodmaster/internal/core"
	"gomodmaster/internal/netutil"
	httptransport "gomodmaster/internal/transport/http"
	"gomodmaster/internal/transport/ws"
	"gomodmaster/internal/tui"
	"gomodmaster/internal/version"

	"github.com/spf13/cobra"
)

func main() {
	cfg := config.DefaultConfig()
	rootCmd := &cobra.Command{
		Use:   "gmm",
		Short: "goModMaster Modbus master",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Run(cfg)
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
		serial    string
		host      string
		framing   string
		parity    string
		dataBits  uint
		stopBits  uint
		speed     uint
		port      int
		unitID    uint
		timeoutMs int64
		address   string
		count     uint
		function  string
		addrBase  uint
		addrFmt   string
		valueBase string
		u16Spec   string
		i16Spec   string
		u32Spec   string
		i32Spec   string
		f32Spec   string
		showVer   bool
	)

	root.PersistentFlags().StringVar(&serial, "serial", cfg.Serial.Device, "serial device path (enables serial mode)")
	root.PersistentFlags().UintVar(&speed, "speed", cfg.Serial.Speed, "serial baud rate")
	root.PersistentFlags().UintVar(&dataBits, "databits", cfg.Serial.DataBits, "serial data bits")
	root.PersistentFlags().UintVar(&stopBits, "stopbits", cfg.Serial.StopBits, "serial stop bits")
	root.PersistentFlags().StringVar(&parity, "parity", cfg.Serial.Parity, "serial parity (none, even, odd)")
	root.PersistentFlags().StringVar(&framing, "framing", string(config.ProtocolRTU), "serial framing (rtu, ascii)")
	root.PersistentFlags().StringVar(&host, "host", cfg.TCP.Host, "tcp host")
	root.PersistentFlags().IntVar(&port, "port", cfg.TCP.Port, "tcp port")
	root.PersistentFlags().UintVar(&unitID, "unit-id", uint(cfg.UnitID), "unit id")
	root.PersistentFlags().Int64Var(&timeoutMs, "timeout", cfg.TimeoutMs, "request timeout (ms)")
	root.PersistentFlags().StringVar(&address, "address", fmt.Sprintf("%d", cfg.ReadAddress), "default read address (decimal or 0x...)")
	root.PersistentFlags().UintVar(&count, "count", uint(cfg.ReadQuantity), "default read count")
	root.PersistentFlags().StringVar(&function, "function", cfg.ReadKind, "default function (01/02/03/04 or coils/discrete_inputs/holding_registers/input_registers)")
	root.PersistentFlags().UintVar(&addrBase, "address-base", uint(cfg.AddressBase), "address base (0 or 1)")
	root.PersistentFlags().StringVar(&addrFmt, "address-format", formatBaseHelp(cfg.AddressFormat), "address format (dec or hex)")
	root.PersistentFlags().StringVar(&valueBase, "value-base", formatBaseHelp(cfg.ValueBase), "value format (dec or hex)")
	root.PersistentFlags().StringVar(&u16Spec, "u16", "", "enable uint16 decoder (be/le[,hf/lf])")
	root.PersistentFlags().StringVar(&i16Spec, "i16", "", "enable int16 decoder (be/le[,hf/lf])")
	root.PersistentFlags().StringVar(&u32Spec, "u32", "", "enable uint32 decoder (be/le[,hf/lf])")
	root.PersistentFlags().StringVar(&i32Spec, "i32", "", "enable int32 decoder (be/le[,hf/lf])")
	root.PersistentFlags().StringVar(&f32Spec, "f32", "", "enable float32 decoder (be/le[,hf/lf])")
	root.PersistentFlags().BoolVar(&showVer, "version", false, "print version and exit")

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if showVer {
			fmt.Println(version.Version)
			os.Exit(0)
		}
		flags := cmd.Flags()
		serialMode := flags.Changed("serial") ||
			flags.Changed("speed") ||
			flags.Changed("databits") ||
			flags.Changed("stopbits") ||
			flags.Changed("parity") ||
			flags.Changed("framing")
		tcpMode := flags.Changed("host") || flags.Changed("port")
		if serialMode && tcpMode {
			return fmt.Errorf("serial flags cannot be combined with --host/--port")
		}
		if serialMode {
			if framing != string(config.ProtocolRTU) {
				return fmt.Errorf("unsupported framing: %s", framing)
			}
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
		readAddress, err := parseReadAddress(address)
		if err != nil {
			return err
		}
		cfg.ReadAddress = readAddress
		if count == 0 || count > 0xffff {
			return fmt.Errorf("count must be 1-65535")
		}
		cfg.ReadQuantity = uint16(count)
		readKind, err := parseReadKind(function)
		if err != nil {
			return err
		}
		cfg.ReadKind = readKind
		base, err := parseAddressBase(addrBase)
		if err != nil {
			return err
		}
		cfg.AddressBase = base
		format, err := parseValueBase(addrFmt)
		if err != nil {
			return err
		}
		cfg.AddressFormat = format
		value, err := parseValueBase(valueBase)
		if err != nil {
			return err
		}
		cfg.ValueBase = value
		if err := applyDecoderOverrides(cfg, defaultsDecoders(cfg), decoderSpec{spec: u16Spec, changed: flags.Changed("u16"), typ: config.DecoderUint16},
			decoderSpec{spec: i16Spec, changed: flags.Changed("i16"), typ: config.DecoderInt16},
			decoderSpec{spec: u32Spec, changed: flags.Changed("u32"), typ: config.DecoderUint32},
			decoderSpec{spec: i32Spec, changed: flags.Changed("i32"), typ: config.DecoderInt32},
			decoderSpec{spec: f32Spec, changed: flags.Changed("f32"), typ: config.DecoderFloat32}); err != nil {
			return err
		}
		return nil
	}
}

func parseReadAddress(value string) (uint16, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return 0, fmt.Errorf("address cannot be empty")
	}
	var parsed int64
	var err error
	if strings.HasPrefix(value, "0x") {
		parsed, err = strconv.ParseInt(value[2:], 16, 32)
	} else {
		parsed, err = strconv.ParseInt(value, 10, 32)
	}
	if err != nil || parsed < 0 || parsed > 0xffff {
		return 0, fmt.Errorf("address must be 0-65535")
	}
	return uint16(parsed), nil
}

func parseReadKind(value string) (string, error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	switch trimmed {
	case "1", "01", "coils", "coil":
		return "coils", nil
	case "2", "02", "discrete_inputs", "discrete":
		return "discrete_inputs", nil
	case "3", "03", "holding_registers", "holding":
		return "holding_registers", nil
	case "4", "04", "input_registers", "input":
		return "input_registers", nil
	default:
		return "", fmt.Errorf("unsupported function: %s", value)
	}
}

func parseAddressBase(value uint) (config.AddressBase, error) {
	switch value {
	case 0:
		return config.AddressBaseZero, nil
	case 1:
		return config.AddressBaseOne, nil
	default:
		return config.AddressBaseZero, fmt.Errorf("address-base must be 0 or 1")
	}
}

func parseValueBase(value string) (config.ValueBase, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "10", "dec", "decimal":
		return config.ValueBaseDec, nil
	case "16", "hex", "hexadecimal":
		return config.ValueBaseHex, nil
	default:
		return config.ValueBaseDec, fmt.Errorf("unsupported base: %s", value)
	}
}

func formatBaseHelp(base config.ValueBase) string {
	if base == config.ValueBaseHex {
		return "hex"
	}
	return "dec"
}

type decoderSpec struct {
	spec    string
	changed bool
	typ     config.DecoderType
}

func applyDecoderOverrides(cfg *config.Config, defaults map[config.DecoderType]config.DecoderConfig, specs ...decoderSpec) error {
	for _, entry := range specs {
		if !entry.changed {
			continue
		}
		next, err := parseDecoderSpec(entry.spec, defaults[entry.typ])
		if err != nil {
			return fmt.Errorf("%s: %w", entry.typ, err)
		}
		next.Type = entry.typ
		setDecoder(cfg, next)
	}
	return nil
}

func defaultsDecoders(cfg *config.Config) map[config.DecoderType]config.DecoderConfig {
	defaults := map[config.DecoderType]config.DecoderConfig{}
	for _, decoder := range cfg.Decoders {
		defaults[decoder.Type] = decoder
	}
	return defaults
}

func parseDecoderSpec(spec string, defaults config.DecoderConfig) (config.DecoderConfig, error) {
	if strings.TrimSpace(spec) == "" {
		return config.DecoderConfig{}, fmt.Errorf("decoder flag requires a value")
	}
	parts := strings.FieldsFunc(strings.ToLower(strings.TrimSpace(spec)), func(r rune) bool {
		return r == ',' || r == ' '
	})
	decoder := defaults
	decoder.Enabled = true
	for _, part := range parts {
		switch part {
		case "be":
			decoder.Endianness = config.EndianBig
		case "le":
			decoder.Endianness = config.EndianLittle
		case "hf":
			decoder.WordOrder = config.WordHighFirst
		case "lf":
			decoder.WordOrder = config.WordLowFirst
		case "":
			continue
		default:
			return config.DecoderConfig{}, fmt.Errorf("unsupported decoder option: %s", part)
		}
	}
	return decoder, nil
}

func setDecoder(cfg *config.Config, decoder config.DecoderConfig) {
	for idx := range cfg.Decoders {
		if cfg.Decoders[idx].Type == decoder.Type {
			cfg.Decoders[idx] = decoder
			return
		}
	}
	cfg.Decoders = append(cfg.Decoders, decoder)
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
	// if cfg.RequireToken && cfg.Token != "" {
	// 	fmt.Printf("Token suffix:\n/?token=%s\n", cfg.Token)
	// }

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
}
