package cmd

import (
	"fmt"

	"github.com/jzelek/bc125-cli/scanner"
	"github.com/spf13/cobra"
)

var tonesCmd = &cobra.Command{
	Use:   "tones",
	Short: "List all valid CTCSS and DCS tone names",
	Long: `Prints all valid tone identifiers that can be used in the ctcss_dcs
column of a CSV file or the ctcss_dcs_code field of a JSON file.

Special values:
  none       No tone squelch (open squelch)
  search     CTCSS/DCS search mode
  no_tone    No tone

CTCSS tones are specified as: ctcss_<freq>  (e.g. ctcss_100.0)
DCS codes are specified as:   dcs_<code>    (e.g. dcs_023)`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Valid CTCSS/DCS tones:")
		fmt.Println()
		fmt.Println("Code  Name")
		fmt.Println("----  ----------------")
		for _, code := range scanner.SortedToneKeys() {
			name := scanner.ValidTones[code]
			fmt.Printf("%-4d  %s\n", code, name)
		}
	},
}

func init() {
	rootCmd.AddCommand(tonesCmd)
}
