package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"
)

func init() {
	filterCmd.Flags().StringArray("state", []string{}, "Filter goroutines by state")
	filterCmd.Flags().Bool("remove-duplicates", false, "Remove duplicate goroutines")
	rootCmd.AddCommand(filterCmd)

	rootCmd.AddCommand(statesCmd)
}

var rootCmd = &cobra.Command{
	Use: "crasha",
}

var filterCmd = &cobra.Command{
	Use:   "filter",
	Short: "TODO",
	Args:  cobra.ExactArgs(1),
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

var statesCmd = &cobra.Command{
	Use:   "states",
	Short: "Extract all goroutine states out of a file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}

		sts, err := Parse(file)
		if err != nil {
			return fmt.Errorf("failed to parse file: %w", err)
		}

		fmt.Printf("Found %d goroutines\n", len(sts))

		var res []string
		for _, st := range sts {
			if !slices.Contains(res, st.GoroutineState) {
				res = append(res, st.GoroutineState)
			}
		}

		for _, state := range res {
			fmt.Println(state)
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
