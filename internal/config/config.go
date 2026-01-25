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
	parts := []string{"gmm", "web"}
	defaults := DefaultConfig()
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
	if !c.RequireToken {
		parts = append(parts, "--no-token")
	}
	return strings.Join(parts, " ")
}
