package prompt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGptIncludeAndIgnore(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "git2gpt-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []struct {
		path     string
		contents string
	}{
		{"file1.txt", "Content of file1"},
		{"file2.txt", "Content of file2"},
		{"file3.txt", "Content of file3"},
		{"src/main.go", "package main\nfunc main() {}"},
		{"src/lib/util.go", "package lib\nfunc Util() {}"},
		{"docs/README.md", "# Documentation"},
	}

	for _, tf := range testFiles {
		fullPath := filepath.Join(tempDir, tf.path)
		// Create directory if it doesn't exist
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		// Write the file
		if err := os.WriteFile(fullPath, []byte(tf.contents), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", fullPath, err)
		}
	}

	// Test cases
	testCases := []struct {
		name           string
		includeContent string
		ignoreContent  string
		expectedFiles  []string
		unexpectedFiles []string
	}{
		{
			name:           "Only include src directory",
			includeContent: "src/**",
			ignoreContent:  "",
			expectedFiles:  []string{"src/main.go", "src/lib/util.go"},
			unexpectedFiles: []string{"file1.txt", "file2.txt", "file3.txt", "docs/README.md"},
		},
		{
			name:           "Include all, but ignore .txt files",
			includeContent: "**",
			ignoreContent:  "*.txt",
			expectedFiles:  []string{"src/main.go", "src/lib/util.go", "docs/README.md"},
			unexpectedFiles: []string{"file1.txt", "file2.txt", "file3.txt"},
		},
		{
			name:           "Include src and docs, but ignore lib directory",
			includeContent: "src/**\ndocs/**",
			ignoreContent:  "src/lib/**",
			expectedFiles:  []string{"src/main.go", "docs/README.md"},
			unexpectedFiles: []string{"file1.txt", "file2.txt", "file3.txt", "src/lib/util.go"},
		},
		{
			name:           "No include file (should include all), ignore .txt files",
			includeContent: "",
			ignoreContent:  "*.txt",
			expectedFiles:  []string{"src/main.go", "src/lib/util.go", "docs/README.md"},
			unexpectedFiles: []string{"file1.txt", "file2.txt", "file3.txt"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create .gptinclude file if needed
			includeFilePath := filepath.Join(tempDir, ".gptinclude")
			if tc.includeContent != "" {
				if err := os.WriteFile(includeFilePath, []byte(tc.includeContent), 0644); err != nil {
					t.Fatalf("Failed to write .gptinclude file: %v", err)
				}
			} else {
				// Ensure no .gptinclude file exists
				os.Remove(includeFilePath)
			}

			// Create .gptignore file if needed
			ignoreFilePath := filepath.Join(tempDir, ".gptignore")
			if tc.ignoreContent != "" {
				if err := os.WriteFile(ignoreFilePath, []byte(tc.ignoreContent), 0644); err != nil {
					t.Fatalf("Failed to write .gptignore file: %v", err)
				}
			} else {
				// Ensure no .gptignore file exists
				os.Remove(ignoreFilePath)
			}

			// Generate include and ignore lists
			includeList := GenerateIncludeList(tempDir, "")
			ignoreList := GenerateIgnoreList(tempDir, "", false)

			// Process the repository
			repo, err := ProcessGitRepo(tempDir, includeList, ignoreList)
			if err != nil {
				t.Fatalf("Failed to process repository: %v", err)
			}

			// Check if expected files are included
			for _, expectedFile := range tc.expectedFiles {
				found := false
				for _, file := range repo.Files {
					if file.Path == expectedFile {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected file %s to be included, but it wasn't", expectedFile)
				}
			}

			// Check if unexpected files are excluded
			for _, unexpectedFile := range tc.unexpectedFiles {
				for _, file := range repo.Files {
					if file.Path == unexpectedFile {
						t.Errorf("File %s should have been excluded, but it was included", unexpectedFile)
						break
					}
				}
			}
		})
	}
}