package defaulteditor

import (
	"os"
	"path/filepath"
)

func find() string {
	return filepath.Join(os.Getenv("windir"), "system32", "notepad.exe")
}
