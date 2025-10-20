package support

import (
	"os"
	"path/filepath"
	"testing"
)

func LoadFile(t *testing.T, filename string) []byte {
	t.Helper()

	path := filepath.Join("testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read test file %s: %v", path, err)
	}
	
	return data
}
