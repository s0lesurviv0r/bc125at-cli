package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var wipeCmd = &cobra.Command{
	Use:   "wipe",
	Short: "Delete all channels from the scanner",
	Long: `Deletes all 500 channels from the scanner.

You will be asked to confirm before any data is erased. This operation
cannot be undone — use 'bc125at import' first to make a backup.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("WARNING: This will permanently delete all 500 channels from the scanner.")
		fmt.Println("Make a backup first with:  bc125at import backup.json")
		fmt.Println()
		fmt.Print("Type \"I understand the consequences.\" to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("could not read confirmation: %w", err)
		}
		line = strings.TrimSpace(line)

		if line != "I understand the consequences." {
			return fmt.Errorf("confirmation text did not match — aborting")
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

		fmt.Println("Wiping channels...")
		if err := sc.WipeAllChannels(func(current, total int) {
			printProgress("Deleting", current, total)
		}); err != nil {
			return err
		}

		fmt.Println("All channels deleted.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(wipeCmd)
}
