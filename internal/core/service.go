package core

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"syscall"
	"time"

	"gomodmaster/internal/config"

	"github.com/simonvetter/modbus"
)

var ErrNotConnected = errors.New("modbus client not connected")

const defaultLogSize = 0

type EventType string

const (
	EventData  EventType = "data"
	EventLog   EventType = "log"
	EventStats EventType = "stats"
	EventError EventType = "error"
	EventStatus EventType = "status"
)

type Event struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
}

type Service struct {
	mu     sync.Mutex
	config config.Config
	client *modbus.ModbusClient
	logs   *LogBuffer
	stats  Stats
	events chan Event
	connecting    bool
	connectStop   chan struct{}
	lastConnError string
}

type ConnectionStatus struct {
	Connected  bool   `json:"connected"`
	Connecting bool   `json:"connecting"`
	LastError  string `json:"lastError,omitempty"`
}

func NewService(cfg config.Config) *Service {
	return &Service{
		config: cfg,
		logs:   NewLogBuffer(defaultLogSize),
		events: make(chan Event, 32),
	}
}

func (s *Service) Events() <-chan Event {
	return s.events
}

func (s *Service) Config() config.Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.config
}

func (s *Service) UpdateConfig(cfg config.Config) {
	s.mu.Lock()
	s.config = cfg
	s.mu.Unlock()
}

func (s *Service) Connect() error {
	s.mu.Lock()
	alreadyConnected := s.client != nil
	alreadyConnecting := s.connecting
	if alreadyConnected || alreadyConnecting {
		s.mu.Unlock()
		if alreadyConnected {
			s.logInfo("connect requested: already connected")
		} else {
			s.logInfo("connect requested: already connecting")
		}
		return nil
	}
	stop := make(chan struct{})
	s.connectStop = stop
	s.connecting = true
	s.mu.Unlock()

	s.logInfo("connect requested: starting loop")
	s.emitStatus()
	go s.connectLoop(stop)
	return nil
}

func (s *Service) Disconnect() error {
	s.mu.Lock()
	stop := s.connectStop
	s.connectStop = nil
	s.connecting = false
	client := s.client
	s.client = nil
	s.lastConnError = ""
	s.mu.Unlock()

	s.logInfo("disconnect requested")
	if stop != nil {
		close(stop)
	}
	if client == nil {
		s.emitStatus()
		return nil
	}
	err := client.Close()
	s.emitStatus()
	return err
}

func (s *Service) Read(req ReadRequest) (ReadResult, error) {
	start := time.Now()
	result := ReadResult{
		Kind:     req.Kind,
		Address:  req.Address,
		Quantity: req.Quantity,
	}

	s.mu.Lock()
	client := s.client
	cfg := s.config
	s.mu.Unlock()

	if client == nil {
		err := ErrNotConnected
		return s.finishWithError(result, start, err)
	}

	if req.UnitID != 0 {
		_ = client.SetUnitId(req.UnitID)
	} else {
		_ = client.SetUnitId(cfg.UnitID)
	}

	addr := applyAddressBase(req.Address, cfg.AddressBase)

	var err error
	s.logRequest(req, addr)
	switch req.Kind {
	case ReadCoils:
		result.BoolValues, err = client.ReadCoils(addr, req.Quantity)
	case ReadDiscreteInputs:
		result.BoolValues, err = client.ReadDiscreteInputs(addr, req.Quantity)
	case ReadHolding:
		result.RegValues, err = client.ReadRegisters(addr, req.Quantity, modbus.HOLDING_REGISTER)
	case ReadInput:
		result.RegValues, err = client.ReadRegisters(addr, req.Quantity, modbus.INPUT_REGISTER)
	default:
		err = fmt.Errorf("unsupported read kind: %s", req.Kind)
	}

	if err != nil {
		return s.finishWithError(result, start, err)
	}

	if len(result.RegValues) > 0 {
		result.Decoded = DecodeValues(result.RegValues, cfg.Decoders)
	}
	result.CompletedAt = time.Now()
	result.LatencyMs = time.Since(start).Milliseconds()

	s.updateStats(result.LatencyMs, "")
	s.logResponse(req, result)
	s.emit(Event{Type: EventData, Payload: result})

	return result, nil
}

