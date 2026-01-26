package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"gomodmaster/internal/config"
	"gomodmaster/internal/core"

	tea "github.com/charmbracelet/bubbletea"
)

type viewMode int

type fieldFocus int

type readKindOption struct {
	label string
	kind  core.ReadKind
	code  string
}

const (
	minWidth  = 80
	minHeight = 24
)

const (
	viewMain viewMode = iota
	viewLogs
	viewHelp
	viewDecoder
	viewConnection
	viewFunctionSelect
	viewDeviceSelect
)

const (
	focusNone fieldFocus = iota
	focusAddress
	focusQuantity
	focusUnitID
	focusConnHost
	focusConnPort
	focusConnDevice
	focusConnSpeed
	focusConnDataBits
	focusConnParity
	focusConnStopBits
	focusConnTimeout
	focusConnUnitID
)

var readKinds = []readKindOption{
	{label: "Coils", kind: core.ReadCoils, code: "01"},
	{label: "Discrete Inputs", kind: core.ReadDiscreteInputs, code: "02"},
	{label: "Holding Registers", kind: core.ReadHolding, code: "03"},
	{label: "Input Registers", kind: core.ReadInput, code: "04"},
}

func readKindIndex(kind string) int {
	normalized := strings.TrimSpace(strings.ToLower(kind))
	for idx, option := range readKinds {
		if normalized == strings.ToLower(string(option.kind)) {
			return idx
		}
	}
	return 0
}

type model struct {
	cfg                config.Config
	service            *core.Service
	width              int
	height             int
	view               viewMode
	addressValue       string
	quantityValue      string
	unitValue          string
	selectedKindIdx    int
	lastResult         *core.ReadResult
	logs               []core.LogEntry
	logLimit           int
	stats              core.Stats
	status             core.ConnectionStatus
	autoConnect        bool
	pendingRead        *core.ReadRequest
	addressError       string
	quantityError      string
	unitError          string
	editActive         bool
	editField          fieldFocus
	editInput          inputModel
	editError          string
	decoderCursor      int
	functionCursor     int
	deviceCursor       int
	deviceList         []string
	valueTableCache    string
	valueTableKey      string
	connectionBoxCache string
	connectionBoxKey   string
	resultBoxCache     string
	resultBoxKey       string
	printInvocation    bool
}

type eventMsg struct {
	event core.Event
}

type readResultMsg struct {
	result core.ReadResult
	err    error
}

type errorMsg struct {
	err error
}

func newModel(cfg config.Config, service *core.Service) model {
	return model{
		cfg:           cfg,
		service:       service,
		view:          viewMain,
		addressValue:  fmt.Sprintf("%d", cfg.ReadAddress),
		quantityValue: fmt.Sprintf("%d", cfg.ReadQuantity),
		unitValue:     fmt.Sprintf("%d", cfg.UnitID),
		selectedKindIdx: func() int {
			return readKindIndex(cfg.ReadKind)
		}(),
		logLimit:      logBufferSize,
		autoConnect:   true,
		status:        service.StatusSnapshot(),
		logs:          service.Logs(),
		editInput:     newInputModel(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateValueTableCache()
		m.updateMainCaches()
		return m, nil
	case eventMsg:
		return m.handleEvent(msg.event)
	case readResultMsg:
		if msg.err != nil {
			return m, nil
		}
		m.lastResult = &msg.result
		m.updateValueTableCache()
		m.updateMainCaches()
		return m, nil
	case errorMsg:
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.printInvocation = true
			return m, tea.Quit
		}
		if m.editActive {
			return m.handleEditKey(msg)
		}
		return m.handleKey(msg)
	}

	var cmd tea.Cmd
	m, cmd = m.updateInputs(msg)
	return m, cmd
}

