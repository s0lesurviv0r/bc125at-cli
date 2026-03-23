package scanner

import (
	"fmt"
	"strconv"
	"strings"
)

// Modulation represents the modulation type for a channel.
type Modulation string

const (
	ModAuto Modulation = "AUTO"
	ModFM   Modulation = "FM"
	ModNFM  Modulation = "NFM"
	ModAM   Modulation = "AM"
)

var validModulations = map[Modulation]bool{
	ModAuto: true,
	ModFM:   true,
	ModNFM:  true,
	ModAM:   true,
}

var validDelays = map[int]bool{
	-10: true,
	-5:  true,
	0:   true,
	1:   true,
	2:   true,
	3:   true,
	4:   true,
	5:   true,
}

// Valid frequency bands for the BC125AT in MHz.
var validFrequencyBands = [][2]float64{
	{25.0, 54.0},
	{108.0, 174.0},
	{225.0, 380.0},
	{400.0, 512.0},
}

// Channel represents a single BC125AT scanner channel.
type Channel struct {
	Index        int        `json:"index"`
	Name         string     `json:"name"`
	FrequencyMHz float64    `json:"frequency_mhz"`
	Modulation   Modulation `json:"modulation"`
	CTCSS        int        `json:"ctcss_dcs_code"`
	Delay        int        `json:"delay"`
	LockedOut    bool       `json:"locked_out"`
	Priority     bool       `json:"priority"`
}

// IsEmpty returns true if the channel has no frequency set.
func (c *Channel) IsEmpty() bool {
	return c.FrequencyMHz == 0
}

// CTCSSName returns the human-readable CTCSS/DCS tone name.
func (c *Channel) CTCSSName() string {
	name, err := ToneToHuman(c.CTCSS)
	if err != nil {
		return fmt.Sprintf("unknown(%d)", c.CTCSS)
	}
	return name
}

// Validate checks that all channel fields are valid.
func (c *Channel) Validate() error {
	if c.Index < 1 || c.Index > 500 {
		return fmt.Errorf("channel %d: index must be 1-500", c.Index)
	}
	if len(c.Name) > 16 {
		return fmt.Errorf("channel %d: name %q exceeds 16 characters", c.Index, c.Name)
	}
	if !c.IsEmpty() {
		if !IsValidFrequency(c.FrequencyMHz) {
			return fmt.Errorf("channel %d: frequency %.4f MHz is out of valid bands (25-54, 108-174, 225-380, 400-512 MHz)", c.Index, c.FrequencyMHz)
		}
	}
	if !validModulations[c.Modulation] {
		return fmt.Errorf("channel %d: invalid modulation %q (must be AUTO, FM, NFM, or AM)", c.Index, c.Modulation)
	}
	if _, ok := ValidTones[c.CTCSS]; !ok {
		return fmt.Errorf("channel %d: invalid CTCSS/DCS code %d", c.Index, c.CTCSS)
	}
	if !validDelays[c.Delay] {
		return fmt.Errorf("channel %d: invalid delay %d (must be -10, -5, 0, or 1-5)", c.Index, c.Delay)
	}
	return nil
}

// IsValidFrequency returns true if the given MHz value is within a valid BC125AT band.
func IsValidFrequency(mhz float64) bool {
	for _, band := range validFrequencyBands {
		if mhz >= band[0] && mhz <= band[1] {
			return true
		}
	}
	return false
}

// FreqToInternal converts a MHz frequency to the BC125AT's internal integer format (× 10000).
func FreqToInternal(mhz float64) int {
	return int(mhz * 10000)
}

// InternalToFreq converts a BC125AT internal frequency integer to MHz.
func InternalToFreq(internal int) float64 {
	return float64(internal) / 10000.0
}

// ParseChannelResponse parses a CIN command response into a Channel.
// Expected format: CIN,<idx>,<name>,<freq>,<mod>,<ctcss>,<delay>,<lock>,<priority>
func ParseChannelResponse(resp string) (*Channel, error) {
	resp = strings.TrimSpace(resp)
	data := strings.TrimPrefix(resp, "CIN,")
	parts := strings.Split(data, ",")
	if len(parts) != 8 {
		return nil, fmt.Errorf("unexpected CIN response format (got %d fields): %s", len(parts), resp)
	}

	idx, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid channel index %q in response", parts[0])
	}

	name := parts[1]

	freqInternal, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid frequency %q in response", parts[2])
	}

	mod := Modulation(strings.ToUpper(parts[3]))

	ctcss, err := strconv.Atoi(parts[4])
	if err != nil {
		return nil, fmt.Errorf("invalid CTCSS code %q in response", parts[4])
	}

	delay, err := strconv.Atoi(parts[5])
	if err != nil {
		return nil, fmt.Errorf("invalid delay %q in response", parts[5])
	}

	lock, err := strconv.Atoi(parts[6])
	if err != nil {
		return nil, fmt.Errorf("invalid lock state %q in response", parts[6])
	}

	priority, err := strconv.Atoi(parts[7])
	if err != nil {
		return nil, fmt.Errorf("invalid priority %q in response", parts[7])
	}

	return &Channel{
		Index:        idx,
		Name:         name,
		FrequencyMHz: InternalToFreq(freqInternal),
		Modulation:   mod,
		CTCSS:        ctcss,
		Delay:        delay,
		LockedOut:    lock == 1,
		Priority:     priority == 1,
	}, nil
}

// FormatChannelCommand returns the CIN command string to write a channel.
func FormatChannelCommand(ch *Channel) string {
	lock := 0
	if ch.LockedOut {
		lock = 1
	}
	priority := 0
	if ch.Priority {
		priority = 1
	}
	freq := 0
	if ch.FrequencyMHz > 0 {
		freq = FreqToInternal(ch.FrequencyMHz)
	}
	return fmt.Sprintf("CIN,%d,%s,%d,%s,%d,%d,%d,%d",
		ch.Index, ch.Name, freq, ch.Modulation, ch.CTCSS, ch.Delay, lock, priority)
}
