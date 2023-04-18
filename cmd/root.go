package cmd

import (
	"fmt"
	"os"

	"github.com/chand1012/git2gpt/prompt"
	"github.com/spf13/cobra"
)

var repoPath string
var preambleFile string
var outputFile string
var estimateTokens bool
var ignoreFilePath string
var ignoreGitignore bool
var outputJSON bool
var debug bool

var rootCmd = &cobra.Command{
	Use:   "git2gpt [flags] /path/to/git/repository",
	Short: "git2gpt is a utility to convert a Git repository to a text file for input into GPT-4",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoPath = args[0]
		ignoreList := prompt.GenerateIgnoreList(repoPath, ignoreFilePath, !ignoreGitignore)
		repo, err := prompt.ProcessGitRepo(repoPath, ignoreList)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
		if outputJSON {
			output, err := prompt.MarshalRepo(repo)
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
				if !debug {
					fmt.Println(string(output))
				}
			}
			return
		}
		output, err := prompt.OutputGitRepo(repo, preambleFile)
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
			if !debug {
				fmt.Println(output)
			}
		}
		if estimateTokens {
			fmt.Printf("Estimated number of tokens: %d\n", prompt.EstimateTokens(output))
		}
	},
}

func init() {
	rootCmd.Flags().StringVarP(&preambleFile, "preamble", "p", "", "path to preamble text file")
	// output to file flag. Should be a string
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "path to output file")
	// estimate tokens. Should be a bool
	rootCmd.Flags().BoolVarP(&estimateTokens, "estimate", "e", false, "estimate the number of tokens in the output")
	// ignore file path. Should be a string
	rootCmd.Flags().StringVarP(&ignoreFilePath, "ignore", "i", "", "path to .gptignore file")
	// ignore gitignore. Should be a bool
	rootCmd.Flags().BoolVarP(&ignoreGitignore, "ignore-gitignore", "g", false, "ignore .gitignore file")
	// output JSON. Should be a bool
	rootCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "output JSON")
	// debug. Should be a bool
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "debug mode. Do not output to standard output")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