func (m model) handleEvent(event core.Event) (tea.Model, tea.Cmd) {
	switch event.Type {
	case core.EventStatus:
		status, ok := event.Payload.(core.ConnectionStatus)
		if ok {
			m.status = status
			if m.pendingRead != nil && status.Connected {
				req := *m.pendingRead
				m.pendingRead = nil
				m.updateMainCaches()
				return m, m.readCmd(req)
			}
			m.updateMainCaches()
		}
	case core.EventLog:
		entry, ok := event.Payload.(core.LogEntry)
		if ok {
			m.logs = append(m.logs, entry)
			if len(m.logs) > m.logLimit {
				m.logs = m.logs[len(m.logs)-m.logLimit:]
			}
		}
	case core.EventStats:
		stats, ok := event.Payload.(core.Stats)
		if ok {
			m.stats = stats
		}
	case core.EventData, core.EventError:
		result, ok := event.Payload.(core.ReadResult)
		if ok {
			m.lastResult = &result
			m.updateValueTableCache()
			m.updateMainCaches()
		}
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if m.view == viewHelp {
		if key == "esc" || key == "?" {
			m.view = viewMain
		}
		return m, nil
	}
	if m.view == viewLogs {
		if key == "esc" || key == "l" {
			m.view = viewMain
		}
		return m, nil
	}
	if m.view == viewDecoder {
		return m.handleDecoderKeys(key)
	}
	if m.view == viewConnection {
		switch key {
		case "esc", "s":
			m.view = viewMain
			return m, nil
		case "g":
			if m.cfg.Protocol == config.ProtocolRTU {
				m.cfg.Protocol = config.ProtocolTCP
			} else {
				m.cfg.Protocol = config.ProtocolRTU
			}
			m.updateConfig(true)
			return m, nil
		case "p":
			m.toggleProtocol()
			return m, nil
		case "h":
			return m.beginEdit(focusConnHost)
		case "r":
			return m.beginEdit(focusConnPort)
		case "d":
			return m.beginEdit(focusConnDevice)
		case "b":
			return m.beginEdit(focusConnSpeed)
		case "t":
			return m.beginEdit(focusConnDataBits)
		case "y":
			return m.beginEdit(focusConnParity)
		case "w":
			return m.beginEdit(focusConnStopBits)
		case "a":
			return m.openDeviceSelect()
		case "u":
			return m.beginEdit(focusConnTimeout)
		case "i":
			return m.beginEdit(focusConnUnitID)
		default:
			return m, nil
		}
	}
	if m.view == viewFunctionSelect {
		return m.handleFunctionKeys(key)
	}
	if m.view == viewDeviceSelect {
		return m.handleDeviceKeys(key)
	}

	switch key {
	case "q":
		m.printInvocation = true
		return m, tea.Quit
	case "?":
		m.view = viewHelp
		return m, nil
	case "l":
		m.view = viewLogs
		return m, nil
	case "d":
		m.view = viewDecoder
		return m, nil
	case "s":
		m.view = viewConnection
		return m, nil
	case "c":
		if m.status.Connected || m.status.Connecting {
			_ = m.service.Disconnect()
		} else {
			_ = m.service.Connect()
		}
		return m, nil
	case "t":
		m.autoConnect = !m.autoConnect
		m.updateMainCaches()
		return m, nil
	case "r":
		return m.triggerRead()
	case "f":
		m.view = viewFunctionSelect
		m.functionCursor = m.selectedKindIdx
		return m, nil
	case "b":
		m.toggleAddressBase()
		return m, nil
	case "v":
		m.toggleValueBase()
		return m, nil
	case "a":
		return m.beginEdit(focusAddress)
	case "n":
		return m.beginEdit(focusQuantity)
	case "i":
		return m.beginEdit(focusUnitID)
	}

	return m, nil
}

func (m model) updateInputs(msg tea.Msg) (model, tea.Cmd) {
	if !m.editActive {
		return m, nil
	}
	var cmd tea.Cmd
	m.editInput, cmd = m.editInput.Update(msg)
	return m, cmd
}

func (m model) beginEdit(field fieldFocus) (tea.Model, tea.Cmd) {
	if m.editActive {
		return m, nil
	}
	m.editError = ""
	m.addressError = ""
	m.quantityError = ""
	m.unitError = ""
	value := ""
	switch field {
	case focusAddress:
		value = m.addressValue
	case focusQuantity:
		value = m.quantityValue
	case focusUnitID:
		value = m.unitValue
	case focusConnHost:
		value = m.cfg.TCP.Host
	case focusConnPort:
		value = fmt.Sprintf("%d", m.cfg.TCP.Port)
	case focusConnDevice:
		value = m.cfg.Serial.Device
	case focusConnSpeed:
		value = fmt.Sprintf("%d", m.cfg.Serial.Speed)
	case focusConnDataBits:
		value = fmt.Sprintf("%d", m.cfg.Serial.DataBits)
	case focusConnParity:
		value = m.cfg.Serial.Parity
	case focusConnStopBits:
		value = fmt.Sprintf("%d", m.cfg.Serial.StopBits)
	case focusConnTimeout:
		value = fmt.Sprintf("%d", m.cfg.TimeoutMs)
	case focusConnUnitID:
		value = fmt.Sprintf("%d", m.cfg.UnitID)
	default:
		return m, nil
	}
	m.editActive = true
	m.editField = field
	m.editInput = m.editInput.Set(fieldLabel(field), value, editHint(field), editCharLimit(field))
	m.editInput.Focus()
	return m, nil
}

func (m model) handleEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.applyEdit()
	case "esc":
		m.cancelEdit()
		return m, nil
	}
	return m.updateInputs(msg)
}

