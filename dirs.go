package desktop

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

// DataDirs returns a slice of directories where desktop entries are stored.
func DataDirs() []string {
	var dataDirs []string

	homeDir := os.Getenv("HOME")
	if strings.TrimSpace(homeDir) == "" {
		homeDir = "~/"
	}
	dataHomeSetting := os.Getenv("XDG_DATA_HOME")
	if dataHomeSetting == "" {
		dataHomeSetting = path.Join(homeDir, ".local/share")
	}
	dataDirs = append(dataDirs, filepath.Join(dataHomeSetting, "applications"))

	dataDirsSetting := strings.Split(os.Getenv("XDG_DATA_DIRS"), ":")
	for _, dataDir := range dataDirsSetting {
		dataDir = strings.TrimSpace(dataDir)
		if dataDir == "" {
			continue
		}

		dataDirs = append(dataDirs, filepath.Join(dataDir, "applications"))
	}
	if len(dataDirs) == 1 {
		dataDirs = append(dataDirs, "/usr/local/share/applications", "/usr/share/applications")
	}

	return dataDirs
}
