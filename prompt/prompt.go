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

// GitFile is a file in a Git repository
type GitFile struct {
        Path     string `json:"path" xml:"path"`     // path to the file relative to the repository root
        Tokens   int64  `json:"tokens" xml:"tokens"`   // number of tokens in the file
        Contents string `json:"contents" xml:"contents"` // contents of the file
}

// GitRepo is a Git repository
type GitRepo struct {
        TotalTokens int64     `json:"total_tokens" xml:"total_tokens"`
        Files       []GitFile `json:"files" xml:"files>file"`
        FileCount   int       `json:"file_count" xml:"file_count"`
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
                // line = filepath.FromSlash(line)
                ignoreList = append(ignoreList, line)
        }
        return ignoreList, scanner.Err()
}

func windowsToUnixPath(windowsPath string) string {
        unixPath := strings.ReplaceAll(windowsPath, "\\", "/")
        return unixPath
}

func shouldIgnore(filePath string, ignoreList []string) bool {
        for _, pattern := range ignoreList {
                g := glob.MustCompile(pattern, '/')
                if g.Match(windowsToUnixPath(filePath)) {
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
        ignoreList = append(ignoreList, ".git/**", ".gitignore", ".gptignore")

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
                                pattern = filepath.Join(pattern, "**")
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

        // write the files to the repoBuilder here
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
    // Prepare XML content
    if scrubComments {
        for i, file := range repo.Files {
            repo.Files[i].Contents = utils.RemoveCodeComments(file.Contents)
        }
    }
    
    // Add XML header
    var result strings.Builder
    result.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
    
    // Use custom marshaling with proper CDATA for code contents
    result.WriteString("<GitRepo>\n")
    
    // Skip the tokens for now
    result.WriteString("  <total_tokens>PLACEHOLDER</total_tokens>\n")
    result.WriteString(fmt.Sprintf("  <file_count>%d</file_count>\n", repo.FileCount))
    result.WriteString("  <files>\n")
    
    for _, file := range repo.Files {
        result.WriteString("    <file>\n")
        result.WriteString(fmt.Sprintf("      <path>%s</path>\n", escapeXML(file.Path)))
        result.WriteString(fmt.Sprintf("      <tokens>%d</tokens>\n", file.Tokens))
        result.WriteString("      <contents><![CDATA[")
        result.WriteString(file.Contents)
        result.WriteString("]]></contents>\n")
        result.WriteString("    </file>\n")
    }
    
    result.WriteString("  </files>\n")
    result.WriteString("</GitRepo>")
    
    // Get the output string
    outputStr := result.String()
    
    // Calculate tokens
    tokenCount := EstimateTokens(outputStr)
    repo.TotalTokens = tokenCount
    
    // Replace the placeholder with the actual token count
    outputStr = strings.Replace(outputStr, "<total_tokens>PLACEHOLDER</total_tokens>", 
                               fmt.Sprintf("<total_tokens>%d</total_tokens>", tokenCount), 1)
    
    return outputStr, nil
}

// escapeXML escapes XML special characters in a string
func escapeXML(s string) string {
    s = strings.ReplaceAll(s, "&", "&amp;")
    s = strings.ReplaceAll(s, "<", "&lt;")
    s = strings.ReplaceAll(s, ">", "&gt;")
    s = strings.ReplaceAll(s, "\"", "&quot;")
    s = strings.ReplaceAll(s, "'", "&apos;")
    return s
}

// ValidateXML checks if the given XML string is well-formed
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
        // run the output function to get the total tokens
        _, err := OutputGitRepo(repo, "", scrubComments)
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
                // Skip symbolic links to avoid issues with directory symlinks (like Laravel's storage link)
                if info.Mode()&os.ModeSymlink != 0 {
                        return nil
                }
                if !info.IsDir() {
                        relativeFilePath, _ := filepath.Rel(repoPath, path)
                        ignore := shouldIgnore(relativeFilePath, ignoreList)
                        // fmt.Println(relativeFilePath, ignore)
                        if !ignore {
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
        tke, err := tiktoken.GetEncoding("cl100k_base")
        if err != nil {
                fmt.Println("Error getting encoding:", err)
                return 0
        }

        tokens := tke.Encode(output, nil, nil)
        return int64(len(tokens))
}
