package tui

import (
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"gomodmaster/internal/config"
	"gomodmaster/internal/core"

	"github.com/charmbracelet/lipgloss"
)

var (
	borderStyle   = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)
	titleStyle    = lipgloss.NewStyle().Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	activeStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
)

func renderTooSmall(width, height int) string {
	return fmt.Sprintf("Terminal too small (min %dx%d). Current: %dx%d\n\nResize to continue.\n", minWidth, minHeight, width, height)
}

func renderMain(m model) string {
	if m.width == 0 {
		return "Loading..."
	}

	connectionBar := m.connectionBoxCache
	if connectionBar == "" {
		connectionBar = renderBox("connection [s]ettings", renderConnectionSummary(m), m.width)
	}
	readBox := renderBox("read", renderReadPanel(m), m.width)
	resultBox := m.resultBoxCache
	if resultBox == "" {
		resultBox = renderBox("values", renderResultPanel(m, m.width), m.width)
	}

	body := lipgloss.JoinVertical(lipgloss.Left, connectionBar, readBox, resultBox)
	return renderScreen(m, body)
}

func renderConnectionSummary(m model) string {
	parts := []string{
		fmt.Sprintf("%s", connectionLabel(m.status)),
		fmt.Sprintf("%s", strings.ToUpper(string(m.cfg.Protocol))),
		connectionTarget(m),
		fmt.Sprintf("unit %d", m.cfg.UnitID),
		fmt.Sprintf("timeout %dms", m.cfg.TimeoutMs),
		fmt.Sprintf("auto %s", onOff(m.autoConnect)),
	}
	if m.status.LastError != "" {
		parts = append(parts, errorStyle.Render(m.status.LastError))
	}
	return strings.Join(parts, " | ")
}

func renderReadPanel(m model) string {
	kind := readKinds[m.selectedKindIdx]
	line1 := strings.Join([]string{
		fmt.Sprintf("[f]unction: %s (%s)", kind.label, kind.code),
		renderFixedField(m, focusAddress, "[a]ddress", m.addressValue, 6),
		renderFixedField(m, focusQuantity, "cou[n]t", m.quantityValue, 5),
	}, " | ")
	line2 := strings.Join([]string{
		renderEditableField(m, focusUnitID, "unit-[i]d", m.unitValue),
		fmt.Sprintf("address [b]ase: %s", formatAddressBase(m.cfg.AddressBase)),
		fmt.Sprintf("value [v]ase: %s", formatBase(m.cfg.ValueBase)),
	}, " | ")
	return strings.Join([]string{line1, line2}, "\n")
}

func renderEditableField(m model, field fieldFocus, label, value string) string {
	if m.editActive && m.editField == field {
		return fmt.Sprintf("%s: %s", label, activeStyle.Render(m.editInput.input.View()))
	}
	if value == "" {
		value = "-"
	}
	return fmt.Sprintf("%s: %s", label, value)
}

func renderFixedField(m model, field fieldFocus, label, value string, width int) string {
	if m.editActive && m.editField == field {
		return fmt.Sprintf("%s: %s", label, activeStyle.Render(fixedWidth(m.editInput.input.View(), width)))
	}
	if value == "" {
		value = "-"
	}
	return fmt.Sprintf("%s: %s", label, fixedWidth(value, width))
}

func renderConnField(m model, field fieldFocus, label, value string) string {
	return renderEditableField(m, field, label, value)
}

func renderResultPanel(m model, panelWidth int) string {
	lines := []string{titleStyle.Render("Results")}
	if m.lastResult == nil {
		lines = append(lines, dimStyle.Render("No results yet"))
		return strings.Join(lines, "\n")
	}

	result := m.lastResult
	lines = append(lines, fmt.Sprintf("Completed: %s", formatTime(result.CompletedAt)))
	lines = append(lines, fmt.Sprintf("Latency: %d ms", result.LatencyMs))
	if result.ErrorMessage != "" {
		lines = append(lines, errorStyle.Render(fmt.Sprintf("Error: %s", result.ErrorMessage)))
		return strings.Join(lines, "\n")
	}

	lines = append(lines, "")
	lines = append(lines, renderValueTable(m, panelWidth))

	if len(result.Decoded) > 0 {
		lines = append(lines, "")
		lines = append(lines, titleStyle.Render("Decoded"))
		for _, dec := range result.Decoded {
			lines = append(lines, fmt.Sprintf("- %s: %v", dec.Type, dec.Value))
		}
	}

	return strings.Join(lines, "\n")
}

