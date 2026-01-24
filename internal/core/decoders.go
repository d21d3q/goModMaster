package core

import (
	"encoding/binary"
	"math"

	"gomodmaster/internal/config"
)

type DecodedValue struct {
	Type  config.DecoderType `json:"type"`
	Value interface{}        `json:"value"`
}

func DecodeValues(regs []uint16, decoders []config.DecoderConfig) []DecodedValue {
	out := make([]DecodedValue, 0, len(decoders))
	for _, decoder := range decoders {
		if !decoder.Enabled {
			continue
		}
		value, ok := decodeWithConfig(regs, decoder)
		if !ok {
			continue
		}
		out = append(out, DecodedValue{Type: decoder.Type, Value: value})
	}
	return out
}

func decodeWithConfig(regs []uint16, decoder config.DecoderConfig) (interface{}, bool) {
	switch decoder.Type {
	case config.DecoderUint16:
		if len(regs) < 1 {
			return nil, false
		}
		return regs[0], true
	case config.DecoderInt16:
		if len(regs) < 1 {
			return nil, false
		}
		return int16(regs[0]), true
	case config.DecoderUint32:
		value, ok := decodeUint32(regs, decoder)
		return value, ok
	case config.DecoderInt32:
		value, ok := decodeUint32(regs, decoder)
		if !ok {
			return nil, false
		}
		return int32(value), true
	case config.DecoderFloat32:
		value, ok := decodeUint32(regs, decoder)
		if !ok {
			return nil, false
		}
		return math.Float32frombits(value), true
	default:
		return nil, false
	}
}

func decodeUint32(regs []uint16, decoder config.DecoderConfig) (uint32, bool) {
	if len(regs) < 2 {
		return 0, false
	}
	ordered := regs[:2]
	if decoder.WordOrder == config.WordLowFirst {
		ordered = []uint16{regs[1], regs[0]}
	}
	bytes := make([]byte, 0, 4)
	for _, reg := range ordered {
		if decoder.Endianness == config.EndianLittle {
			bytes = append(bytes, byte(reg&0xff), byte(reg>>8))
		} else {
			bytes = append(bytes, byte(reg>>8), byte(reg&0xff))
		}
	}
	return binary.BigEndian.Uint32(bytes), true
}
