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

func fields(s string) []string {
	result := make([]string, 0, 2)
	for len(s) > 0 {
		for len(s) > 0 && s[0] == ' ' {
			s = s[1:]
		}
		q := false
		var buffer strings.Builder
		for {
			if len(s) <= 0 {
				if buffer.Len() <= 0 {
					return result
				}
				break
			}
			if !q && s[0] == ' ' {
				break
			}
			if s[0] == '"' {
				q = !q
			} else {
				buffer.WriteByte(s[0])
			}
			s = s[1:]
		}
		result = append(result, buffer.String())
	}
	return result
}

func whichEditor() []string {
	json1, err := config()
	if err != nil {
		return []string{""}
	}
	return fields(json1.Editor)
}

func callEditor(draft []byte) ([]byte, error) {
	editor := whichEditor()
	if len(editor) <= 0 || editor[0] == "" {
		return nil, errors.New(`editor not found. Please set $EDITOR or { "editor":"(YOUR-EDITOR)}" on ~/.htnblog`)
	}
	tempPath := filepath.Join(os.TempDir(), fmt.Sprintf("htnblog-%d.md", os.Getpid()))
	os.WriteFile(tempPath, draft, 0600)
	defer trash.Throw(tempPath)

	args := append(editor[1:], tempPath)
	for {
		cmd := exec.Command(editor[0], args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("%w\n\"%s\" aborted", err, strings.Join(editor, `" "`))
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