func renderValueTable(m model, panelWidth int) string {
	if m.lastResult == nil {
		return ""
	}
	if m.valueTableCache != "" && m.width == panelWidth {
		return m.valueTableCache
	}
	width := panelWidth - 4
	if width < 24 {
		width = 24
	}

	labelWidth := 12
	cellWidth := 8
	columns := (width - labelWidth) / cellWidth
	if columns < 1 {
		columns = 1
	}
	if columns > 8 {
		columns = 8
	}

	rows := buildRows(m, columns)
	head := fmt.Sprintf("%s %s", col("Base", labelWidth), renderHeaderCells(columns, cellWidth))
	sep := strings.Repeat("-", width)
	lines := []string{head, sep}
	for _, row := range rows {
		lines = append(lines, renderRow(row, labelWidth, cellWidth))
	}
	return strings.Join(lines, "\n")
}

func buildValueTable(m model, panelWidth int) string {
	return renderValueTable(m, panelWidth)
}

func renderLogs(m model) string {
	box := renderBox("logs", renderLogBody(m), m.width)
	return renderScreen(m, box)
}

func renderLogBody(m model) string {
	var b strings.Builder
	visible := m.height - 6
	if visible < 1 {
		visible = len(m.logs)
	}
	start := 0
	if visible > 0 && len(m.logs) > visible {
		start = len(m.logs) - visible
	}

	for _, entry := range m.logs[start:] {
		b.WriteString(fmt.Sprintf("%s %s\n", formatTime(entry.Time), entry.Message))
	}
	return b.String()
}

func renderHelp(m model) string {
	lines := []string{
		"Help (press [?] or [esc] to return)",
		"",
		"Editing:",
		"  Press field key to edit",
		"  [enter] accept, [esc] discard",
		"",
		"Fields:",
		"  [a] Address",
		"  qua[n]tity",
		"  unit-[i]d",
		"  address [b]ase",
		"  value [v]ase",
		"",
		"Actions:",
		"  [f] Function select",
		"  [r] Read now",
		"  [c] Connect/disconnect",
		"  [s] Connection settings",
		"  [d] Decoder settings",
		"  [l] Raw logs",
		"  [q] Quit (prints invocation)",
	}
	box := renderBox("help", strings.Join(lines, "\n"), m.width)
	return renderScreen(m, box)
}

func renderDecoder(m model) string {
	var b strings.Builder
	b.WriteString("Keys: [j]/[k] move  [space] toggle  [e] endianness  [w] word order\n\n")

	for idx, dec := range m.cfg.Decoders {
		cursor := " "
		if idx == m.decoderCursor {
			cursor = ">"
		}
		enabled := " "
		if dec.Enabled {
			enabled = "x"
		}
		line := fmt.Sprintf("%s [%s] %-8s endian=%-6s word=%s", cursor, enabled, dec.Type, string(dec.Endianness), string(dec.WordOrder))
		if idx == m.decoderCursor {
			line = selectedStyle.Render(line)
		}
		b.WriteString(line + "\n")
	}

	box := renderBox("decoders", b.String(), m.width)
	return renderScreen(m, box)
}

func renderConnectionDetails(m model) string {
	lines := []string{
		fmt.Sprintf("[p]rotocol: %s", strings.ToUpper(string(m.cfg.Protocol))),
	}
	if m.cfg.Protocol == config.ProtocolTCP {
		lines = append(lines,
			renderConnField(m, focusConnHost, "[h]ost", m.cfg.TCP.Host),
			renderConnField(m, focusConnPort, "po[r]t", fmt.Sprintf("%d", m.cfg.TCP.Port)),
		)
	} else {
		lines = append(lines,
			renderConnField(m, focusConnDevice, "[d]evice", m.cfg.Serial.Device),
			dimStyle.Render("[a]vailable devices"),
			renderConnField(m, focusConnSpeed, "[b]aud", fmt.Sprintf("%d", m.cfg.Serial.Speed)),
			renderConnField(m, focusConnDataBits, "da[t]a bits", fmt.Sprintf("%d", m.cfg.Serial.DataBits)),
			renderConnField(m, focusConnParity, "parit[y]", m.cfg.Serial.Parity),
			renderConnField(m, focusConnStopBits, "stop [w]bits", fmt.Sprintf("%d", m.cfg.Serial.StopBits)),
		)
		lines = append(lines, dimStyle.Render("type a custom path to override"))
	}
	lines = append(lines,
		renderConnField(m, focusConnTimeout, "timeo[u]t", fmt.Sprintf("%d ms", m.cfg.TimeoutMs)),
		renderConnField(m, focusConnUnitID, "unit-[i]d", fmt.Sprintf("%d", m.cfg.UnitID)),
	)
	if m.status.LastError != "" {
		lines = append(lines, "", errorStyle.Render(fmt.Sprintf("Last error: %s", m.status.LastError)))
	}
	box := renderBox("connection settings", strings.Join(lines, "\n"), m.width)
	return renderScreen(m, box)
}

