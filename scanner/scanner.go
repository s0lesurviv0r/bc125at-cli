package scanner

import (
	"fmt"
	"strconv"
	"strings"
)

// Scanner provides high-level operations for the BC125AT.
type Scanner struct {
	Conn *Connection
}

// New creates a Scanner wrapping the given Connection.
func New(conn *Connection) *Scanner {
	return &Scanner{Conn: conn}
}

// stripPrefix removes a leading prefix from a response string.
func stripPrefix(resp, prefix string) string {
	return strings.TrimPrefix(resp, prefix)
}

// expectOK returns an error if the response does not end with ",OK".
func expectOK(cmd, resp string) error {
	if !strings.HasSuffix(resp, ",OK") && resp != cmd+",OK" {
		return fmt.Errorf("unexpected response to %q: %s", cmd, resp)
	}
	return nil
}

// GetModel returns the scanner model string (e.g. "BC125AT").
func (s *Scanner) GetModel() (string, error) {
	resp, err := s.Conn.Exec("MDL")
	if err != nil {
		return "", err
	}
	return stripPrefix(resp, "MDL,"), nil
}

// GetVersion returns the firmware version string.
func (s *Scanner) GetVersion() (string, error) {
	resp, err := s.Conn.Exec("VER")
	if err != nil {
		return "", err
	}
	return stripPrefix(resp, "VER,"), nil
}

// EnterProgramMode puts the scanner into programming mode.
func (s *Scanner) EnterProgramMode() error {
	resp, err := s.Conn.Exec("PRG")
	if err != nil {
		return fmt.Errorf("enter program mode: %w", err)
	}
	return expectOK("PRG", resp)
}

// ExitProgramMode exits programming mode. Also used to unlock a stuck scanner.
func (s *Scanner) ExitProgramMode() error {
	resp, err := s.Conn.Exec("EPG")
	if err != nil {
		return fmt.Errorf("exit program mode: %w", err)
	}
	return expectOK("EPG", resp)
}

// GetChannel reads channel data from the scanner. Must be in PRG mode.
func (s *Scanner) GetChannel(index int) (*Channel, error) {
	resp, err := s.Conn.Exec(fmt.Sprintf("CIN,%d", index))
	if err != nil {
		return nil, fmt.Errorf("read channel %d: %w", index, err)
	}
	return ParseChannelResponse(resp)
}

// SetChannel writes channel data to the scanner. Must be in PRG mode.
func (s *Scanner) SetChannel(ch *Channel) error {
	cmd := FormatChannelCommand(ch)
	resp, err := s.Conn.Exec(cmd)
	if err != nil {
		return fmt.Errorf("write channel %d: %w", ch.Index, err)
	}
	return expectOK("CIN", resp)
}

// DeleteChannel deletes a channel (sets it to empty). Must be in PRG mode.
func (s *Scanner) DeleteChannel(index int) error {
	resp, err := s.Conn.Exec(fmt.Sprintf("DCH,%d", index))
	if err != nil {
		return fmt.Errorf("delete channel %d: %w", index, err)
	}
	return expectOK("DCH", resp)
}

// GetAllChannels reads all 500 channels. Must be in PRG mode.
// progress(current, total) is called after each read if not nil.
func (s *Scanner) GetAllChannels(progress func(current, total int)) ([]*Channel, error) {
	const total = 500
	channels := make([]*Channel, total)
	for i := 1; i <= total; i++ {
		ch, err := s.GetChannel(i)
		if err != nil {
			return nil, fmt.Errorf("error at channel %d: %w", i, err)
		}
		channels[i-1] = ch
		if progress != nil {
			progress(i, total)
		}
	}
	return channels, nil
}

// SetAllChannels writes channels to the scanner. Must be in PRG mode.
// progress(current, total) is called after each write if not nil.
func (s *Scanner) SetAllChannels(channels []*Channel, progress func(current, total int)) error {
	total := len(channels)
	for i, ch := range channels {
		if err := s.SetChannel(ch); err != nil {
			return err
		}
		if progress != nil {
			progress(i+1, total)
		}
	}
	return nil
}

// WipeAllChannels deletes all 500 channels. Must be in PRG mode.
func (s *Scanner) WipeAllChannels(progress func(current, total int)) error {
	const total = 500
	for i := 1; i <= total; i++ {
		if err := s.DeleteChannel(i); err != nil {
			return err
		}
		if progress != nil {
			progress(i, total)
		}
	}
	return nil
}

// --- Settings getters/setters ---

