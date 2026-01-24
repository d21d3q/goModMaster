package core

import (
	"testing"

	"gomodmaster/internal/config"

	"github.com/stretchr/testify/require"
)

func TestDecodeUint32BigEndianHighFirst(t *testing.T) {
	regs := []uint16{0x1234, 0x5678}
	dec := config.DecoderConfig{Type: config.DecoderUint32, Endianness: config.EndianBig, WordOrder: config.WordHighFirst, Enabled: true}
	values := DecodeValues(regs, []config.DecoderConfig{dec})

	require.Len(t, values, 1)
	require.Equal(t, uint32(0x12345678), values[0].Value)
}

func TestDecodeUint32LittleEndianLowFirst(t *testing.T) {
	regs := []uint16{0x1122, 0x3344}
	dec := config.DecoderConfig{Type: config.DecoderUint32, Endianness: config.EndianLittle, WordOrder: config.WordLowFirst, Enabled: true}
	values := DecodeValues(regs, []config.DecoderConfig{dec})

	require.Len(t, values, 1)
	require.Equal(t, uint32(0x44332211), values[0].Value)
}

func TestDecodeInt16(t *testing.T) {
	regs := []uint16{0xFFFE}
	dec := config.DecoderConfig{Type: config.DecoderInt16, Endianness: config.EndianBig, WordOrder: config.WordHighFirst, Enabled: true}
	values := DecodeValues(regs, []config.DecoderConfig{dec})

	require.Len(t, values, 1)
	require.Equal(t, int16(-2), values[0].Value)
}
