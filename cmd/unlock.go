package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var unlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Rescue a scanner stuck in Program Mode",
	Long: `Sends the EPG (Exit Program) command to the scanner.

If the scanner has become stuck in Program Mode (e.g. after a crashed session),
run this command to return it to normal operation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sc, conn, err := openScanner()
		if err != nil {
			return err
		}
		defer conn.Close()

		fmt.Println("Sending exit-program-mode command...")
		if err := sc.ExitProgramMode(); err != nil {
			return fmt.Errorf("unlock failed: %w", err)
		}
		fmt.Println("Scanner unlocked successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(unlockCmd)
}
