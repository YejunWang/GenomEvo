package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bactcg",
	Short: "BactCG is a fast genome and core gene analysis pipeline toolkit",
	Long:  `A Go port of CG_muti_auto logic for genome file cleaning, parsing, and filtering.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
