package logger

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckNotExist(t *testing.T) {
	// Test with existing file
	tmpFile, err := ioutil.TempFile("", "test_exist")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	if checkNotExist(tmpFile.Name()) {
		t.Error("checkNotExist should return false for existing file")
	}

	// Test with non-existing file
	nonExistentFile := "/tmp/non_existent_file_12345"
	if !checkNotExist(nonExistentFile) {
		t.Error("checkNotExist should return true for non-existing file")
	}
}

func TestCheckPermission(t *testing.T) {
	// Create a temporary file
	tmpFile, err := ioutil.TempFile("", "test_permission")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Test with accessible file
	if checkPermission(tmpFile.Name()) {
		t.Error("checkPermission should return false for accessible file")
	}

	// Note: It's difficult to test permission denied scenarios in a portable way
	// without actually creating files with restricted permissions, which might
	// require special privileges or fail on different systems
}

func TestMkDir(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := ioutil.TempDir("", "test_mkdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test creating a new directory
	newDir := filepath.Join(tmpDir, "new_directory")
	err = mkDir(newDir)
	if err != nil {
		t.Errorf("mkDir should succeed: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		t.Error("Directory should have been created")
	}

	// Test creating nested directories
	nestedDir := filepath.Join(tmpDir, "level1", "level2", "level3")
	err = mkDir(nestedDir)
	if err != nil {
		t.Errorf("mkDir should create nested directories: %v", err)
	}

	// Verify nested directory was created
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Error("Nested directory should have been created")
	}
}

func TestIsNotExistMkDir(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := ioutil.TempDir("", "test_isnotexistmkdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with non-existing directory
	newDir := filepath.Join(tmpDir, "new_directory")
	err = isNotExistMkDir(newDir)
	if err != nil {
		t.Errorf("isNotExistMkDir should succeed for non-existing directory: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		t.Error("Directory should have been created")
	}

	// Test with existing directory (should not error)
	err = isNotExistMkDir(newDir)
	if err != nil {
		t.Errorf("isNotExistMkDir should not error for existing directory: %v", err)
	}
}

func TestMustOpen(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := ioutil.TempDir("", "test_mustopen")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test opening a new file
	fileName := "test.log"
	file, err := mustOpen(fileName, tmpDir)
	if err != nil {
		t.Errorf("mustOpen should succeed: %v", err)
	}
	if file == nil {
		t.Error("mustOpen should return a file")
	}
	file.Close()

	// Verify file was created
	expectedPath := filepath.Join(tmpDir, fileName)
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("File should have been created")
	}

	// Test opening same file again (should append)
	file2, err := mustOpen(fileName, tmpDir)
	if err != nil {
		t.Errorf("mustOpen should succeed for existing file: %v", err)
	}
	if file2 == nil {
		t.Error("mustOpen should return a file for existing file")
	}
	file2.Close()
}

func TestMustOpenWithNonExistentDir(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := ioutil.TempDir("", "test_mustopen_parent")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test opening file in non-existent subdirectory
	subDir := filepath.Join(tmpDir, "new_subdir")
	fileName := "test.log"
	file, err := mustOpen(fileName, subDir)
	if err != nil {
		t.Errorf("mustOpen should create directory and succeed: %v", err)
	}
	if file == nil {
		t.Error("mustOpen should return a file")
	}
	file.Close()

	// Verify directory and file were created
	expectedPath := filepath.Join(subDir, fileName)
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("File should have been created in new directory")
	}
}

func TestMustOpenInvalidPath(t *testing.T) {
	// Test with invalid directory path (root directory which typically has restricted permissions)
	fileName := "test.log"
	invalidDir := "/root"
	
	file, err := mustOpen(fileName, invalidDir)
	if err == nil {
		if file != nil {
			file.Close()
		}
		// This might succeed on some systems, so we don't fail the test
		t.Log("mustOpen succeeded on restricted directory (might be normal depending on system)")
	} else {
		// Expected error case
		if file != nil {
			t.Error("mustOpen should return nil file on error")
		}
	}
}

func TestMustOpenFileWriteAndRead(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := ioutil.TempDir("", "test_mustopen_writeread")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	fileName := "test.log"
	
	// Open file and write some content
	file, err := mustOpen(fileName, tmpDir)
	if err != nil {
		t.Errorf("mustOpen should succeed: %v", err)
	}
	defer file.Close()

	testContent := "Hello, World!"
	_, err = file.WriteString(testContent)
	if err != nil {
		t.Errorf("Writing to file should succeed: %v", err)
	}

	// Close and reopen to verify append mode
	file.Close()
	
	file2, err := mustOpen(fileName, tmpDir)
	if err != nil {
		t.Errorf("mustOpen should succeed for existing file: %v", err)
	}
	defer file2.Close()

	moreContent := " Appended content!"
	_, err = file2.WriteString(moreContent)
	if err != nil {
		t.Errorf("Appending to file should succeed: %v", err)
	}

	// Read and verify content
	file2.Close()
	expectedPath := filepath.Join(tmpDir, fileName)
	content, err := ioutil.ReadFile(expectedPath)
	if err != nil {
		t.Errorf("Reading file should succeed: %v", err)
	}

	expectedContent := testContent + moreContent
	if string(content) != expectedContent {
		t.Errorf("File content should be %q, got %q", expectedContent, string(content))
	}
}