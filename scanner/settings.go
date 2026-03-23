package scanner

import "fmt"

// BacklightMode represents the scanner's backlight setting.
type BacklightMode string

const (
	BacklightAlwaysOn BacklightMode = "AO"
	Backlight10Sec    BacklightMode = "10"
	Backlight30Sec    BacklightMode = "30"
	BacklightSquelch  BacklightMode = "SQ"
)

var validBacklightModes = map[BacklightMode]bool{
	BacklightAlwaysOn: true,
	Backlight10Sec:    true,
	Backlight30Sec:    true,
	BacklightSquelch:  true,
}

// PriorityMode represents the scanner's priority scan mode.
type PriorityMode int

const (
	PriorityOff  PriorityMode = 0
	PriorityOn   PriorityMode = 1
	PriorityPlus PriorityMode = 2
)

// Settings holds all non-channel scanner settings.
type Settings struct {
	// BLT - backlight mode: "AO", "10", "30", "SQ"
	Backlight BacklightMode `json:"backlight"`
	// BSV - battery charge timer in hours (0=off, 1-16)
	BatterySave int `json:"battery_save"`
	// KBP - key beep level (0-99) and key lock (true/false)
	KeyBeep int  `json:"key_beep"`
	KeyLock bool `json:"key_lock"`
	// PRI - priority mode (0=off, 1=on, 2=plus)
	Priority PriorityMode `json:"priority"`
	// VOL - volume (0-15)
	Volume int `json:"volume"`
	// SQL - squelch (0-15)
	Squelch int `json:"squelch"`
	// CNT - contrast (1-15)
	Contrast int `json:"contrast"`
	// SCG - active scan banks (10-char string, each char '0' or '1')
	ScanBanks string `json:"scan_banks"`
	// WXS - weather alert (0=off, 1=on)
	WeatherScan int `json:"weather_scan"`
}

// Validate checks that all settings are within valid ranges.
func (s *Settings) Validate() error {
	if !validBacklightModes[s.Backlight] {
		return fmt.Errorf("invalid backlight mode %q (must be AO, 10, 30, or SQ)", s.Backlight)
	}
	if s.BatterySave < 0 || s.BatterySave > 16 {
		return fmt.Errorf("battery_save must be 0-16, got %d", s.BatterySave)
	}
	if s.KeyBeep < 0 || s.KeyBeep > 99 {
		return fmt.Errorf("key_beep must be 0-99, got %d", s.KeyBeep)
	}
	if s.Priority < 0 || s.Priority > 2 {
		return fmt.Errorf("priority must be 0, 1, or 2, got %d", s.Priority)
	}
	if s.Volume < 0 || s.Volume > 15 {
		return fmt.Errorf("volume must be 0-15, got %d", s.Volume)
	}
	if s.Squelch < 0 || s.Squelch > 15 {
		return fmt.Errorf("squelch must be 0-15, got %d", s.Squelch)
	}
	if s.Contrast < 1 || s.Contrast > 15 {
		return fmt.Errorf("contrast must be 1-15, got %d", s.Contrast)
	}
	if len(s.ScanBanks) != 10 {
		return fmt.Errorf("scan_banks must be a 10-character string, got %d characters", len(s.ScanBanks))
	}
	for i, c := range s.ScanBanks {
		if c != '0' && c != '1' {
			return fmt.Errorf("scan_banks[%d] must be '0' or '1', got %q", i, c)
		}
	}
	if s.WeatherScan < 0 || s.WeatherScan > 1 {
		return fmt.Errorf("weather_scan must be 0 or 1, got %d", s.WeatherScan)
	}
	return nil
}

// DefaultSettings returns a Settings struct with sensible defaults.
func DefaultSettings() *Settings {
	return &Settings{
		Backlight:   BacklightAlwaysOn,
		BatterySave: 0,
		KeyBeep:     3,
		KeyLock:     false,
		Priority:    PriorityOff,
		Volume:      10,
		Squelch:     3,
		Contrast:    7,
		ScanBanks:   "1111111111",
		WeatherScan: 0,
	}
}
