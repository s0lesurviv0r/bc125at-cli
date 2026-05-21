package cmd

import (
	"fmt"
	"os"

	"github.com/jacobzelek/bc125-cli/formats"
	"github.com/spf13/cobra"
)

var exportCSV bool

var exportCmd = &cobra.Command{
	Use:   "export <file>",
	Short: "Read data from a file and write it to the scanner",
	Long: `Reads channel data (and settings for JSON) from a file and programs
the scanner.

With --csv / -c: reads channel-only CSV format.
Without --csv:   reads a full JSON file (settings + channels).

Use "-" as the file name to read from stdin.

The file is validated before any data is written to the scanner.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// Open input
		var in *os.File
		var err error
		if filePath == "-" {
			in = os.Stdin
		} else {
			in, err = os.Open(filePath)
			if err != nil {
				return fmt.Errorf("cannot open %s: %w", filePath, err)
			}
			defer in.Close()
		}

		// Parse and validate before touching the scanner
		if exportCSV {
			return exportFromCSV(in)
		}
		return exportFromJSON(in)
	},
}

func exportFromCSV(in *os.File) error {
	fmt.Println("Reading CSV...")
	channels, err := formats.ReadCSV(in)
	if err != nil {
		return err
	}

	errs := formats.ValidateCSVChannels(channels)
	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "Validation errors (%d):\n", len(errs))
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  %v\n", e)
		}
		return fmt.Errorf("aborting: fix validation errors before exporting")
	}

	sc, conn, err := openScanner()
	if err != nil {
		return err
	}
	defer conn.Close()

	fmt.Println("Entering program mode...")
	if err := sc.EnterProgramMode(); err != nil {
		return err
	}
	defer func() {
		if err := sc.ExitProgramMode(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to exit program mode: %v\n", err)
		}
	}()

	fmt.Printf("Writing %d channels to scanner...\n", len(channels))
	return sc.SetAllChannels(channels, func(current, total int) {
		printProgress("Writing", current, total)
	})
}

func exportFromJSON(in *os.File) error {
	fmt.Println("Reading JSON...")
	data, err := formats.ReadJSON(in)
	if err != nil {
		return err
	}

	errs := formats.ValidateJSONData(data)
	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "Validation errors (%d):\n", len(errs))
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  %v\n", e)
		}
		return fmt.Errorf("aborting: fix validation errors before exporting")
	}

	sc, conn, err := openScanner()
	if err != nil {
		return err
	}
	defer conn.Close()

	fmt.Println("Entering program mode...")
	if err := sc.EnterProgramMode(); err != nil {
		return err
	}
	defer func() {
		if err := sc.ExitProgramMode(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to exit program mode: %v\n", err)
		}
	}()

	if data.Settings != nil {
		fmt.Println("Writing settings...")
		if err := sc.SetSettings(data.Settings); err != nil {
			return fmt.Errorf("write settings: %w", err)
		}
	}

	if len(data.Channels) > 0 {
		fmt.Printf("Writing %d channels to scanner...\n", len(data.Channels))
		if err := sc.SetAllChannels(data.Channels, func(current, total int) {
			printProgress("Writing", current, total)
		}); err != nil {
			return err
		}
	}

	fmt.Println("Export complete.")
	return nil
}

func init() {
	exportCmd.Flags().BoolVarP(&exportCSV, "csv", "c", false, "Read CSV format (channels only)")
	rootCmd.AddCommand(exportCmd)
}
