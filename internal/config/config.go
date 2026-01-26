package config

import (
	"fmt"
	"strings"
	"time"
)

type Protocol string

type AddressBase int

type ValueBase int

const (
	ProtocolTCP Protocol = "tcp"
	ProtocolRTU Protocol = "rtu"

	AddressBaseZero AddressBase = 0
	AddressBaseOne  AddressBase = 1

	ValueBaseDec ValueBase = 10
	ValueBaseHex ValueBase = 16
)

type Endianness string

type WordOrder string

type DecoderType string

const (
	EndianBig    Endianness = "big"
	EndianLittle Endianness = "little"

	WordHighFirst WordOrder = "high-first"
	WordLowFirst  WordOrder = "low-first"

	DecoderUint16  DecoderType = "uint16"
	DecoderInt16   DecoderType = "int16"
	DecoderUint32  DecoderType = "uint32"
	DecoderInt32   DecoderType = "int32"
	DecoderFloat32 DecoderType = "float32"
)

type SerialConfig struct {
	Device   string `json:"device"`
	Speed    uint   `json:"speed"`
	DataBits uint   `json:"dataBits"`
	Parity   string `json:"parity"`
	StopBits uint   `json:"stopBits"`
}

type TCPConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type DecoderConfig struct {
	Type       DecoderType `json:"type"`
	Endianness Endianness  `json:"endianness"`
	WordOrder  WordOrder   `json:"wordOrder"`
	Enabled    bool        `json:"enabled"`
}

type Config struct {
	Protocol      Protocol        `json:"protocol"`
	UnitID        uint8           `json:"unitId"`
	TimeoutMs     int64           `json:"timeoutMs"`
	ReadKind      string          `json:"readKind"`
	ReadAddress   uint16          `json:"readAddress"`
	ReadQuantity  uint16          `json:"readQuantity"`
	AddressBase   AddressBase     `json:"addressBase"`
	AddressFormat ValueBase       `json:"addressFormat"`
	ValueBase     ValueBase       `json:"valueBase"`
	Serial        SerialConfig    `json:"serial"`
	TCP           TCPConfig       `json:"tcp"`
	Decoders      []DecoderConfig `json:"decoders"`
	ListenAddr    string          `json:"listenAddr"`
	RequireToken  bool            `json:"requireToken"`
	Token         string          `json:"token"`
}

func DefaultConfig() Config {
	return Config{
		Protocol:      ProtocolTCP,
		UnitID:        1,
		TimeoutMs:     int64((1 * time.Second).Milliseconds()),
		ReadKind:      "holding_registers",
		ReadAddress:   0,
		ReadQuantity:  1,
		AddressBase:   AddressBaseZero,
		AddressFormat: ValueBaseDec,
		ValueBase:     ValueBaseDec,
		Serial: SerialConfig{
			Device:   "/dev/ttyUSB0",
			Speed:    9600,
			DataBits: 8,
			Parity:   "none",
			StopBits: 1,
		},
		TCP: TCPConfig{
			Host: "127.0.0.1",
			Port: 502,
		},
		Decoders: []DecoderConfig{
			{Type: DecoderUint16, Endianness: EndianBig, WordOrder: WordHighFirst, Enabled: false},
			{Type: DecoderInt16, Endianness: EndianBig, WordOrder: WordHighFirst, Enabled: false},
			{Type: DecoderUint32, Endianness: EndianBig, WordOrder: WordHighFirst, Enabled: false},
			{Type: DecoderInt32, Endianness: EndianBig, WordOrder: WordHighFirst, Enabled: false},
			{Type: DecoderFloat32, Endianness: EndianBig, WordOrder: WordHighFirst, Enabled: false},
		},
		ListenAddr:   "0.0.0.0:8502",
		RequireToken: true,
	}
}

func (c Config) Invocation() string {
	return c.invocation(true, false)
}

func (c Config) InvocationTUI() string {
	return c.invocation(false, false)
}

func (c Config) InvocationFull() string {
	return c.invocation(true, true)
}

func (c Config) InvocationFullTUI() string {
	return c.invocation(false, true)
}