func renderFunctionSelect(m model) string {
	lines := []string{"Use [j]/[k] to move, [enter] to select, [esc] to cancel", ""}
	for idx, option := range readKinds {
		label := fmt.Sprintf("%s (%s)", option.label, option.code)
		if idx == m.functionCursor {
			label = selectedStyle.Render("> " + label)
		} else {
			label = "  " + label
		}
		lines = append(lines, label)
	}
	box := renderBox("function", strings.Join(lines, "\n"), m.width)
	return renderScreen(m, box)
}

func renderScreen(m model, content string) string {
	status := renderStatusClue(m)
	commands := renderCommandsLine(m)
	if m.editActive {
		return strings.Join([]string{content, status, commands}, "\n")
	}
	body := padToHeight(content, m.height-2)
	return strings.Join([]string{body, status, commands}, "\n")
}

func renderStatusClue(m model) string {
	status := fmt.Sprintf("Status: %s", connectionLabel(m.status))
	if m.status.LastError != "" {
		status = fmt.Sprintf("%s | %s", status, m.status.LastError)
	}
	clue := ""
	if m.editActive {
		clue = fmt.Sprintf("Editing %s: enter=save  esc=cancel", fieldLabel(m.editField))
	}
	if err := footerError(m); err != "" {
		clue = errorStyle.Render(err)
	}
	if clue == "" {
		return status
	}
	return fmt.Sprintf("%s | %s", status, clue)
}

func renderCommandsLine(m model) string {
	if m.editActive {
		return ""
	}
	switch m.view {
	case viewConnection:
		return "[esc] back"
	case viewFunctionSelect:
		return "[enter] select  [esc] back"
	case viewDeviceSelect:
		return "[enter] select  [esc] back"
	case viewDecoder:
		return "[space] toggle  [e] endianness  [w] word order  [j]/[k] move  [esc] back"
	case viewLogs, viewHelp:
		return "[esc] back"
	default:
		return "[r] read  [c] connect  [d] decoders  [l] logs  [?] help  [q] quit"
	}
}

func renderDeviceSelect(m model) string {
	lines := []string{"Use [j]/[k] to move, [enter] to select, [esc] to cancel", ""}
	if len(m.deviceList) == 0 {
		lines = append(lines, dimStyle.Render("No devices detected"))
	} else {
		for idx, device := range m.deviceList {
			label := device
			if idx == m.deviceCursor {
				label = selectedStyle.Render("> " + label)
			} else {
				label = "  " + label
			}
			lines = append(lines, label)
		}
	}
	box := renderBox("available devices", strings.Join(lines, "\n"), m.width)
	return renderScreen(m, box)
}

func formatBase(base config.ValueBase) string {
	if base == config.ValueBaseHex {
		return "hex"
	}
	return "dec"
}

func formatAddressBase(base config.AddressBase) string {
	if base == config.AddressBaseOne {
		return "1-based"
	}
	return "0-based"
}

func footerError(m model) string {
	if m.editError != "" {
		return m.editError
	}
	if m.addressError != "" {
		return m.addressError
	}
	if m.quantityError != "" {
		return m.quantityError
	}
	if m.unitError != "" {
		return m.unitError
	}
	return ""
}

func fieldLabel(field fieldFocus) string {
	switch field {
	case focusAddress:
		return "address"
	case focusQuantity:
		return "quantity"
	case focusUnitID:
		return "unit-id"
	case focusConnHost:
		return "host"
	case focusConnPort:
		return "port"
	case focusConnDevice:
		return "device"
	case focusConnSpeed:
		return "baud"
	case focusConnDataBits:
		return "data bits"
	case focusConnParity:
		return "parity"
	case focusConnStopBits:
		return "stop bits"
	case focusConnTimeout:
		return "timeout"
	case focusConnUnitID:
		return "unit-id"
	default:
		return "field"
	}
}