func (m model) applyEdit() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(m.editInput.Value())
	switch m.editField {
	case focusAddress:
		if parseAddress(value) == nil {
			m.editError = "Invalid address"
			return m, nil
		}
		m.addressValue = value
		m.cfg.ReadAddress = *parseAddress(value)
	case focusQuantity:
		if !validQuantity(value) {
			m.editError = "Quantity must be 1-65535"
			return m, nil
		}
		m.quantityValue = value
		parsed, _ := parseUint16(value)
		m.cfg.ReadQuantity = parsed
	case focusUnitID:
		if !validUnitID(value) {
			m.editError = "Unit ID must be 0-255"
			return m, nil
		}
		m.unitValue = value
		m.cfg.UnitID = parseUint8(value)
		m.updateConfig(false)
	case focusConnHost:
		if strings.TrimSpace(value) == "" {
			m.editError = "Host cannot be empty"
			return m, nil
		}
		m.cfg.TCP.Host = strings.TrimSpace(value)
		m.updateConfig(true)
	case focusConnPort:
		port, ok := parseUint16(value)
		if !ok || port == 0 {
			m.editError = "Port must be 1-65535"
			return m, nil
		}
		m.cfg.TCP.Port = int(port)
		m.updateConfig(true)
	case focusConnDevice:
		if strings.TrimSpace(value) == "" {
			m.editError = "Device cannot be empty"
			return m, nil
		}
		m.cfg.Serial.Device = strings.TrimSpace(value)
		m.updateConfig(true)
	case focusConnSpeed:
		speed, ok := parseUint32(value)
		if !ok || speed == 0 {
			m.editError = "Speed must be > 0"
			return m, nil
		}
		m.cfg.Serial.Speed = uint(speed)
		m.updateConfig(true)
	case focusConnDataBits:
		bits, ok := parseUint32(value)
		if !ok || bits == 0 {
			m.editError = "Data bits must be > 0"
			return m, nil
		}
		m.cfg.Serial.DataBits = uint(bits)
		m.updateConfig(true)
	case focusConnParity:
		normalized := strings.ToLower(strings.TrimSpace(value))
		switch normalized {
		case "none", "even", "odd":
			m.cfg.Serial.Parity = normalized
			m.updateConfig(true)
		default:
			m.editError = "Parity must be none/even/odd"
			return m, nil
		}
	case focusConnStopBits:
		bits, ok := parseUint32(value)
		if !ok || bits == 0 {
			m.editError = "Stop bits must be > 0"
			return m, nil
		}
		m.cfg.Serial.StopBits = uint(bits)
		m.updateConfig(true)
	case focusConnTimeout:
		timeout, ok := parseUint32(value)
		if !ok || timeout == 0 {
			m.editError = "Timeout must be > 0"
			return m, nil
		}
		m.cfg.TimeoutMs = int64(timeout)
		m.updateConfig(true)
	case focusConnUnitID:
		if !validUnitID(value) {
			m.editError = "Unit ID must be 0-255"
			return m, nil
		}
		m.cfg.UnitID = parseUint8(value)
		m.unitValue = value
		m.updateConfig(true)
	}
	m.finishEdit()
	return m, nil
}