func (s *Service) Stats() Stats {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stats
}

func (s *Service) Logs() []LogEntry {
	return s.logs.Snapshot()
}

func (s *Service) IsConnected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.client != nil
}

func (s *Service) IsConnecting() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.connecting
}

func (s *Service) LastConnectError() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastConnError
}

func (s *Service) finishWithError(result ReadResult, start time.Time, err error) (ReadResult, error) {
	result.CompletedAt = time.Now()
	result.LatencyMs = time.Since(start).Milliseconds()
	result.ErrorMessage = err.Error()
	result.ErrorKind = errorKind(err)

	s.updateStats(result.LatencyMs, result.ErrorMessage)
	s.logError(err.Error())
	s.emit(Event{Type: EventError, Payload: result})
	s.maybeReconnect(err)
	return result, err
}

func (s *Service) updateStats(latencyMs int64, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if errMsg != "" {
		s.stats.ErrorCount++
	} else {
		s.stats.ReadCount++
	}
	s.stats.LastLatencyMs = latencyMs
	s.emit(Event{Type: EventStats, Payload: s.stats})
}

func (s *Service) emit(event Event) {
	select {
	case s.events <- event:
	default:
	}
}

func (s *Service) logRequest(req ReadRequest, addr uint16) {
	unit := req.UnitID
	if unit == 0 {
		unit = s.config.UnitID
	}
	msg := fmt.Sprintf("tx %s fc=%s addr=0x%04x qty=0x%04x unit=0x%02x", req.Kind, functionCode(req.Kind), addr, req.Quantity, unit)
	entry := LogEntry{Time: time.Now(), Direction: "tx", Message: msg}
	s.logs.Add(entry)
	s.emit(Event{Type: EventLog, Payload: entry})
}

func (s *Service) logResponse(req ReadRequest, result ReadResult) {
	msg := fmt.Sprintf("rx %s fc=%s addr=0x%04x qty=0x%04x latency=%dms", req.Kind, functionCode(req.Kind), result.Address, result.Quantity, result.LatencyMs)
	entry := LogEntry{Time: time.Now(), Direction: "rx", Message: msg}
	s.logs.Add(entry)
	s.emit(Event{Type: EventLog, Payload: entry})
}

func functionCode(kind ReadKind) string {
	switch kind {
	case ReadCoils:
		return "01"
	case ReadDiscreteInputs:
		return "02"
	case ReadHolding:
		return "03"
	case ReadInput:
		return "04"
	default:
		return "--"
	}
}

func (s *Service) logError(msg string) {
	entry := LogEntry{Time: time.Now(), Direction: "err", Message: msg}
	s.logs.Add(entry)
	s.emit(Event{Type: EventLog, Payload: entry})
}

func (s *Service) logInfo(msg string) {
	entry := LogEntry{Time: time.Now(), Direction: "sys", Message: msg}
	s.logs.Add(entry)
	s.emit(Event{Type: EventLog, Payload: entry})
}

func (s *Service) connectLoop(stop <-chan struct{}) {
	backoff := 500 * time.Millisecond
	attempt := 0
	for {
		select {
		case <-stop:
			s.logInfo("connect stopped")
			return
		default:
		}

		s.mu.Lock()
		cfg := s.config
		s.mu.Unlock()

		attempt++
		s.logInfo(fmt.Sprintf("connect attempt %d: %s", attempt, connectionSummary(cfg)))
		client, err := newClient(cfg)
		if err == nil {
			err = client.Open()
		}

		if err == nil {
			s.mu.Lock()
			s.client = client
			s.connecting = false
			s.lastConnError = ""
			s.mu.Unlock()
			s.logInfo("connect succeeded")
			s.emitStatus()
			return
		}

		s.mu.Lock()
		s.lastConnError = err.Error()
		s.mu.Unlock()
		s.logError(fmt.Sprintf("connect failed: %v", err))
		s.emitStatus()

		timer := time.NewTimer(backoff)
		select {
		case <-stop:
			timer.Stop()
			return
		case <-timer.C:
		}
		if backoff < 5*time.Second {
			backoff *= 2
		}
	}
}

