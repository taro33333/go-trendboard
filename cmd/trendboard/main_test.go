package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInitCmd is a simple integration test for the 'init' command.
func TestInitCmd(t *testing.T) {
	// Create a temporary directory for the test to run in.
	tempDir := t.TempDir()
	
	// Change working directory to the temp dir
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	// Restore original working directory at the end of the test
	defer os.Chdir(originalWd)

	// Set a dummy token for the test, as config loading requires it.
	t.Setenv("GITHUB_TOKEN", "dummy-token-for-test")

	// Redirect stdout to a buffer to capture output if needed
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	// Execute the init command
	rootCmd.SetArgs([]string{"init"})
	err = rootCmd.Execute()
	require.NoError(t, err)

	// Verify that the repos.json file was created
	reposFilePath := filepath.Join(tempDir, "repos.json")
	_, err = os.Stat(reposFilePath)
	require.NoError(t, err, "repos.json should be created")

	// Verify the content of the file
	content, err := os.ReadFile(reposFilePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "gin-gonic/gin")
	assert.Contains(t, string(content), "spf13/cobra")
}
