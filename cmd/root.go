package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jzelek/bc125-cli/scanner"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

// Global flags shared by all scanner commands.
var (
	flagPort    string
	flagVerbose bool
)

var rootCmd = &cobra.Command{
	Use:   "bc125at",
	Short: "A CLI for programming the Uniden BC125AT scanner",
	Long: `bc125at - Uniden BC125AT Scanner CLI

Read and write channels and settings on a BC125AT radio scanner
over USB serial. Supports JSON and CSV file formats.

The scanner must be powered on and connected via USB before use.`,
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagPort, "port", "p", "", "Serial port to use (e.g. /dev/ttyACM0, COM3). Auto-detected if omitted.")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Enable verbose output (shows raw AT commands)")
}

// resolvePort returns the port to use. If flagPort is set it is used directly;
// otherwise candidate ports are enumerated and the user is prompted if there
// is more than one.
func resolvePort() (string, error) {
	if flagPort != "" {
		return flagPort, nil
	}

	ports, err := scanner.FindPorts()
	if err != nil {
		return "", fmt.Errorf("port detection failed: %w", err)
	}

	if len(ports) == 0 {
		return "", fmt.Errorf("no scanner port found — is the BC125AT connected and powered on?\n" +
			"Use --port to specify the port manually.")
	}

	if len(ports) == 1 {
		fmt.Printf("Using port: %s\n", ports[0])
		return ports[0], nil
	}

	// Multiple candidates — ask the user
	fmt.Println("Multiple serial ports found. Select one:")
	for i, p := range ports {
		fmt.Printf("  [%d] %s\n", i+1, p)
	}
	fmt.Print("Enter number: ")
	var choice int
	if _, err := fmt.Scan(&choice); err != nil || choice < 1 || choice > len(ports) {
		return "", fmt.Errorf("invalid selection")
	}
	return ports[choice-1], nil
}

// openScanner opens a connection to the scanner, resolving the port automatically.
func openScanner() (*scanner.Scanner, *scanner.Connection, error) {
	port, err := resolvePort()
	if err != nil {
		return nil, nil, err
	}

	conn, err := scanner.Open(port, flagVerbose)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot open scanner: %w", err)
	}

	sc := scanner.New(conn)
	return sc, conn, nil
}

// printProgress displays a progress line on a single terminal line.
func printProgress(prefix string, current, total int) {
	pct := current * 100 / total
	bar := strings.Repeat("=", pct/2) + strings.Repeat(" ", 50-pct/2)
	fmt.Printf("\r%s [%s] %d/%d", prefix, bar, current, total)
	if current == total {
		fmt.Println()
	}
}
