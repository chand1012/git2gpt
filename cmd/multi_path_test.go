package cmd

import (
       "bytes"
       "io"
       "os"
       "path/filepath"
       "strings"
       "testing"
)

func TestMultiPathInput(t *testing.T) {
       // Create temporary directories for test repositories
       tempDir1, err := os.MkdirTemp("", "repo1")
       if err != nil {
               t.Fatalf("Failed to create temp directory: %v", err)
       }
       defer os.RemoveAll(tempDir1)

       tempDir2, err := os.MkdirTemp("", "repo2")
       if err != nil {
               t.Fatalf("Failed to create temp directory: %v", err)
       }
       defer os.RemoveAll(tempDir2)

       // Create test files in the first repository
       file1Path := filepath.Join(tempDir1, "file1.txt")
       err = os.WriteFile(file1Path, []byte("Content of file1"), 0644)
       if err != nil {
               t.Fatalf("Failed to write test file: %v", err)
       }

       // Create test files in the second repository
       file2Path := filepath.Join(tempDir2, "file2.txt")
       err = os.WriteFile(file2Path, []byte("Content of file2"), 0644)
       if err != nil {
               t.Fatalf("Failed to write test file: %v", err)
       }

       // Capture stdout
       oldStdout := os.Stdout
       r, w, _ := os.Pipe()
       os.Stdout = w

       // Set up command line arguments
       oldArgs := os.Args
       os.Args = []string{"git2gpt", tempDir1, tempDir2}

       // Execute the command
       Execute()

       // Restore stdout and args
       w.Close()
       os.Stdout = oldStdout
       os.Args = oldArgs

       // Read captured output
       var buf bytes.Buffer
       io.Copy(&buf, r)
       output := buf.String()

       // Verify that both files are included in the output
       if !strings.Contains(output, "file1.txt") || !strings.Contains(output, "Content of file1") {
               t.Errorf("Output does not contain content from the first repository")
       }
       if !strings.Contains(output, "file2.txt") || !strings.Contains(output, "Content of file2") {
               t.Errorf("Output does not contain content from the second repository")
       }
}
