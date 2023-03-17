package cmd

import (
	"fmt"
	"os"

	"github.com/chand1012/git2gpt/lib"
	"github.com/spf13/cobra"
)

var repoPath string
var preambleFile string
var outputFile string
var estimateTokens bool

var rootCmd = &cobra.Command{
	Use:   "git2gpt [flags] /path/to/git/repository",
	Short: "git2gpt is a utility to convert a Git repository to a text file for input into GPT-4",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoPath = args[0]
		output, err := lib.ProcessGitRepo(repoPath, preambleFile)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
		if outputFile != "" {
			// if output file exists, throw error
			if _, err := os.Stat(outputFile); err == nil {
				fmt.Printf("Error: output file %s already exists\n", outputFile)
				os.Exit(1)
			}
			err = os.WriteFile(outputFile, []byte(output), 0644)
			if err != nil {
				fmt.Printf("Error: could not write to output file %s\n", outputFile)
				os.Exit(1)
			}
		} else {
			fmt.Println(output)
		}
		if estimateTokens {
			fmt.Printf("Estimated number of tokens: %d\n", lib.EstimateTokens(output))
		}
	},
}

func init() {
	rootCmd.Flags().StringVarP(&preambleFile, "preamble", "p", "", "path to preamble text file")
	// output to file flag. Should be a string
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "path to output file")
	// estimate tokens. Should be a bool
	rootCmd.Flags().BoolVarP(&estimateTokens, "estimate", "e", false, "estimate the number of tokens in the output")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