func (m *model) cancelEdit() {
	m.editError = ""
	m.finishEdit()
}

func (m *model) finishEdit() {
	m.editActive = false
	m.editField = focusNone
	m.editInput.Blur()
	m.editError = ""
}

func (m model) triggerRead() (tea.Model, tea.Cmd) {
	req, ok := m.buildReadRequest()
	if !ok {
		return m, nil
	}

	if !m.status.Connected && m.autoConnect {
		m.pendingRead = &req
		_ = m.service.Connect()
		return m, nil
	}
	if !m.status.Connected {
		return m, nil
	}

	return m, m.readCmd(req)
}

func (m *model) buildReadRequest() (core.ReadRequest, bool) {
	m.addressError = ""
	m.quantityError = ""
	m.unitError = ""

	addr := parseAddress(m.addressValue)
	if addr == nil {
		m.addressError = "Invalid address"
		return core.ReadRequest{}, false
	}

	quantityValue, err := strconv.Atoi(strings.TrimSpace(m.quantityValue))
	if err != nil || quantityValue < 1 || quantityValue > 0xffff {
		m.quantityError = "Quantity must be >= 1"
		return core.ReadRequest{}, false
	}

	unitValue := strings.TrimSpace(m.unitValue)
	unitID := m.cfg.UnitID
	if unitValue != "" {
		parsed, err := strconv.Atoi(unitValue)
		if err != nil || parsed < 0 || parsed > 255 {
			m.unitError = "Unit ID must be 0-255"
			return core.ReadRequest{}, false
		}
		unitID = uint8(parsed)
	}

	m.cfg.UnitID = unitID

	kind := readKinds[m.selectedKindIdx].kind
	return core.ReadRequest{
		Kind:     kind,
		Address:  *addr,
		Quantity: uint16(quantityValue),
		UnitID:   unitID,
	}, true
}

func (m model) readCmd(req core.ReadRequest) tea.Cmd {
	return func() tea.Msg {
		result, err := m.service.Read(context.Background(), req)
		return readResultMsg{result: result, err: err}
	}
}

func parseAddress(input string) *uint16 {
	trimmed := strings.TrimSpace(strings.ToLower(input))
	if trimmed == "" {
		return nil
	}
	var value int64
	var err error
	if strings.HasPrefix(trimmed, "0x") {
		value, err = strconv.ParseInt(trimmed[2:], 16, 32)
	} else {
		value, err = strconv.ParseInt(trimmed, 10, 32)
	}
	if err != nil || value < 0 || value > 0xffff {
		return nil
	}
	result := uint16(value)
	return &result
}

func validQuantity(value string) bool {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	return err == nil && parsed >= 1 && parsed <= 0xffff
}

func validUnitID(value string) bool {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	return err == nil && parsed >= 0 && parsed <= 255
}

func parseUint16(value string) (uint16, bool) {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed < 0 || parsed > 0xffff {
		return 0, false
	}
	return uint16(parsed), true
}

func parseUint8(value string) uint8 {
	parsed, _ := strconv.Atoi(strings.TrimSpace(value))
	if parsed < 0 {
		parsed = 0
	}
	if parsed > 255 {
		parsed = 255
	}
	return uint8(parsed)
}

func parseUint32(value string) (uint32, bool) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 32)
	if err != nil {
		return 0, false
	}
	return uint32(parsed), true
}

func editCharLimit(field fieldFocus) int {
	switch field {
	case focusAddress:
		return 6
	case focusQuantity:
		return 5
	case focusUnitID:
		return 3
	case focusConnDevice:
		return 128
	case focusConnHost:
		return 64
	default:
		return 16
	}
}

func editHint(field fieldFocus) string {
	switch field {
	case focusConnParity:
		return ""
	case focusConnHost:
		return ""
	case focusConnPort:
		return ""
	case focusConnDevice:
		return ""
	case focusConnUnitID, focusUnitID:
		return ""
	case focusConnTimeout:
		return ""
	default:
		return ""
	}
}

