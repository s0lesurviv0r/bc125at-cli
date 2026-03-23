package formats

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/jzelek/bc125-cli/scanner"
)

// ScannerData is the top-level JSON structure for a full scanner backup.
type ScannerData struct {
	Model    string            `json:"model"`
	Firmware string            `json:"firmware"`
	Settings *scanner.Settings `json:"settings"`
	Channels []*scanner.Channel `json:"channels"`
}

// WriteJSON encodes scanner data as pretty-printed JSON to w.
func WriteJSON(w io.Writer, data *ScannerData) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// ReadJSON decodes scanner data from JSON in r.
func ReadJSON(r io.Reader) (*ScannerData, error) {
	var data ScannerData
	dec := json.NewDecoder(r)
	if err := dec.Decode(&data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return &data, nil
}

// ValidateJSONData validates all settings and channels in a ScannerData struct.
func ValidateJSONData(data *ScannerData) []error {
	var errs []error

	if data.Settings != nil {
		if err := data.Settings.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("settings: %w", err))
		}
	}

	for _, ch := range data.Channels {
		if err := ch.Validate(); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
