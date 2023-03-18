package prompt

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/gobwas/glob"
)

// GitFile is a file in a Git repository
type GitFile struct {
	Path     string `json:"path"`     // path to the file relative to the repository root
	Tokens   int64  `json:"tokens"`   // number of tokens in the file
	Contents string `json:"contents"` // contents of the file
}

// GitRepo is a Git repository
type GitRepo struct {
	TotalTokens int64     `json:"total_tokens"`
	Files       []GitFile `json:"files"`
	FileCount   int       `json:"file_count"`
}

// contains checks if a string is in a slice of strings
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func getIgnoreList(ignoreFilePath string) ([]string, error) {
	var ignoreList []string
	file, err := os.Open(ignoreFilePath)
	if err != nil {
		return ignoreList, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// if the line ends with a slash, add a globstar to the end
		if strings.HasSuffix(line, "/") {
			line = line + "**"
		}
		// remove all preceding slashes
		line = strings.TrimPrefix(line, "/")
		ignoreList = append(ignoreList, line)
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

// GenerateIgnoreList generates a list of ignore patterns from the .gptignore file and the .gitignore file. Returns a slice of strings. Will return an empty slice if no ignore files exist.
func GenerateIgnoreList(repoPath, ignoreFilePath string, useGitignore bool) []string {
	if ignoreFilePath == "" {
		ignoreFilePath = filepath.Join(repoPath, ".gptignore")
	}

	var ignoreList []string
	if _, err := os.Stat(ignoreFilePath); err == nil {
		// .gptignore file exists
		ignoreList, _ = getIgnoreList(ignoreFilePath)
	}
	ignoreList = append(ignoreList, ".git/**", ".gitignore")

	if useGitignore {
		gitignorePath := filepath.Join(repoPath, ".gitignore")
		if _, err := os.Stat(gitignorePath); err == nil {
			// .gitignore file exists
			gitignoreList, _ := getIgnoreList(gitignorePath)
			ignoreList = append(ignoreList, gitignoreList...)
		}
	}

	var finalIgnoreList []string
	// loop through the ignore list and remove any duplicates
	// also check if any pattern is a directory and add a globstar to the end
	for _, pattern := range ignoreList {
		if !contains(finalIgnoreList, pattern) {
			// check if the pattern is a directory
			info, err := os.Stat(filepath.Join(repoPath, pattern))
			if err == nil && info.IsDir() {
				pattern = pattern + "/**"
			}
			finalIgnoreList = append(finalIgnoreList, pattern)
		}
	}

	return finalIgnoreList
}

// ProcessGitRepo processes a Git repository and returns a GitRepo object
func ProcessGitRepo(repoPath string, ignoreList []string) (*GitRepo, error) {

	var repo GitRepo

	err := processRepository(repoPath, ignoreList, &repo)
	if err != nil {
		return nil, fmt.Errorf("error processing repository: %w", err)
	}

	return &repo, nil
}

// OutputGitRepo outputs a Git repository to a text file
func OutputGitRepo(repo *GitRepo, preambleFile string) (string, error) {
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

	// write the files to the repoBuilder here
	for _, file := range repo.Files {
		repoBuilder.WriteString("----\n")
		repoBuilder.WriteString(fmt.Sprintf("%s\n", file.Path))
		repoBuilder.WriteString(fmt.Sprintf("%s\n", file.Contents))
	}

	repoBuilder.WriteString("--END--")

	output := repoBuilder.String()

	repo.TotalTokens = EstimateTokens(output)

	return output, nil
}

func MarshalRepo(repo *GitRepo) ([]byte, error) {
	// run the output function to get the total tokens
	_, err := OutputGitRepo(repo, "")
	if err != nil {
		return nil, fmt.Errorf("error marshalling repo: %w", err)
	}
	return json.Marshal(repo)
}

func processRepository(repoPath string, ignoreList []string, repo *GitRepo) error {
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relativeFilePath, _ := filepath.Rel(repoPath, path)
			if !shouldIgnore(relativeFilePath, ignoreList) {
				contents, err := os.ReadFile(path)
				// if the file is not valid UTF-8, skip it
				if !utf8.Valid(contents) {
					return nil
				}
				if err != nil {
					return err
				}
				var file GitFile
				file.Path = relativeFilePath
				file.Contents = string(contents)
				file.Tokens = EstimateTokens(file.Contents)
				repo.Files = append(repo.Files, file)
			}
		}
		return nil
	})

	repo.FileCount = len(repo.Files)

	if err != nil {
		return fmt.Errorf("error walking the path %q: %w", repoPath, err)
	}

	return nil
}

// EstimateTokens estimates the number of tokens in a string
func EstimateTokens(output string) int64 {
	tokenCount := float64(len(output))
	// divide by 3.5 to account for the fact that GPT-4 uses (roughly) 3.5 tokens per character
	tokenCount = tokenCount / 3.5
	// round up to the nearest integer
	return int64(math.Ceil(tokenCount))
}
