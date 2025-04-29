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
var includeFilePath string // New: Add variable for include file path
var ignoreGitignore bool
var outputJSON bool
var outputXML bool
var debug bool
var scrubComments bool
var rootCmd = &cobra.Command{
        Use:   "git2gpt [flags] /path/to/git/repository [/path/to/another/repository ...]",
        Short: "git2gpt is a utility to convert one or more Git repositories to a text file for input into an LLM",
        Args:  cobra.MinimumNArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
                combinedRepo := &prompt.GitRepo{
                        Files: []prompt.GitFile{},
                }
                for _, path := range args {
                        repoPath = path
                        ignoreList := prompt.GenerateIgnoreList(repoPath, ignoreFilePath, !ignoreGitignore)
                        includeList := prompt.GenerateIncludeList(repoPath, includeFilePath) // New: Generate include list
                        repo, err := prompt.ProcessGitRepo(repoPath, includeList, ignoreList) // Modified: Pass includeList
                        if err != nil {
                                fmt.Printf("Error processing %s: %s\n", repoPath, err)
                                os.Exit(1)
                        }
                        combinedRepo.Files = append(combinedRepo.Files, repo.Files...)
                }
                combinedRepo.FileCount = len(combinedRepo.Files)
                if outputJSON {
                        output, err := prompt.MarshalRepo(combinedRepo, scrubComments)
                        if err != nil {
                                fmt.Printf("Error: %s\n", err)
                                os.Exit(1)
                        }
                        if outputFile != "" {
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
                if outputXML {
                        output, err := prompt.OutputGitRepoXML(combinedRepo, scrubComments)
                        if err != nil {
                                fmt.Printf("Error: %s\n", err)
                                os.Exit(1)
                        }
                        if err := prompt.ValidateXML(output); err != nil {
                                fmt.Printf("Error: %s\n", err)
                                os.Exit(1)
                        }
                        if outputFile != "" {
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
                        return
                }
                output, err := prompt.OutputGitRepo(combinedRepo, preambleFile, scrubComments)
                if err != nil {
                        fmt.Printf("Error: %s\n", err)
                        os.Exit(1)
                }
                if outputFile != "" {
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
        rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "path to output file")
        rootCmd.Flags().BoolVarP(&estimateTokens, "estimate", "e", false, "estimate the number of tokens in the output")
        rootCmd.Flags().StringVarP(&ignoreFilePath, "ignore", "i", "", "path to .gptignore file")
        rootCmd.Flags().StringVarP(&includeFilePath, "include", "I", "", "path to .gptinclude file") // New: Add flag for include file
        rootCmd.Flags().BoolVarP(&ignoreGitignore, "ignore-gitignore", "g", false, "ignore .gitignore file")
        rootCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "output JSON")
        rootCmd.Flags().BoolVarP(&outputXML, "xml", "x", false, "output XML")
        rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "debug mode. Do not output to standard output")
        rootCmd.Flags().BoolVarP(&scrubComments, "scrub-comments", "s", false, "scrub comments from the output. Decreases token count")
        rootCmd.Example = "  git2gpt /path/to/repo1 /path/to/repo2\n  git2gpt -o output.txt /path/to/repo1 /path/to/repo2"
}
func Execute() {
        if err := rootCmd.Execute(); err != nil {
                fmt.Println(err)
                os.Exit(1)
        }
}