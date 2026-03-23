package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test the connection to the scanner",
	Long: `Connects to the scanner and retrieves the model and firmware version.
Useful for verifying that the scanner is detected and communicating correctly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sc, conn, err := openScanner()
		if err != nil {
			return err
		}
		defer conn.Close()

		model, err := sc.GetModel()
		if err != nil {
			return fmt.Errorf("failed to read model: %w\n\nMake sure the scanner is powered on and not in Program Mode.", err)
		}

		version, err := sc.GetVersion()
		if err != nil {
			return fmt.Errorf("failed to read firmware version: %w", err)
		}

		fmt.Printf("Connection OK\n")
		fmt.Printf("  Model:    %s\n", model)
		fmt.Printf("  Firmware: %s\n", version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
