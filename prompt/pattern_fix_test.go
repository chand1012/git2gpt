package prompt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtensionPatternFix(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "git2gpt-pattern-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []struct {
		path     string
		contents string
	}{
		{"file.go", "package main"},                    // Root level .go file
		{"config.json", `{"test": true}`},              // Root level .json file
		{"README.md", "# Root readme"},                 // Root level .md file
		{"src/main.go", "package main\nfunc main(){}"},// Nested .go file
		{"src/config.json", `{"nested": true}`},        // Nested .json file
		{"docs/api.md", "# API docs"},                  // Nested .md file
		{"deep/nested/file.go", "package deep"},        // Deeply nested .go file
	}

	for _, tf := range testFiles {
		fullPath := filepath.Join(tempDir, tf.path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(tf.contents), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", fullPath, err)
		}
	}

	// Test case: Include all .md files (both root and nested), ignore all .go files (both root and nested)
	t.Run("Extension patterns should match both root and nested files", func(t *testing.T) {
		// Create .gptinclude with *.md pattern
		includeFilePath := filepath.Join(tempDir, ".gptinclude")
		if err := os.WriteFile(includeFilePath, []byte("*.md"), 0644); err != nil {
			t.Fatalf("Failed to write .gptinclude file: %v", err)
		}

		// Create .gptignore with *.go pattern  
		ignoreFilePath := filepath.Join(tempDir, ".gptignore")
		if err := os.WriteFile(ignoreFilePath, []byte("*.go"), 0644); err != nil {
			t.Fatalf("Failed to write .gptignore file: %v", err)
		}

		// Generate include and ignore lists
		includeList := GenerateIncludeList(tempDir, "")
		ignoreList := GenerateIgnoreList(tempDir, "", false)

		// Process the repository
		repo, err := ProcessGitRepo(tempDir, includeList, ignoreList)
		if err != nil {
			t.Fatalf("Failed to process repository: %v", err)
		}

		// Expected: Only .md files should be included (both root and nested)
		expectedFiles := []string{"README.md", "docs/api.md"}
		unexpectedFiles := []string{"file.go", "src/main.go", "deep/nested/file.go", "config.json", "src/config.json"}

		// Check expected files are included
		for _, expectedFile := range expectedFiles {
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

		// Check unexpected files are excluded
		for _, unexpectedFile := range unexpectedFiles {
			for _, file := range repo.Files {
				if file.Path == unexpectedFile {
					t.Errorf("File %s should have been excluded, but it was included", unexpectedFile)
					break
				}
			}
		}
	})

	// Test case: Ensure existing patterns with ** still work
	t.Run("Existing ** patterns should continue to work", func(t *testing.T) {
		// Create .gptinclude with existing **/*.json pattern (should not be modified)
		includeFilePath := filepath.Join(tempDir, ".gptinclude")
		if err := os.WriteFile(includeFilePath, []byte("**/*.json"), 0644); err != nil {
			t.Fatalf("Failed to write .gptinclude file: %v", err)
		}

		// Remove .gptignore for this test
		ignoreFilePath := filepath.Join(tempDir, ".gptignore")
		os.Remove(ignoreFilePath)

		// Generate include and ignore lists
		includeList := GenerateIncludeList(tempDir, "")
		ignoreList := GenerateIgnoreList(tempDir, "", false)

		// Process the repository
		repo, err := ProcessGitRepo(tempDir, includeList, ignoreList)
		if err != nil {
			t.Fatalf("Failed to process repository: %v", err)
		}

		// Expected: Only nested .json files should be included (not root level)
		expectedFiles := []string{"src/config.json"}
		unexpectedFiles := []string{"config.json", "README.md", "file.go", "src/main.go"}

		// Check expected files are included
		for _, expectedFile := range expectedFiles {
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

		// Check unexpected files are excluded  
		for _, unexpectedFile := range unexpectedFiles {
			for _, file := range repo.Files {
				if file.Path == unexpectedFile {
					t.Errorf("File %s should have been excluded, but it was included", unexpectedFile)
					break
				}
			}
		}
	})
}