func (m *model) updateConfig(reconnect bool) {
	m.service.UpdateConfig(m.cfg)
	m.updateMainCaches()
	if !reconnect {
		return
	}
	if m.status.Connected || m.status.Connecting {
		_ = m.service.Disconnect()
		_ = m.service.Connect()
	}
}

func (m *model) updateValueTableCache() {
	if m.width <= 0 || m.lastResult == nil {
		m.valueTableCache = ""
		m.valueTableKey = ""
		return
	}
	key := m.valueTableCacheKey()
	if key == m.valueTableKey {
		return
	}
	tmp := *m
	tmp.valueTableCache = ""
	m.valueTableCache = buildValueTable(tmp, m.width)
	m.valueTableKey = key
}

func (m *model) updateMainCaches() {
	if m.width <= 0 {
		m.connectionBoxCache = ""
		m.resultBoxCache = ""
		m.connectionBoxKey = ""
		m.resultBoxKey = ""
		return
	}
	connKey := fmt.Sprintf("w=%d|status=%s|err=%s|proto=%s|target=%s|unit=%d|timeout=%d|auto=%t",
		m.width,
		connectionLabel(m.status),
		m.status.LastError,
		m.cfg.Protocol,
		connectionTarget(*m),
		m.cfg.UnitID,
		m.cfg.TimeoutMs,
		m.autoConnect,
	)
	if connKey != m.connectionBoxKey {
		m.connectionBoxCache = renderBox("connection [s]ettings", renderConnectionSummary(*m), m.width)
		m.connectionBoxKey = connKey
	}

	resultKey := fmt.Sprintf("w=%d|result=%t|err=%s|done=%d|table=%s",
		m.width,
		m.lastResult != nil,
		func() string {
			if m.lastResult == nil {
				return ""
			}
			return m.lastResult.ErrorMessage
		}(),
		func() int64 {
			if m.lastResult == nil {
				return 0
			}
			return m.lastResult.CompletedAt.UnixNano()
		}(),
		m.valueTableKey,
	)
	if resultKey != m.resultBoxKey {
		m.resultBoxCache = renderBox("values", renderResultPanel(*m, m.width), m.width)
		m.resultBoxKey = resultKey
	}
}

func (m *model) valueTableCacheKey() string {
	result := m.lastResult
	if result == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "w=%d|addr=%d|qty=%d|done=%d|reg=%d|bool=%d|fmt=%d|base=%d|",
		m.width,
		result.Address,
		result.Quantity,
		result.CompletedAt.UnixNano(),
		len(result.RegValues),
		len(result.BoolValues),
		m.cfg.AddressFormat,
		m.cfg.ValueBase,
	)
	for _, decoder := range m.cfg.Decoders {
		fmt.Fprintf(&b, "%s:%t:%s:%s|", decoder.Type, decoder.Enabled, decoder.Endianness, decoder.WordOrder)
	}
	return b.String()
}

func (m model) handleDecoderKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc", "d":
		m.view = viewMain
		return m, nil
	case "up", "k":
		if m.decoderCursor > 0 {
			m.decoderCursor--
		}
		return m, nil
	case "down", "j":
		if m.decoderCursor < len(m.cfg.Decoders)-1 {
			m.decoderCursor++
		}
		return m, nil
	case " ":
		m.toggleDecoderEnabled(m.decoderCursor)
		m.updateValueTableCache()
		m.updateMainCaches()
		return m, nil
	case "e":
		m.toggleDecoderEndianness(m.decoderCursor)
		m.updateValueTableCache()
		m.updateMainCaches()
		return m, nil
	case "w":
		m.toggleDecoderWordOrder(m.decoderCursor)
		m.updateValueTableCache()
		m.updateMainCaches()
		return m, nil
	}
	return m, nil
}

func (m model) handleFunctionKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc", "f":
		m.view = viewMain
		return m, nil
	case "1", "2", "3", "4":
		idx := int(key[0] - '1')
		if idx >= 0 && idx < len(readKinds) {
			m.selectedKindIdx = idx
			m.cfg.ReadKind = string(readKinds[idx].kind)
			m.view = viewMain
		}
		return m, nil
	case "up", "k":
		if m.functionCursor > 0 {
			m.functionCursor--
		}
		return m, nil
	case "down", "j":
		if m.functionCursor < len(readKinds)-1 {
			m.functionCursor++
		}
		return m, nil
	case "enter":
		m.selectedKindIdx = m.functionCursor
		m.cfg.ReadKind = string(readKinds[m.functionCursor].kind)
		m.view = viewMain
		return m, nil
	}
	return m, nil
}

