package desktop

import (
	"fmt"
	"os"
	"path"
	"testing"
)

func TestScan(t *testing.T) {
	dirs, err := getTestScanDirs()
	if err != nil {
		t.Fatalf("failed to get test scan dirs: %s", err)
	}

	entries, err := Scan(dirs)
	if err != nil {
		t.Fatalf("failed to scan %s: %s", dirs[0], err)
	}
	_ = entries
}

func BenchmarkScan(b *testing.B) {
	dirs, err := getTestScanDirs()
	if err != nil {
		b.Fatalf("failed to get test scan dirs: %s", err)
	}

	var entries [][]*Entry
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entries, err = Scan(dirs)
		if err != nil {
			b.Fatal(err)
		}
	}
	_ = entries
}

func getTestScanDirs() ([]string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %s", err)
	}

	return []string{path.Join(wd, "test")}, nil
}
