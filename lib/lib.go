package lib

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

func getIgnoreList(ignoreFilePath string) ([]string, error) {
	var ignoreList []string
	file, err := os.Open(ignoreFilePath)
	if err != nil {
		return ignoreList, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ignoreList = append(ignoreList, scanner.Text())
	}
	return ignoreList, scanner.Err()
}

func shouldIgnore(filePath string, ignoreList []string) bool {
	for _, pattern := range ignoreList {
		g := glob.MustCompile(pattern)
		if g.Match(filePath) {
			return true
		}
	}
	return false
}

func ProcessGitRepo(repoPath, preambleFile string) (string, error) {
	ignoreFilePath := filepath.Join(repoPath, ".gptignore")
	gitignoreFilePath := filepath.Join(repoPath, ".gitignore")

	var ignoreList []string
	if _, err := os.Stat(ignoreFilePath); err == nil {
		// .gptignore file exists
		ignoreList, _ = getIgnoreList(ignoreFilePath)
	}
	ignoreList = append(ignoreList, ".git/*", ".gitignore")

	if _, err := os.Stat(gitignoreFilePath); err == nil {
		// .gitignore file exists
		// append .gitignore to ignoreList
		gitignoreList, _ := getIgnoreList(gitignoreFilePath)
		ignoreList = append(ignoreList, gitignoreList...)
	}

	var repoBuilder strings.Builder

	if preambleFile != "" {
		preambleText, err := os.ReadFile(preambleFile)
		if err != nil {
			return "", fmt.Errorf("error reading preamble file: %w", err)
		}
		repoBuilder.WriteString(fmt.Sprintf("%s\n", string(preambleText)))
	} else {
		repoBuilder.WriteString("The following text is a Git repository with code. The structure of the text are sections that begin with ----, followed by a single line containing the file path and file name, followed by a variable amount of lines containing the file contents. The text representing the Git repository ends when the symbols --END-- are encounted. Any further text beyond --END-- are meant to be interpreted as instructions using the aforementioned Git repository as context.\n")
	}

	err := processRepository(repoPath, ignoreList, &repoBuilder)
	if err != nil {
		return "", fmt.Errorf("error processing repository: %w", err)
	}

	repoBuilder.WriteString("--END--")

	return repoBuilder.String(), nil
}

func processRepository(repoPath string, ignoreList []string, repoBuilder *strings.Builder) error {
	return filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relativeFilePath, _ := filepath.Rel(repoPath, path)
			if !shouldIgnore(relativeFilePath, ignoreList) {
				contents, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				repoBuilder.WriteString("----\n")
				repoBuilder.WriteString(fmt.Sprintf("%s\n", relativeFilePath))
				repoBuilder.WriteString(fmt.Sprintf("%s\n", string(contents)))
			}
		}
		return nil
	})
}

func EstimateTokens(output string) int64 {
	tokenCount := float64(len(output))
	// divide by 3.5 to account for the fact that GPT-4 uses (roughly) 3.5 tokens per character
	tokenCount = tokenCount / 3.5
	// round up to the nearest integer
	return int64(math.Ceil(tokenCount))
}
