package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mattn/go-tty"

	"github.com/nyaosorg/go-readline-ny"

	"github.com/hymkor/trash-go"
)

func askYesNoEdit() (rune, error) {
	tty1, err := tty.Open()
	if err != nil {
		return 0, err
	}
	defer tty1.Close()

	for {
		io.WriteString(os.Stderr, "Are you sure to post ? (y/n/edit): ")
		key, err := readline.GetKey(tty1)
		fmt.Fprintln(os.Stderr, key)
		if err != nil {
			return 0, err
		}
		key = strings.ToLower(key)
		if key[0] == 'y' || key[0] == '\r' {
			return 'y', nil
		}
		if key[0] == 'n' {
			return 'n', nil
		}
		if key[0] == 'e' {
			return 'e', nil
		}
	}
}

func whichEditor() string {
	json1, err := config()
	if err == nil && json1.Editor != "" {
		return json1.Editor
	}
	editor, ok := os.LookupEnv("EDITOR")
	if !ok {
		return ""
	}
	return editor
}

func callEditor(draft []byte) ([]byte, error) {
	editor := whichEditor()
	if editor == "" {
		return nil, errors.New(`editor not found. Please set $EDITOR or { "editor":"(YOUR-EDITOR)}" on ~/.htnblog`)
	}
	tempPath := filepath.Join(os.TempDir(), fmt.Sprintf("htnblog-%d.md", os.Getpid()))
	os.WriteFile(tempPath, draft, 0600)
	defer trash.Throw(tempPath)

	for {
		cmd := exec.Command(editor, tempPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("%w\n\"%s\" aborted", err, editor)
		}
		text, err := os.ReadFile(tempPath)
		if err != nil {
			return nil, err
		}
		key, err := askYesNoEdit()
		if err != nil {
			return nil, err
		}
		if key == 'y' {
			return text, nil
		} else if key == 'n' {
			return nil, errors.New("post is canceled")
		}
	}
}