func (c Config) invocation(includeWeb bool, includeAll bool) string {
	parts := []string{"gmm"}
	if includeWeb {
		parts = append(parts, "web")
	}
	defaults := DefaultConfig()
	if includeAll {
		if c.Protocol == ProtocolRTU {
			parts = append(parts, "--serial", c.Serial.Device)
			if c.Serial.Speed != defaults.Serial.Speed {
				parts = append(parts, "--speed", fmt.Sprintf("%d", c.Serial.Speed))
			}
			if c.Serial.DataBits != defaults.Serial.DataBits {
				parts = append(parts, "--databits", fmt.Sprintf("%d", c.Serial.DataBits))
			}
			if c.Serial.Parity != defaults.Serial.Parity {
				parts = append(parts, "--parity", c.Serial.Parity)
			}
			if c.Serial.StopBits != defaults.Serial.StopBits {
				parts = append(parts, "--stopbits", fmt.Sprintf("%d", c.Serial.StopBits))
			}
		} else {
			if c.TCP.Host != defaults.TCP.Host {
				parts = append(parts, "--host", c.TCP.Host)
			}
			if c.TCP.Port != defaults.TCP.Port {
				parts = append(parts, "--port", fmt.Sprintf("%d", c.TCP.Port))
			}
		}
		if c.UnitID != defaults.UnitID {
			parts = append(parts, "--unit-id", fmt.Sprintf("%d", c.UnitID))
		}
		if c.TimeoutMs != defaults.TimeoutMs {
			parts = append(parts, "--timeout", fmt.Sprintf("%d", c.TimeoutMs))
		}
		if c.ReadAddress != defaults.ReadAddress {
			parts = append(parts, "--address", formatReadAddress(c.ReadAddress, c.AddressFormat))
		}
		if c.ReadQuantity != defaults.ReadQuantity {
			parts = append(parts, "--count", fmt.Sprintf("%d", c.ReadQuantity))
		}
		if c.ReadKind != defaults.ReadKind {
			parts = append(parts, "--function", readKindCode(c.ReadKind))
		}
		if c.AddressBase != defaults.AddressBase {
			parts = append(parts, "--address-base", fmt.Sprintf("%d", c.AddressBase))
		}
		if c.AddressFormat != defaults.AddressFormat {
			parts = append(parts, "--address-format", formatBaseFlag(c.AddressFormat))
		}
		if c.ValueBase != defaults.ValueBase {
			parts = append(parts, "--value-base", formatBaseFlag(c.ValueBase))
		}
		defaultDecoders := map[DecoderType]DecoderConfig{}
		for _, dec := range defaults.Decoders {
			defaultDecoders[dec.Type] = dec
		}
		for _, decoder := range c.Decoders {
			if !decoder.Enabled {
				continue
			}
			flag, value, ok := decoderInvocation(decoder, defaultDecoders[decoder.Type])
			if ok {
				parts = append(parts, flag, value)
			}
		}
		if includeWeb {
			if c.ListenAddr != defaults.ListenAddr {
				parts = append(parts, "--listen", c.ListenAddr)
			}
			if !c.RequireToken {
				parts = append(parts, "--no-token")
			}
		}
	} else {
		if c.Protocol == ProtocolRTU {
			parts = append(parts, "--serial", c.Serial.Device)
			if c.Serial.Speed != defaults.Serial.Speed {
				parts = append(parts, "--speed", fmt.Sprintf("%d", c.Serial.Speed))
			}
			if c.Serial.DataBits != defaults.Serial.DataBits {
				parts = append(parts, "--databits", fmt.Sprintf("%d", c.Serial.DataBits))
			}
			if c.Serial.Parity != defaults.Serial.Parity {
				parts = append(parts, "--parity", c.Serial.Parity)
			}
			if c.Serial.StopBits != defaults.Serial.StopBits {
				parts = append(parts, "--stopbits", fmt.Sprintf("%d", c.Serial.StopBits))
			}
		} else {
			if c.TCP.Host != defaults.TCP.Host {
				parts = append(parts, "--host", c.TCP.Host)
			}
			if c.TCP.Port != defaults.TCP.Port {
				parts = append(parts, "--port", fmt.Sprintf("%d", c.TCP.Port))
			}
		}
		if c.UnitID != defaults.UnitID {
			parts = append(parts, "--unit-id", fmt.Sprintf("%d", c.UnitID))
		}
		if includeWeb && !c.RequireToken {
			parts = append(parts, "--no-token")
		}
	}
	return strings.Join(parts, " ")
}

func readKindCode(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "coils":
		return "01"
	case "discrete_inputs":
		return "02"
	case "holding_registers":
		return "03"
	case "input_registers":
		return "04"
	default:
		if kind != "" {
			return kind
		}
		return "03"
	}
}

func formatReadAddress(address uint16, format ValueBase) string {
	if format == ValueBaseHex {
		return fmt.Sprintf("0x%X", address)
	}
	return fmt.Sprintf("%d", address)
}

func formatBaseFlag(base ValueBase) string {
	if base == ValueBaseHex {
		return "hex"
	}
	return "dec"
}

func decoderInvocation(decoder DecoderConfig, defaults DecoderConfig) (string, string, bool) {
	if !decoder.Enabled {
		return "", "", false
	}
	flag := ""
	switch decoder.Type {
	case DecoderUint16:
		flag = "--u16"
	case DecoderInt16:
		flag = "--i16"
	case DecoderUint32:
		flag = "--u32"
	case DecoderInt32:
		flag = "--i32"
	case DecoderFloat32:
		flag = "--f32"
	default:
		return "", "", false
	}
	parts := []string{}
	if decoder.Endianness != defaults.Endianness {
		if decoder.Endianness == EndianLittle {
			parts = append(parts, "le")
		} else {
			parts = append(parts, "be")
		}
	}
	if decoder.WordOrder != defaults.WordOrder {
		if decoder.WordOrder == WordLowFirst {
			parts = append(parts, "lf")
		} else {
			parts = append(parts, "hf")
		}
	}
	if len(parts) == 0 {
		if defaults.Endianness == EndianLittle {
			parts = append(parts, "le")
		} else {
			parts = append(parts, "be")
		}
	}
	return flag, strings.Join(parts, ","), true
}
