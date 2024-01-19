package defaulteditor

import (
	"os"
)

func Find() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	return find()
}
