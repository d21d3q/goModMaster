package core

import "time"

type ReadKind string

const (
	ReadCoils          ReadKind = "coils"
	ReadDiscreteInputs ReadKind = "discrete_inputs"
	ReadHolding        ReadKind = "holding_registers"
	ReadInput          ReadKind = "input_registers"
)

type ReadRequest struct {
	Kind     ReadKind `json:"kind"`
	Address  uint16   `json:"address"`
	Quantity uint16   `json:"quantity"`
	UnitID   uint8    `json:"unitId"`
}

type ReadResult struct {
	Kind         ReadKind       `json:"kind"`
	Address      uint16         `json:"address"`
	Quantity     uint16         `json:"quantity"`
	BoolValues   []bool         `json:"boolValues,omitempty"`
	RegValues    []uint16       `json:"regValues,omitempty"`
	Decoded      []DecodedValue `json:"decoded,omitempty"`
	LatencyMs    int64          `json:"latencyMs"`
	CompletedAt  time.Time      `json:"completedAt"`
	ErrorMessage string         `json:"errorMessage,omitempty"`
}

type Stats struct {
	ReadCount     int   `json:"readCount"`
	ErrorCount    int   `json:"errorCount"`
	LastLatencyMs int64 `json:"lastLatencyMs"`
}
