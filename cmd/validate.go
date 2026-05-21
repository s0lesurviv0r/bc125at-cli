package cmd

import (
	"fmt"
	"os"

	"github.com/jacobzelek/bc125-cli/formats"
	"github.com/spf13/cobra"
)

var validateCSV bool

var validateCmd = &cobra.Command{
	Use:   "validate <file>",
	Short: "Validate a scanner backup file without writing to the device",
	Long: `Reads and validates a JSON or CSV backup file without connecting to
the scanner. Reports any errors found in channel data or settings.

With --csv / -c: validate a CSV channel file.
Without --csv:   validate a JSON backup file.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		in, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("cannot open %s: %w", filePath, err)
		}
		defer in.Close()

		var errs []error

		if validateCSV {
			channels, err := formats.ReadCSV(in)
			if err != nil {
				return fmt.Errorf("parse error: %w", err)
			}
			errs = formats.ValidateCSVChannels(channels)
			if len(errs) == 0 {
				fmt.Printf("OK — %d channels, no errors\n", len(channels))
			}
		} else {
			data, err := formats.ReadJSON(in)
			if err != nil {
				return fmt.Errorf("parse error: %w", err)
			}
			errs = formats.ValidateJSONData(data)
			if len(errs) == 0 {
				fmt.Printf("OK — %d channels, no errors\n", len(data.Channels))
			}
		}

		if len(errs) > 0 {
			fmt.Fprintf(os.Stderr, "Validation errors (%d):\n", len(errs))
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "  %v\n", e)
			}
			return fmt.Errorf("%d validation error(s) found", len(errs))
		}
		return nil
	},
}

func init() {
	validateCmd.Flags().BoolVarP(&validateCSV, "csv", "c", false, "Validate a CSV file (channels only)")
	rootCmd.AddCommand(validateCmd)
}
