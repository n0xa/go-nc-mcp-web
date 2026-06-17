package nextcloud_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func TestWebDAVFullLifecycle(t *testing.T) {
	nc := loadTestEnv(t)

	testDir := fmt.Sprintf("ncmcp-test-%d", time.Now().UnixNano())

	// Create a directory
	err := nc.FilesMkdir(testDir)
	if err != nil {
		t.Fatalf("FilesMkdir(%s): %v", testDir, err)
	}
	t.Logf("created directory %s", testDir)

	// List root — should contain our directory
	files, err := nc.FilesList("")
	if err != nil {
		t.Fatalf("FilesList(''): %v", err)
	}
	found := false
	for _, f := range files {
		if f.Name == testDir {
			found = true
			if !f.IsDir {
				t.Errorf("expected %s to be a directory", testDir)
			}
			break
		}
	}
	if !found {
		t.Errorf("created directory %s not found in root listing (%d files)", testDir, len(files))
	}

	// Write a file
	content := []byte("Hello from go-nc-mcp integration tests!\n")
	filePath := testDir + "/test.txt"
	err = nc.FilesWrite(filePath, content)
	if err != nil {
		t.Fatalf("FilesWrite(%s): %v", filePath, err)
	}
	t.Logf("wrote file %s (%d bytes)", filePath, len(content))

	// Read the file back
	readBack, err := nc.FilesRead(filePath)
	if err != nil {
		t.Fatalf("FilesRead(%s): %v", filePath, err)
	}
	if !bytes.Equal(readBack, content) {
		t.Errorf("FilesRead content mismatch: got %q, want %q", string(readBack), string(content))
	}

	// List directory contents
	dirFiles, err := nc.FilesList(testDir)
	if err != nil {
		t.Fatalf("FilesList(%s): %v", testDir, err)
	}
	if len(dirFiles) == 0 {
		t.Error("FilesList returned 0 files in test directory")
	}
	foundFile := false
	for _, f := range dirFiles {
		if f.Name == "test.txt" {
			foundFile = true
			if f.Size != int64(len(content)) {
				t.Errorf("file size = %d, want %d", f.Size, len(content))
			}
			break
		}
	}
	if !foundFile {
		t.Error("test.txt not found in directory listing")
	}

	// Move the file
	movedPath := testDir + "/moved.txt"
	err = nc.FilesMove(filePath, movedPath)
	if err != nil {
		t.Fatalf("FilesMove(%s, %s): %v", filePath, movedPath, err)
	}
	t.Logf("moved %s -> %s", filePath, movedPath)

	// Verify move — old path should fail
	_, err = nc.FilesRead(filePath)
	if err == nil {
		t.Error("expected error reading moved file at old path, got nil")
	}

	// New path should work
	readBack, err = nc.FilesRead(movedPath)
	if err != nil {
		t.Fatalf("FilesRead(%s) after move: %v", movedPath, err)
	}
	if !bytes.Equal(readBack, content) {
		t.Errorf("content at new path mismatch after move")
	}

	// Search for the file (Nextcloud uses substring matching)
	results, err := nc.FilesSearch("moved")
	if err != nil {
		t.Fatalf("FilesSearch: %v", err)
	}
	foundSearch := false
	for _, f := range results {
		if f.Name == "moved.txt" {
			foundSearch = true
			break
		}
	}
	if !foundSearch {
		t.Logf("FilesSearch('moved') returned %d results (file may not be indexed yet)", len(results))
	}

	// Delete the file
	err = nc.FilesDelete(movedPath)
	if err != nil {
		t.Fatalf("FilesDelete(%s): %v", movedPath, err)
	}
	t.Logf("deleted %s", movedPath)

	// Delete the directory
	err = nc.FilesDelete(testDir)
	if err != nil {
		t.Fatalf("FilesDelete(%s): %v", testDir, err)
	}
	t.Logf("deleted directory %s", testDir)

	// Verify deletion
	_, err = nc.FilesRead(testDir + "/anything")
	if err == nil {
		t.Error("expected error reading from deleted directory")
	}
}

func TestWebDAVWriteLargeFile(t *testing.T) {
	nc := loadTestEnv(t)

	// Write a 1MB file to test chunking/limits
	content := bytes.Repeat([]byte("ABCDEFGHIJ"), 100000) // 1MB
	path := "ncmcp-large-test.bin"

	err := nc.FilesWrite(path, content)
	if err != nil {
		t.Fatalf("FilesWrite (1MB): %v", err)
	}

	readBack, err := nc.FilesRead(path)
	if err != nil {
		t.Fatalf("FilesRead (1MB): %v", err)
	}
	if len(readBack) != len(content) {
		t.Errorf("large file size mismatch: got %d bytes, want %d", len(readBack), len(content))
	}

	// Cleanup
	err = nc.FilesDelete(path)
	if err != nil {
		t.Errorf("cleanup FilesDelete: %v", err)
	}
}