// GetSettings reads all scanner settings. Must be in PRG mode.
func (s *Scanner) GetSettings() (*Settings, error) {
	settings := &Settings{}

	// BLT - backlight
	resp, err := s.Conn.Exec("BLT")
	if err != nil {
		return nil, fmt.Errorf("read backlight: %w", err)
	}
	settings.Backlight = BacklightMode(stripPrefix(resp, "BLT,"))

	// BSV - battery save
	resp, err = s.Conn.Exec("BSV")
	if err != nil {
		return nil, fmt.Errorf("read battery save: %w", err)
	}
	if settings.BatterySave, err = strconv.Atoi(stripPrefix(resp, "BSV,")); err != nil {
		return nil, fmt.Errorf("parse battery save: %w", err)
	}

	// KBP - key beep + key lock  (format: KBP,<beep>,<lock>)
	resp, err = s.Conn.Exec("KBP")
	if err != nil {
		return nil, fmt.Errorf("read key beep: %w", err)
	}
	kbpData := stripPrefix(resp, "KBP,")
	kbpParts := strings.Split(kbpData, ",")
	if len(kbpParts) == 2 {
		if settings.KeyBeep, err = strconv.Atoi(kbpParts[0]); err != nil {
			return nil, fmt.Errorf("parse key beep: %w", err)
		}
		lock, err := strconv.Atoi(kbpParts[1])
		if err != nil {
			return nil, fmt.Errorf("parse key lock: %w", err)
		}
		settings.KeyLock = lock == 1
	}

	// PRI - priority mode
	resp, err = s.Conn.Exec("PRI")
	if err != nil {
		return nil, fmt.Errorf("read priority: %w", err)
	}
	priVal, err := strconv.Atoi(stripPrefix(resp, "PRI,"))
	if err != nil {
		return nil, fmt.Errorf("parse priority: %w", err)
	}
	settings.Priority = PriorityMode(priVal)

	// VOL - volume
	resp, err = s.Conn.Exec("VOL")
	if err != nil {
		return nil, fmt.Errorf("read volume: %w", err)
	}
	if settings.Volume, err = strconv.Atoi(stripPrefix(resp, "VOL,")); err != nil {
		return nil, fmt.Errorf("parse volume: %w", err)
	}

	// SQL - squelch
	resp, err = s.Conn.Exec("SQL")
	if err != nil {
		return nil, fmt.Errorf("read squelch: %w", err)
	}
	if settings.Squelch, err = strconv.Atoi(stripPrefix(resp, "SQL,")); err != nil {
		return nil, fmt.Errorf("parse squelch: %w", err)
	}

	// CNT - contrast
	resp, err = s.Conn.Exec("CNT")
	if err != nil {
		return nil, fmt.Errorf("read contrast: %w", err)
	}
	if settings.Contrast, err = strconv.Atoi(stripPrefix(resp, "CNT,")); err != nil {
		return nil, fmt.Errorf("parse contrast: %w", err)
	}

	// SCG - scan banks
	resp, err = s.Conn.Exec("SCG")
	if err != nil {
		return nil, fmt.Errorf("read scan banks: %w", err)
	}
	settings.ScanBanks = stripPrefix(resp, "SCG,")

	// WXS - weather scan
	resp, err = s.Conn.Exec("WXS")
	if err != nil {
		return nil, fmt.Errorf("read weather scan: %w", err)
	}
	if settings.WeatherScan, err = strconv.Atoi(stripPrefix(resp, "WXS,")); err != nil {
		return nil, fmt.Errorf("parse weather scan: %w", err)
	}

	return settings, nil
}

// SetSettings writes all scanner settings. Must be in PRG mode.
func (s *Scanner) SetSettings(settings *Settings) error {
	// BLT
	resp, err := s.Conn.Exec(fmt.Sprintf("BLT,%s", settings.Backlight))
	if err != nil {
		return fmt.Errorf("set backlight: %w", err)
	}
	if err := expectOK("BLT", resp); err != nil {
		return err
	}

	// BSV
	resp, err = s.Conn.Exec(fmt.Sprintf("BSV,%d", settings.BatterySave))
	if err != nil {
		return fmt.Errorf("set battery save: %w", err)
	}
	if err := expectOK("BSV", resp); err != nil {
		return err
	}

	// KBP
	lockVal := 0
	if settings.KeyLock {
		lockVal = 1
	}
	resp, err = s.Conn.Exec(fmt.Sprintf("KBP,%d,%d", settings.KeyBeep, lockVal))
	if err != nil {
		return fmt.Errorf("set key beep: %w", err)
	}
	if err := expectOK("KBP", resp); err != nil {
		return err
	}

	// PRI
	resp, err = s.Conn.Exec(fmt.Sprintf("PRI,%d", settings.Priority))
	if err != nil {
		return fmt.Errorf("set priority: %w", err)
	}
	if err := expectOK("PRI", resp); err != nil {
		return err
	}

	// VOL
	resp, err = s.Conn.Exec(fmt.Sprintf("VOL,%d", settings.Volume))
	if err != nil {
		return fmt.Errorf("set volume: %w", err)
	}
	if err := expectOK("VOL", resp); err != nil {
		return err
	}

	// SQL
	resp, err = s.Conn.Exec(fmt.Sprintf("SQL,%d", settings.Squelch))
	if err != nil {
		return fmt.Errorf("set squelch: %w", err)
	}
	if err := expectOK("SQL", resp); err != nil {
		return err
	}

	// CNT
	resp, err = s.Conn.Exec(fmt.Sprintf("CNT,%d", settings.Contrast))
	if err != nil {
		return fmt.Errorf("set contrast: %w", err)
	}
	if err := expectOK("CNT", resp); err != nil {
		return err
	}

	// SCG
	resp, err = s.Conn.Exec(fmt.Sprintf("SCG,%s", settings.ScanBanks))
	if err != nil {
		return fmt.Errorf("set scan banks: %w", err)
	}
	if err := expectOK("SCG", resp); err != nil {
		return err
	}

	// WXS
	resp, err = s.Conn.Exec(fmt.Sprintf("WXS,%d", settings.WeatherScan))
	if err != nil {
		return fmt.Errorf("set weather scan: %w", err)
	}
	return expectOK("WXS", resp)
}