func (m model) handleDeviceKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc", "a":
		m.view = viewConnection
		return m, nil
	case "up", "k":
		if m.deviceCursor > 0 {
			m.deviceCursor--
		}
		return m, nil
	case "down", "j":
		if m.deviceCursor < len(m.deviceList)-1 {
			m.deviceCursor++
		}
		return m, nil
	case "enter":
		if m.deviceCursor >= 0 && m.deviceCursor < len(m.deviceList) {
			m.cfg.Serial.Device = m.deviceList[m.deviceCursor]
			m.updateConfig(true)
		}
		m.view = viewConnection
		return m, nil
	}
	return m, nil
}

func (m *model) toggleDecoderEnabled(idx int) {
	if idx < 0 || idx >= len(m.cfg.Decoders) {
		return
	}
	dec := m.cfg.Decoders[idx]
	dec.Enabled = !dec.Enabled
	m.cfg.Decoders[idx] = dec
	m.service.UpdateConfig(m.cfg)
}

func (m *model) toggleDecoderEndianness(idx int) {
	if idx < 0 || idx >= len(m.cfg.Decoders) {
		return
	}
	dec := m.cfg.Decoders[idx]
	if dec.Endianness == config.EndianLittle {
		dec.Endianness = config.EndianBig
	} else {
		dec.Endianness = config.EndianLittle
	}
	m.cfg.Decoders[idx] = dec
	m.service.UpdateConfig(m.cfg)
}

func (m *model) toggleDecoderWordOrder(idx int) {
	if idx < 0 || idx >= len(m.cfg.Decoders) {
		return
	}
	dec := m.cfg.Decoders[idx]
	if dec.WordOrder == config.WordLowFirst {
		dec.WordOrder = config.WordHighFirst
	} else {
		dec.WordOrder = config.WordLowFirst
	}
	m.cfg.Decoders[idx] = dec
	m.service.UpdateConfig(m.cfg)
}

func (m model) View() string {
	if m.width > 0 && (m.width < minWidth || m.height < minHeight) {
		return renderTooSmall(m.width, m.height)
	}
	return renderBase(m)
}

func renderBase(m model) string {
	switch m.view {
	case viewHelp:
		return renderHelp(m)
	case viewLogs:
		return renderLogs(m)
	case viewDecoder:
		return renderDecoder(m)
	case viewConnection:
		return renderConnectionDetails(m)
	case viewFunctionSelect:
		return renderFunctionSelect(m)
	case viewDeviceSelect:
		return renderDeviceSelect(m)
	default:
		return renderMain(m)
	}
}

func (m *model) openDeviceSelect() (tea.Model, tea.Cmd) {
	m.deviceList = listSerialDevices()
	m.deviceCursor = 0
	m.view = viewDeviceSelect
	return *m, nil
}

func (m *model) toggleProtocol() {
	if m.cfg.Protocol == config.ProtocolRTU {
		m.cfg.Protocol = config.ProtocolTCP
	} else {
		m.cfg.Protocol = config.ProtocolRTU
	}
	m.updateConfig(true)
}

func (m *model) toggleAddressBase() {
	if m.cfg.AddressBase == config.AddressBaseZero {
		m.cfg.AddressBase = config.AddressBaseOne
	} else {
		m.cfg.AddressBase = config.AddressBaseZero
	}
	m.updateConfig(false)
	m.updateValueTableCache()
	m.updateMainCaches()
}

func (m *model) toggleValueBase() {
	if m.cfg.ValueBase == config.ValueBaseHex {
		m.cfg.ValueBase = config.ValueBaseDec
	} else {
		m.cfg.ValueBase = config.ValueBaseHex
	}
	m.updateConfig(false)
	m.updateValueTableCache()
	m.updateMainCaches()
}
