package utils

import (
	"archive/tar"
	"io"
	"testing"
)

func TestAdd(t *testing.T) {
	expectedFileName := "file_name"
	expectedContent := "Hello world!"

	archive := NewInMemoryArchive()
	if archive.Add(expectedFileName, expectedContent) != nil {
		t.Fatalf("Failed to add file %s to archive", expectedFileName)
	}
	buffer, err := archive.Close()
	if err != nil {
		t.Fatalf("Failed to create archive")
	}

	reader := tar.NewReader(buffer)
	hdr, err := reader.Next()
	if err != nil {
		t.Fatalf("Failed to advance to first file in the archive [%s]", err)
	}

	actualContent := make([]byte, hdr.Size)
	size, err := reader.Read(actualContent)
	if err != nil {
		t.Fatalf("Failed to read first file from the archive [%s]", err)
	}
	if int64(size) != hdr.Size {
		t.Fatalf("Read unexpected number of bytes [expected=%d, actual=%d]", hdr.Size, size)
	}
	if hdr.Name != expectedFileName {
		t.Fatalf("Unexpected file name [expected=%s, actual=%s]", expectedFileName, hdr.Name)
	}
	if string(actualContent) != expectedContent {
		t.Fatalf("Invalid content read [expected=%s, actual=%s]", expectedContent, actualContent)
	}

	_, err = reader.Next()
	if err == nil || err != io.EOF {
		t.Fatalf("Expected archive to have only one entry but more found")
	}
}
