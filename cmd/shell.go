package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jzelek/bc125-cli/scanner"
	"github.com/spf13/cobra"
)

var shellClearHistory bool

var shellCmd = &cobra.Command{
	Use:   "shell [file]",
	Short: "Open an interactive AT command shell",
	Long: `Opens an interactive shell for sending raw AT commands to the scanner.

If a script file is given its commands are executed in order then the shell exits.
Use "-" as the file name to read commands from stdin (non-interactive).

The scanner does NOT need to be in Program Mode for all commands, but most
channel/settings commands require you to send PRG first.

Built-in shell commands:
  help           Show this help
  quit / exit    Close the shell
  history        Print command history

Example AT commands:
  MDL            Get model name
  VER            Get firmware version
  PRG            Enter program mode
  EPG            Exit program mode
  CIN,1          Read channel 1
  DCH,1          Delete channel 1
  VOL,10         Set volume to 10`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine input source
		var input *os.File
		interactive := false

		if len(args) == 0 {
			input = os.Stdin
			interactive = true
		} else if args[0] == "-" {
			input = os.Stdin
		} else {
			f, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("cannot open script file: %w", err)
			}
			defer f.Close()
			input = f
		}

		// Connect to scanner
		port, err := resolvePort()
		if err != nil {
			return err
		}
		conn, err := scanner.Open(port, false) // shell uses its own verbose output
		if err != nil {
			return fmt.Errorf("cannot open scanner: %w", err)
		}
		defer conn.Close()

		if interactive {
			fmt.Println("BC125AT Shell — type 'help' for commands, 'quit' to exit")
		}

		return runShell(conn, input, interactive, shellClearHistory)
	},
}

// historyFile returns the path to the shell history file.
func historyFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".bc125at_history")
}

func runShell(conn *scanner.Connection, input *os.File, interactive bool, clearHistory bool) error {
	// Load history
	var history []string
	hfile := historyFile()

	if !clearHistory && hfile != "" {
		if data, err := os.ReadFile(hfile); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if line = strings.TrimSpace(line); line != "" {
					history = append(history, line)
				}
			}
		}
	}

	scanner := bufio.NewScanner(input)

	for {
		if interactive {
			fmt.Print("bc125at> ")
		}

		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		upper := strings.ToUpper(line)

		// Built-in shell commands
		switch upper {
		case "QUIT", "EXIT":
			if interactive {
				fmt.Println("Goodbye.")
			}
			saveHistory(hfile, history)
			return nil

		case "HELP":
			printShellHelp()
			continue

		case "HISTORY":
			for i, h := range history {
				fmt.Printf("%4d  %s\n", i+1, h)
			}
			continue
		}

		// Echo non-interactive script lines
		if !interactive {
			fmt.Printf(">> %s\n", line)
		}

		// Send to scanner
		resp, err := conn.ExecRaw(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			if !interactive {
				return err
			}
		} else {
			fmt.Println(resp)
		}

		history = append(history, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("input error: %w", err)
	}

	saveHistory(hfile, history)
	return nil
}

func saveHistory(path string, history []string) {
	if path == "" || len(history) == 0 {
		return
	}
	// Keep at most 500 entries
	if len(history) > 500 {
		history = history[len(history)-500:]
	}
	data := strings.Join(history, "\n") + "\n"
	_ = os.WriteFile(path, []byte(data), 0600)
}

func printShellHelp() {
	fmt.Print(`Shell built-in commands:
  help       Show this help
  history    Print command history
  quit/exit  Close the shell

Common AT commands (send PRG first for most):
  MDL        Get model
  VER        Get firmware version
  PRG        Enter program mode
  EPG        Exit program mode
  CIN,<n>    Read channel n (1-500)
  CIN,<n>,<name>,<freq>,<mod>,<ctcss>,<delay>,<lock>,<pri>  Write channel
  DCH,<n>    Delete channel n
  VOL[,<n>]  Get/set volume (0-15)
  SQL[,<n>]  Get/set squelch (0-15)
  BLT[,<m>]  Get/set backlight (AO/10/30/SQ)

Lines starting with # are treated as comments.
`)
}

func init() {
	shellCmd.Flags().BoolVarP(&shellClearHistory, "clear-history", "c", false, "Clear shell history before starting")
	rootCmd.AddCommand(shellCmd)
}
