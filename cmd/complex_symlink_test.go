package cmd

import (
       "path/filepath"
       "testing"

       "github.com/chand1012/git2gpt/prompt"
)

func TestComplexSymlinkHandling(t *testing.T) {
       // Create a temporary directory for the test
       testDir := "/workspace/test_complex_symlink"

       // Generate an ignore list
       ignoreList := prompt.GenerateIgnoreList(testDir, "", true)

       // Process the repository
       repo, err := prompt.ProcessGitRepo(testDir, ignoreList)
       if err != nil {
               t.Fatalf("Error processing repository with symlinks: %v", err)
       }

       // Verify that the repository was processed successfully
       if repo == nil {
               t.Fatal("Repository is nil")
       }

       // Check if the regular files were included
       file1Found := false
       file2Found := false
       for _, file := range repo.Files {
               if file.Path == filepath.Join("dir1", "file1.txt") {
                       file1Found = true
               }
               if file.Path == filepath.Join("dir2", "file2.txt") {
                       file2Found = true
               }
       }

       if !file1Found {
               t.Fatal("Expected to find dir1/file1.txt in the repository")
       }
       if !file2Found {
               t.Fatal("Expected to find dir2/file2.txt in the repository")
       }

       // Verify that the symlinks were included
       link1Found := false
       link2Found := false
       for _, file := range repo.Files {
               if file.Path == filepath.Join("public", "link1") {
                       link1Found = true
                       if file.Contents != "../dir1" {
                               t.Fatalf("Expected link1 content to be '../dir1', got '%s'", file.Contents)
                       }
               }
               if file.Path == filepath.Join("public", "link2.txt") {
                       link2Found = true
                       if file.Contents != "../dir2/file2.txt" {
                               t.Fatalf("Expected link2.txt content to be '../dir2/file2.txt', got '%s'", file.Contents)
                       }
               }
       }

       if !link1Found {
               t.Fatal("Expected to find public/link1 in the repository")
       }
       if !link2Found {
               t.Fatal("Expected to find public/link2.txt in the repository")
       }
}
