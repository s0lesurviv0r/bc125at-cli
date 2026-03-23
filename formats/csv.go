package formats

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/jzelek/bc125-cli/scanner"
)

// CSV column headers
var csvHeaders = []string{
	"index", "name", "frequency_mhz", "modulation", "ctcss_dcs", "delay", "locked_out", "priority",
}

// WriteCSV writes channel data to w in CSV format.
func WriteCSV(w io.Writer, channels []*scanner.Channel) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(csvHeaders); err != nil {
		return err
	}
	for _, ch := range channels {
		if err := cw.Write(channelToCSVRow(ch)); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// channelToCSVRow converts a Channel to a CSV row slice.
func channelToCSVRow(ch *scanner.Channel) []string {
	freqStr := ""
	if !ch.IsEmpty() {
		freqStr = strconv.FormatFloat(ch.FrequencyMHz, 'f', 4, 64)
	}
	lockedStr := "0"
	if ch.LockedOut {
		lockedStr = "1"
	}
	priorityStr := "0"
	if ch.Priority {
		priorityStr = "1"
	}
	toneName, _ := scanner.ToneToHuman(ch.CTCSS)
	return []string{
		strconv.Itoa(ch.Index),
		ch.Name,
		freqStr,
		string(ch.Modulation),
		toneName,
		strconv.Itoa(ch.Delay),
		lockedStr,
		priorityStr,
	}
}

// ReadCSV reads channels from r in CSV format.
func ReadCSV(r io.Reader) ([]*scanner.Channel, error) {
	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true

	// Read header row
	headers, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("read CSV header: %w", err)
	}
	// Build column index map (case-insensitive)
	colIdx := make(map[string]int)
	for i, h := range headers {
		colIdx[strings.ToLower(strings.TrimSpace(h))] = i
	}

	// Validate required columns
	required := []string{"index", "frequency_mhz", "modulation", "ctcss_dcs", "delay"}
	for _, col := range required {
		if _, ok := colIdx[col]; !ok {
			return nil, fmt.Errorf("CSV missing required column %q", col)
		}
	}

	var channels []*scanner.Channel
	lineNum := 1
	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		lineNum++
		if err != nil {
			return nil, fmt.Errorf("CSV line %d: %w", lineNum, err)
		}

		ch, err := csvRowToChannel(row, colIdx, lineNum)
		if err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, nil
}

func csvRowToChannel(row []string, colIdx map[string]int, lineNum int) (*scanner.Channel, error) {
	get := func(col string) string {
		idx, ok := colIdx[col]
		if !ok || idx >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[idx])
	}

	idx, err := strconv.Atoi(get("index"))
	if err != nil {
		return nil, fmt.Errorf("CSV line %d: invalid index %q", lineNum, get("index"))
	}

	name := get("name")

	var freqMHz float64
	freqStr := get("frequency_mhz")
	if freqStr != "" {
		freqMHz, err = strconv.ParseFloat(freqStr, 64)
		if err != nil {
			return nil, fmt.Errorf("CSV line %d: invalid frequency %q", lineNum, freqStr)
		}
	}

	mod := scanner.Modulation(strings.ToUpper(get("modulation")))
	if mod == "" {
		mod = scanner.ModFM
	}

	toneStr := get("ctcss_dcs")
	ctcss := 0
	if toneStr != "" {
		ctcss, err = scanner.ToneToInternal(toneStr)
		if err != nil {
			return nil, fmt.Errorf("CSV line %d: %w", lineNum, err)
		}
	}

	delay := 2
	delayStr := get("delay")
	if delayStr != "" {
		delay, err = strconv.Atoi(delayStr)
		if err != nil {
			return nil, fmt.Errorf("CSV line %d: invalid delay %q", lineNum, delayStr)
		}
	}

	lockedOut := false
	if lo := get("locked_out"); lo == "1" || strings.EqualFold(lo, "true") {
		lockedOut = true
	}

	priority := false
	if p := get("priority"); p == "1" || strings.EqualFold(p, "true") {
		priority = true
	}

	return &scanner.Channel{
		Index:        idx,
		Name:         name,
		FrequencyMHz: freqMHz,
		Modulation:   mod,
		CTCSS:        ctcss,
		Delay:        delay,
		LockedOut:    lockedOut,
		Priority:     priority,
	}, nil
}

// ValidateCSVChannels validates all channels in a CSV-sourced slice and returns any errors.
func ValidateCSVChannels(channels []*scanner.Channel) []error {
	var errs []error
	for _, ch := range channels {
		if err := ch.Validate(); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
