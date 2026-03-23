package cmd

import (
	"fmt"
	"os"

	"github.com/jzelek/bc125-cli/formats"
	"github.com/spf13/cobra"
)

var importCSV bool

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Read data from the scanner and write to a file",
	Long: `Reads channel data (and settings for JSON) from the scanner and saves
to a file.

With --csv / -c: writes only channel data in CSV format.
Without --csv:   writes a full JSON backup including settings and all channels.

Use "-" as the file name to write to stdout.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

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

		// Read channels
		fmt.Println("Reading channels from scanner...")
		channels, err := sc.GetAllChannels(func(current, total int) {
			printProgress("Reading", current, total)
		})
		if err != nil {
			return err
		}

		// Open output
		var out *os.File
		if filePath == "-" {
			out = os.Stdout
		} else {
			out, err = os.Create(filePath)
			if err != nil {
				return fmt.Errorf("cannot create file %s: %w", filePath, err)
			}
			defer out.Close()
		}

		if importCSV {
			if err := formats.WriteCSV(out, channels); err != nil {
				return fmt.Errorf("write CSV: %w", err)
			}
		} else {
			// JSON: also read settings
			fmt.Println("Reading settings from scanner...")
			settings, err := sc.GetSettings()
			if err != nil {
				return fmt.Errorf("read settings: %w", err)
			}

			model, _ := sc.GetModel()
			version, _ := sc.GetVersion()

			data := &formats.ScannerData{
				Model:    model,
				Firmware: version,
				Settings: settings,
				Channels: channels,
			}
			if err := formats.WriteJSON(out, data); err != nil {
				return fmt.Errorf("write JSON: %w", err)
			}
		}

		if filePath != "-" {
			fmt.Printf("Saved to %s\n", filePath)
		}
		return nil
	},
}

func init() {
	importCmd.Flags().BoolVarP(&importCSV, "csv", "c", false, "Write CSV format (channels only)")
	rootCmd.AddCommand(importCmd)
}
