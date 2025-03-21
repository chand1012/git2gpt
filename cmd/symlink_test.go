package cmd

import (
        "path/filepath"
        "testing"

        "github.com/chand1012/git2gpt/prompt"
)

func TestSymlinkHandling(t *testing.T) {
        // Create a temporary directory for the test
        testDir := "/workspace/test_symlink"

        // Generate an ignore list
        ignoreList := prompt.GenerateIgnoreList(testDir, "", true)

        // Process the repository
        repo, err := prompt.ProcessGitRepo(testDir, ignoreList)
        if err != nil {
                t.Fatalf("Error processing repository with symlink: %v", err)
        }

        // Verify that the repository was processed successfully
        if repo == nil {
                t.Fatal("Repository is nil")
        }

        // Check if the test.txt file was included
        found := false
        for _, file := range repo.Files {
                if file.Path == filepath.Join("storage", "test.txt") {
                        found = true
                        break
                }
        }

        if !found {
                t.Fatal("Expected to find storage/test.txt in the repository")
        }

        // Verify that the symlink was skipped
        for _, file := range repo.Files {
                if file.Path == filepath.Join("public", "storage") {
                        t.Fatal("Symlink should have been skipped")
                }
        }
}
