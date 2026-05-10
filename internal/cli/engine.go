package cli

import (
	"os"
	"path/filepath"
)

// sprigDir returns the path to ~/.sprig.
func sprigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), ".sprig")
	}
	return filepath.Join(home, ".sprig")
}
