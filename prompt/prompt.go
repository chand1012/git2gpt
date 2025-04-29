package prompt

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
	"github.com/chand1012/git2gpt/utils"
	"github.com/gobwas/glob"
	"github.com/pkoukk/tiktoken-go"
)

type GitFile struct {
	Path     string `json:"path" xml:"path"`     // path to the file relative to the repository root
	Tokens   int64  `json:"tokens" xml:"tokens"`   // number of tokens in the file
	Contents string `json:"contents" xml:"contents"` // contents of the file
}

type GitRepo struct {
	TotalTokens int64     `json:"total_tokens" xml:"total_tokens"`
	Files       []GitFile `json:"files" xml:"files>file"`
	FileCount   int       `json:"file_count" xml:"file_count"`
}

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

// Similar to getIgnoreList, but for .gptinclude files
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

func windowsToUnixPath(windowsPath string) string {
	unixPath := strings.ReplaceAll(windowsPath, "\\", "/")
	return unixPath
}

// This function is kept for backward compatibility
func shouldIgnore(filePath string, ignoreList []string) bool {
	for _, pattern := range ignoreList {
		g := glob.MustCompile(pattern, '/')
		if g.Match(windowsToUnixPath(filePath)) {
			return true
		}
	}
	return false
}

// Determines if a file should be included in the output
// First checks if the file matches the include list (if provided)
// Then checks if the file is excluded by the ignore list
func shouldProcess(filePath string, includeList, ignoreList []string) bool {
	// If includeList is provided, check if the file is included
	if len(includeList) > 0 {
		included := false
		for _, pattern := range includeList {
			g := glob.MustCompile(pattern, '/')
			if g.Match(windowsToUnixPath(filePath)) {
				included = true
				break
			}
		}
		if !included {
			return false // If not in the include list, skip it
		}
	}
	
	// Check if the file is excluded by ignoreList
	for _, pattern := range ignoreList {
		g := glob.MustCompile(pattern, '/')
		if g.Match(windowsToUnixPath(filePath)) {
			return false // If in the ignore list, skip it
		}
	}
	
	return true // Process this file
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

// Generate include list from .gptinclude file
func GenerateIncludeList(repoPath, includeFilePath string) []string {
	if includeFilePath == "" {
		includeFilePath = filepath.Join(repoPath, ".gptinclude")
	}
	var includeList []string
	if _, err := os.Stat(includeFilePath); err == nil {
		includeList, _ = getIncludeList(includeFilePath)
	}
	
	var finalIncludeList []string
	for _, pattern := range includeList {
		if !contains(finalIncludeList, pattern) {
			info, err := os.Stat(filepath.Join(repoPath, pattern))
			if err == nil && info.IsDir() {
				pattern = filepath.Join(pattern, "**")
			}
			finalIncludeList = append(finalIncludeList, pattern)
		}
	}
	return finalIncludeList
}

// Update the function signature to accept includeList
func ProcessGitRepo(repoPath string, includeList, ignoreList []string) (*GitRepo, error) {
	var repo GitRepo
	err := processRepository(repoPath, includeList, ignoreList, &repo)
	if err != nil {
		return nil, fmt.Errorf("error processing repository: %w", err)
	}
	return &repo, nil
}

func OutputGitRepo(repo *GitRepo, preambleFile string, scrubComments bool) (string, error) {
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
	for _, file := range repo.Files {
		repoBuilder.WriteString("----\n")
		repoBuilder.WriteString(fmt.Sprintf("%s\n", file.Path))
		if scrubComments {
			file.Contents = utils.RemoveCodeComments(file.Contents)
		}
		repoBuilder.WriteString(fmt.Sprintf("%s\n", file.Contents))
	}
	repoBuilder.WriteString("--END--")
	output := repoBuilder.String()
	repo.TotalTokens = EstimateTokens(output)
	return output, nil
}


func OutputGitRepoXML(repo *GitRepo, scrubComments bool) (string, error) {
	if scrubComments {
		for i, file := range repo.Files {
			repo.Files[i].Contents = utils.RemoveCodeComments(file.Contents)
		}
	}
	var result strings.Builder
	result.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	result.WriteString("<root>\n")
	
	result.WriteString("    <total_tokens>PLACEHOLDER</total_tokens>\n")
	result.WriteString(fmt.Sprintf("    <file_count>%d</file_count>\n", repo.FileCount))
	result.WriteString("    <files>\n")
	
	for _, file := range repo.Files {
		result.WriteString("        <file>\n")
		result.WriteString(fmt.Sprintf("            <path>%s</path>\n", escapeXML(file.Path)))
		result.WriteString(fmt.Sprintf("            <tokens>%d</tokens>\n", file.Tokens))
		
		// Split content around CDATA end marker (]]>) and create multiple CDATA sections
		contents := file.Contents
		result.WriteString("            <contents>")
		
		for {
			idx := strings.Index(contents, "]]>")
			if idx == -1 {
				// No more CDATA end markers, write remaining content in one CDATA section
				result.WriteString("<![CDATA[")
				result.WriteString(contents)
				result.WriteString("]]>")
				break
			}
			
			// Write content up to the CDATA end marker
			result.WriteString("<![CDATA[")
			result.WriteString(contents[:idx+2]) // Include the "]]" part
			result.WriteString("]]>") // Close this CDATA section
			
			// Start a new CDATA section with the ">" character
			result.WriteString("<![CDATA[>")
			
			// Move past the "]]>" in the original content
			contents = contents[idx+3:]
		}
		
		result.WriteString("</contents>\n")
		result.WriteString("        </file>\n")
	}
	
	result.WriteString("    </files>\n")
	result.WriteString("</root>\n")
	
	outputStr := result.String()
	
	tokenCount := EstimateTokens(outputStr)
	repo.TotalTokens = tokenCount
	
	outputStr = strings.Replace(
		outputStr, 
		"<total_tokens>PLACEHOLDER</total_tokens>", 
		fmt.Sprintf("<total_tokens>%d</total_tokens>", tokenCount), 
		1,
	)
	
	return outputStr, nil
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

func ValidateXML(xmlString string) error {
    decoder := xml.NewDecoder(strings.NewReader(xmlString))
    for {
        _, err := decoder.Token()
        if err == io.EOF {
            break
        }
        if err != nil {
            return fmt.Errorf("XML validation error: %w", err)
        }
    }
    return nil
}

func MarshalRepo(repo *GitRepo, scrubComments bool) ([]byte, error) {
	_, err := OutputGitRepo(repo, "", scrubComments)
	if err != nil {
		return nil, fmt.Errorf("error marshalling repo: %w", err)
	}
	return json.Marshal(repo)
}

// Update the function signature to accept includeList and use shouldProcess
func processRepository(repoPath string, includeList, ignoreList []string, repo *GitRepo) error {
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relativeFilePath, _ := filepath.Rel(repoPath, path)
			process := shouldProcess(relativeFilePath, includeList, ignoreList)
			if process {
				contents, err := os.ReadFile(path)
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

func EstimateTokens(output string) int64 {
	tke, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		fmt.Println("Error getting encoding:", err)
		return 0
	}
	tokens := tke.Encode(output, nil, nil)
	return int64(len(tokens))
}