func renderBox(title, content string, width int) string {
	if width < 4 {
		return content
	}
	border := lipgloss.NormalBorder()
	innerWidth := width - 2
	label := fmt.Sprintf(" %s ", title)
	label = clamp(label, innerWidth)
	top := border.TopLeft + label + strings.Repeat(border.Top, innerWidth-lipgloss.Width(label)) + border.TopRight

	lineWidth := innerWidth - 2
	if lineWidth < 1 {
		lineWidth = 1
	}
	lineStyle := lipgloss.NewStyle().Width(lineWidth).MaxWidth(lineWidth)

	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}
	var body []string
	for _, line := range lines {
		body = append(body, border.Left+" "+lineStyle.Render(line)+" "+border.Right)
	}

	bottom := border.BottomLeft + strings.Repeat(border.Bottom, innerWidth) + border.BottomRight
	return strings.Join(append(append([]string{top}, body...), bottom), "\n")
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "--:--:--"
	}
	return t.Format("15:04:05")
}

func col(value string, width int) string {
	return lipgloss.NewStyle().Width(width).MaxWidth(width).MaxHeight(1).Render(value)
}

func fixedWidth(value string, width int) string {
	return lipgloss.NewStyle().Width(width).MaxWidth(width).MaxHeight(1).Render(value)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type tableCell struct {
	value   string
	colSpan int
}

type tableRow struct {
	label string
	cells []tableCell
}

func buildRows(m model, columns int) []tableRow {
	result := m.lastResult
	if result == nil {
		return nil
	}
	values := result.RegValues
	if len(values) == 0 && len(result.BoolValues) > 0 {
		values = make([]uint16, 0, len(result.BoolValues))
		for _, value := range result.BoolValues {
			if value {
				values = append(values, 1)
			} else {
				values = append(values, 0)
			}
		}
	}

	rows := []tableRow{}
	for offset := 0; offset < len(values); offset += columns {
		slice := values[offset:min(offset+columns, len(values))]
		baseAddress := int(result.Address) + offset
		row := tableRow{
			label: formatAddress(baseAddress, m.cfg.AddressFormat),
			cells: make([]tableCell, 0, len(slice)),
		}
		for _, value := range slice {
			row.cells = append(row.cells, tableCell{value: formatValue(int(value), m.cfg.ValueBase), colSpan: 1})
		}
		rows = append(rows, row)

		if len(result.RegValues) == 0 {
			continue
		}

		decoders := enabledDecoders(m.cfg.Decoders)
		if len(decoders) == 0 {
			continue
		}
		regSlice := result.RegValues[offset:min(offset+columns, len(result.RegValues))]
		for _, decoder := range decoders {
			rows = append(rows, tableRow{
				label: "↳ " + string(decoder.Type),
				cells: decodeRegisters(regSlice, decoder, columns),
			})
		}
	}

	return rows
}

func enabledDecoders(decoders []config.DecoderConfig) []config.DecoderConfig {
	order := map[config.DecoderType]int{
		config.DecoderUint16:  0,
		config.DecoderInt16:   1,
		config.DecoderUint32:  2,
		config.DecoderInt32:   3,
		config.DecoderFloat32: 4,
	}
	enabled := []config.DecoderConfig{}
	for _, decoder := range decoders {
		if decoder.Enabled {
			enabled = append(enabled, decoder)
		}
	}
	sort.Slice(enabled, func(i, j int) bool {
		return order[enabled[i].Type] < order[enabled[j].Type]
	})
	return enabled
}

func decodeRegisters(regs []uint16, decoder config.DecoderConfig, columns int) []tableCell {
	switch decoder.Type {
	case config.DecoderUint16:
		return mapCells(regs, func(v uint16) string { return formatValue(int(v), 10) })
	case config.DecoderInt16:
		return mapCells(regs, func(v uint16) string { return fmt.Sprintf("%d", toInt16(v)) })
	case config.DecoderUint32, config.DecoderInt32, config.DecoderFloat32:
		cells := []tableCell{}
		for i := 0; i+1 < len(regs); i += 2 {
			high := regs[i]
			low := regs[i+1]
			first, second := high, low
			if decoder.WordOrder == config.WordLowFirst {
				first, second = low, high
			}
			var bytes []byte
			if decoder.Endianness == config.EndianLittle {
				bytes = []byte{byte(first & 0xff), byte(first >> 8), byte(second & 0xff), byte(second >> 8)}
			} else {
				bytes = []byte{byte(first >> 8), byte(first & 0xff), byte(second >> 8), byte(second & 0xff)}
			}
			value := decodeBytes(bytes, decoder.Type)
			cells = append(cells, tableCell{value: value, colSpan: 2})
		}
		if len(cells) == 0 {
			return []tableCell{{value: "—", colSpan: columns}}
		}
		return cells
	default:
		return mapCells(regs, func(v uint16) string { return fmt.Sprintf("%d", v) })
	}
}

func decodeBytes(bytes []byte, decoderType config.DecoderType) string {
	if len(bytes) < 4 {
		return "—"
	}
	value := binary.BigEndian.Uint32(bytes)
	switch decoderType {
	case config.DecoderUint32:
		return fmt.Sprintf("%d", value)
	case config.DecoderInt32:
		return fmt.Sprintf("%d", int32(value))
	case config.DecoderFloat32:
		return formatFloat(math.Float32frombits(value))
	default:
		return fmt.Sprintf("%d", value)
	}
}

func mapCells(regs []uint16, fn func(uint16) string) []tableCell {
	cells := make([]tableCell, 0, len(regs))
	for _, value := range regs {
		cells = append(cells, tableCell{value: fn(value), colSpan: 1})
	}
	return cells
}

func renderHeaderCells(columns, cellWidth int) string {
	cells := make([]string, 0, columns)
	for i := 0; i < columns; i++ {
		cells = append(cells, col(fmt.Sprintf("+%d", i), cellWidth))
	}
	return strings.Join(cells, " ")
}

func renderRow(row tableRow, labelWidth, cellWidth int) string {
	cells := make([]string, 0, len(row.cells))
	for _, cell := range row.cells {
		width := cellWidth * cell.colSpan
		cells = append(cells, col(cell.value, width))
	}
	return fmt.Sprintf("%s %s", col(row.label, labelWidth), strings.Join(cells, " "))
}

func formatAddress(address int, format config.ValueBase) string {
	if format == config.ValueBaseHex {
		return fmt.Sprintf("0x%04x", address)
	}
	return fmt.Sprintf("%d", address)
}

func formatValue(value int, base config.ValueBase) string {
	if base == config.ValueBaseHex {
		return fmt.Sprintf("0x%04x", value)
	}
	return fmt.Sprintf("%d", value)
}

func toInt16(value uint16) int16 {
	masked := int32(value & 0xffff)
	if masked&0x8000 != 0 {
		return int16(masked - 0x10000)
	}
	return int16(masked)
}

func formatFloat(value float32) string {
	if !isFinite(value) {
		return fmt.Sprintf("%v", value)
	}
	if math.Abs(float64(value)) >= 1e21 {
		return fmt.Sprintf("%.3e", value)
	}
	return fmt.Sprintf("%.3f", value)
}

func isFinite(value float32) bool {
	return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0)
}

func clamp(value string, width int) string {
	if width <= 0 {
		return ""
	}
	return lipgloss.NewStyle().MaxWidth(width).Render(value)
}

func clampInt(value, max int) int {
	if value > max {
		return max
	}
	return value
}

func padToHeight(content string, height int) string {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if height <= 0 {
		return content
	}
	if len(lines) >= height {
		return strings.Join(lines[:height], "\n")
	}
	padding := make([]string, height-len(lines))
	return strings.Join(append(lines, padding...), "\n")
}

func connectionLabel(status core.ConnectionStatus) string {
	if status.Connecting {
		return "CONNECTING"
	}
	if status.Connected {
		return "CONNECTED"
	}
	return "DISCONNECTED"
}

func connectionTarget(m model) string {
	if m.cfg.Protocol == config.ProtocolRTU {
		return m.cfg.Serial.Device
	}
	return fmt.Sprintf("%s:%d", m.cfg.TCP.Host, m.cfg.TCP.Port)
}

func onOff(value bool) string {
	if value {
		return "ON"
	}
	return "OFF"
}
