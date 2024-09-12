package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.Flags().StringArray("state", []string{}, "Filter goroutines by state")
	rootCmd.Flags().Bool("remove-duplicates", false, "Remove duplicate goroutines")
}

var rootCmd = &cobra.Command{
	Use: "crasha",
	//Short: "Hugo is a very fast static site generator",
	//Long: `A Fast and Flexible Static Site Generator built with
	//			  love by spf13 and friends in Go.
	//			  Complete documentation is available at http://hugo.spf13.com`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		states, err := cmd.Flags().GetStringArray("state")
		if err != nil {
			return fmt.Errorf("failed to get state flag: %w", err)
		}

		removeDuplicates, err := cmd.Flags().GetBool("remove-duplicates")
		if err != nil {
			return fmt.Errorf("failed to get remove-duplicates flag: %w", err)
		}

		file, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}

		sts, err := Parse(file)
		if err != nil {
			return fmt.Errorf("failed to parse file: %w", err)
		}

		fmt.Printf("Found %d goroutines\n", len(sts))

		if len(states) > 0 {
			for i := len(sts) - 1; i >= 0; i-- {
				if !slices.Contains(states, sts[i].GoroutineState) {
					sts = append(sts[:i], sts[i+1:]...)
				}
			}
		}

		if removeDuplicates {
			for i := len(sts) - 1; i >= 0; i-- {
				for j := i - 1; j >= 0; j-- {
					if sts[i].Equal(sts[j]) {
						sts = append(sts[:i], sts[i+1:]...)
						break
					}
				}
			}
		}

		fmt.Printf("Filtered to %d goroutines\n", len(sts))

		for _, st := range sts {
			fmt.Println(st)
		}
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
