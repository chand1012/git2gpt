package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/gobwas/glob"
)

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
		if strings.HasSuffix(line, "/") {
			line = line + "**"
		}
		line = strings.TrimPrefix(line, "/")
		ignoreList = append(ignoreList, line)
	}
	return ignoreList, scanner.Err()
}

func getIncludeList(includeFilePath string) ([]string, error) {
	var includeList []string
	file, err := os.Open(includeFilePath)
	if err != nil {
		return includeList, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasSuffix(line, "/") {
			line = line + "**"
		}
		line = strings.TrimPrefix(line, "/")
		includeList = append(includeList, line)
	}
	return includeList, scanner.Err()
}

func GenerateIgnoreList(repoPath, ignoreFilePath string, useGitignore bool) []string {
	if ignoreFilePath == "" {
		ignoreFilePath = filepath.Join(repoPath, ".gptignore")
	}
	var ignoreList []string
	if _, err := os.Stat(ignoreFilePath); err == nil {
		ignoreList, _ = getIgnoreList(ignoreFilePath)
	}
	ignoreList = append(ignoreList, ".git/**", ".gitignore", ".gptignore", ".gptinclude")
	if useGitignore {
		gitignorePath := filepath.Join(repoPath, ".gitignore")
		if _, err := os.Stat(gitignorePath); err == nil {
			gitignoreList, _ := getIgnoreList(gitignorePath)
			ignoreList = append(ignoreList, gitignoreList...)
		}
	}
	var finalIgnoreList []string
	for _, pattern := range ignoreList {
		if !contains(finalIgnoreList, pattern) {
			info, err := os.Stat(filepath.Join(repoPath, pattern))
			if err == nil && info.IsDir() {
				pattern = filepath.Join(pattern, "**")
			}
			finalIgnoreList = append(finalIgnoreList, pattern)
		}
	}
	return finalIgnoreList
}

func GenerateIncludeList(repoPath, includeFilePath string) []string {
	if includeFilePath == "" {
		includeFilePath = filepath.Join(repoPath, ".gptinclude")
	}
	var includeList []string
	if _, err := os.Stat(includeFilePath); err == nil {
		includeList, _ = getIncludeList(includeFilePath)
	}
	
	fmt.Printf("Raw include list from file: %v\n", includeList)
	
	var finalIncludeList []string
	for _, pattern := range includeList {
		if !contains(finalIncludeList, pattern) {
			fmt.Printf("Processing pattern: %s\n", pattern)
			info, err := os.Stat(filepath.Join(repoPath, pattern))
			fmt.Printf("  Stat result: err=%v, isDir=%v\n", err, err == nil && info != nil && info.IsDir())
			if err == nil && info.IsDir() {
				pattern = filepath.Join(pattern, "**")
				fmt.Printf("  Modified pattern to: %s\n", pattern)
			}
			finalIncludeList = append(finalIncludeList, pattern)
		}
	}
	fmt.Printf("Final include list: %v\n", finalIncludeList)
	return finalIncludeList
}

func windowsToUnixPath(windowsPath string) string {
	unixPath := strings.ReplaceAll(windowsPath, "\\", "/")
	return unixPath
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run debug.go <repo_path>")
		return
	}
	
	repoPath := os.Args[1]
	includePatterns := GenerateIncludeList(repoPath, "")
	ignorePatterns := GenerateIgnoreList(repoPath, "", false)
	
	fmt.Printf("Include patterns: %v\n", includePatterns)
	fmt.Printf("Ignore patterns: %v\n", ignorePatterns)
	
	// Walk through all files in the repository
	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relativeFilePath, _ := filepath.Rel(repoPath, path)
			
			fmt.Printf("\nTesting file: %s\n", relativeFilePath)
			
			// Test include logic
			included := false
			if len(includePatterns) > 0 {
				for _, pattern := range includePatterns {
					g := glob.MustCompile(pattern, '/')
					matches := g.Match(windowsToUnixPath(relativeFilePath))
					fmt.Printf("  Include pattern %s matches: %v\n", pattern, matches)
					if matches {
						included = true
					}
				}
			} else {
				included = true
				fmt.Printf("  No include patterns, would include by default\n")
			}
			
			// Test ignore logic
			ignored := false
			for _, pattern := range ignorePatterns {
				g := glob.MustCompile(pattern, '/')
				matches := g.Match(windowsToUnixPath(relativeFilePath))
				fmt.Printf("  Ignore pattern %s matches: %v\n", pattern, matches)
				if matches {
					ignored = true
				}
			}
			
			finalDecision := included && !ignored
			fmt.Printf("  Final decision: %s (included=%v, ignored=%v)\n", 
				map[bool]string{true: "INCLUDE", false: "EXCLUDE"}[finalDecision], included, ignored)
		}
		return nil
	})
}