func (s *Service) emitStatus() {
	status := s.statusSnapshot()
	s.logInfo(fmt.Sprintf("status: connected=%t connecting=%t lastError=%q", status.Connected, status.Connecting, status.LastError))
	s.emit(Event{Type: EventStatus, Payload: status})
}

func (s *Service) statusSnapshot() ConnectionStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return ConnectionStatus{
		Connected:  s.client != nil,
		Connecting: s.connecting,
		LastError:  s.lastConnError,
	}
}

func (s *Service) StatusSnapshot() ConnectionStatus {
	return s.statusSnapshot()
}

func (s *Service) maybeReconnect(err error) {
	if !isConnectionError(err) {
		return
	}
	s.mu.Lock()
	alreadyConnecting := s.connecting
	hasClient := s.client != nil
	s.mu.Unlock()
	if !hasClient || alreadyConnecting {
		return
	}
	s.logInfo("connection lost; reconnecting")
	go func() {
		_ = s.Disconnect()
		s.mu.Lock()
		s.lastConnError = err.Error()
		s.mu.Unlock()
		s.emitStatus()
		_ = s.Connect()
	}()
}

func errorKind(err error) string {
	if isConnectionError(err) {
		return "connection"
	}
	return "modbus"
}

func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNABORTED) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "use of closed network connection")
}

func applyAddressBase(addr uint16, base config.AddressBase) uint16 {
	if base == config.AddressBaseOne && addr > 0 {
		return addr - 1
	}
	return addr
}

func newClient(cfg config.Config) (*modbus.ModbusClient, error) {
	var url string
	clientConfig := &modbus.ClientConfiguration{}

	switch cfg.Protocol {
	case config.ProtocolRTU:
		url = fmt.Sprintf("rtu://%s", cfg.Serial.Device)
		clientConfig.Speed = cfg.Serial.Speed
		clientConfig.DataBits = cfg.Serial.DataBits
		clientConfig.StopBits = cfg.Serial.StopBits
		clientConfig.Parity = parseParity(cfg.Serial.Parity)
	case config.ProtocolTCP:
		url = fmt.Sprintf("tcp://%s:%d", cfg.TCP.Host, cfg.TCP.Port)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", cfg.Protocol)
	}

	clientConfig.URL = url
	clientConfig.Timeout = time.Duration(cfg.TimeoutMs) * time.Millisecond

	client, err := modbus.NewClient(clientConfig)
	if err != nil {
		return nil, err
	}
	_ = client.SetUnitId(cfg.UnitID)
	return client, nil
}

func connectionSummary(cfg config.Config) string {
	switch cfg.Protocol {
	case config.ProtocolRTU:
		return fmt.Sprintf(
			"rtu://%s speed=%d data=%d stop=%d parity=%s timeout=%dms",
			cfg.Serial.Device,
			cfg.Serial.Speed,
			cfg.Serial.DataBits,
			cfg.Serial.StopBits,
			cfg.Serial.Parity,
			cfg.TimeoutMs,
		)
	case config.ProtocolTCP:
		return fmt.Sprintf("tcp://%s:%d timeout=%dms", cfg.TCP.Host, cfg.TCP.Port, cfg.TimeoutMs)
	default:
		return fmt.Sprintf("unknown protocol: %s", cfg.Protocol)
	}
}

func parseParity(value string) uint {
	switch value {
	case "even":
		return modbus.PARITY_EVEN
	case "odd":
		return modbus.PARITY_ODD
	default:
		return modbus.PARITY_NONE
	}
}
