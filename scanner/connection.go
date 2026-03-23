package scanner

import (
	"bufio"
	"fmt"
	"runtime"
	"strings"
	"time"

	"go.bug.st/serial"
)

// Connection manages the serial connection to a BC125AT scanner.
type Connection struct {
	port    serial.Port
	reader  *bufio.Reader
	verbose bool
}

// FindPorts returns candidate serial port names for the BC125AT.
// On Linux it prefers /dev/ttyACM*, on macOS /dev/tty.usbmodem*, on Windows COM*.
func FindPorts() ([]string, error) {
	all, err := serial.GetPortsList()
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate serial ports: %w", err)
	}

	var candidates []string
	for _, p := range all {
		if isCandidatePort(p) {
			candidates = append(candidates, p)
		}
	}
	return candidates, nil
}

func isCandidatePort(name string) bool {
	switch runtime.GOOS {
	case "linux":
		return strings.HasPrefix(name, "/dev/ttyACM") ||
			strings.HasPrefix(name, "/dev/ttyUSB")
	case "darwin":
		return strings.HasPrefix(name, "/dev/tty.usbmodem") ||
			strings.HasPrefix(name, "/dev/cu.usbmodem")
	case "windows":
		return strings.HasPrefix(name, "COM")
	}
	return true
}

// Open opens a serial connection to the BC125AT on the given port.
func Open(portName string, verbose bool) (*Connection, error) {
	mode := &serial.Mode{
		BaudRate: 9600,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", portName, err)
	}

	// 3-second read timeout per command
	if err := port.SetReadTimeout(3 * time.Second); err != nil {
		port.Close()
		return nil, fmt.Errorf("failed to set read timeout: %w", err)
	}

	conn := &Connection{
		port:    port,
		reader:  bufio.NewReader(port),
		verbose: verbose,
	}
	return conn, nil
}

// Close closes the serial connection.
func (c *Connection) Close() error {
	return c.port.Close()
}

// Exec sends a raw AT command string and returns the response.
// The \r line terminator is appended automatically.
// Returns an error if the scanner responds with an error suffix.
func (c *Connection) Exec(cmd string) (string, error) {
	if c.verbose {
		fmt.Printf("  >> %s\n", cmd)
	}

	// Send command
	if _, err := c.port.Write([]byte(cmd + "\r")); err != nil {
		return "", fmt.Errorf("write error: %w", err)
	}

	// Read until carriage return
	resp, err := c.reader.ReadString('\r')
	if err != nil {
		return "", fmt.Errorf("read error: %w", err)
	}
	resp = strings.TrimRight(resp, "\r\n")

	if c.verbose {
		fmt.Printf("  << %s\n", resp)
	}

	// Check for scanner error responses
	if strings.HasSuffix(resp, "ERR") || strings.HasSuffix(resp, ",NG") {
		return "", fmt.Errorf("scanner returned error: %s", resp)
	}

	return resp, nil
}

// ExecRaw sends a command and returns the raw response without error-checking.
// Useful for the interactive shell where the user should see all responses.
func (c *Connection) ExecRaw(cmd string) (string, error) {
	if _, err := c.port.Write([]byte(cmd + "\r")); err != nil {
		return "", fmt.Errorf("write error: %w", err)
	}
	resp, err := c.reader.ReadString('\r')
	if err != nil {
		return "", fmt.Errorf("read error (is the scanner connected and powered on?): %w", err)
	}
	return strings.TrimRight(resp, "\r\n"), nil
}
