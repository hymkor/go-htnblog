package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hymkor/go-htnblog"
)

var config = sync.OnceValues(func() ([]byte, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(home, ".htnblog")
	bin, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", configPath, err)
	}
	return bin, nil
})

func list(blog *htnblog.Blog) error {
	entries, err := blog.List()
	if err != nil {
		return err
	}
	for _, entry1 := range entries {
		fmt.Println(entry1.Title)
		fmt.Println(entry1.EditUrl())
	}
	return nil
}

type jsonEditor struct {
	Editor string `json:"editor"`
}

func whichEditor() string {
	configBin, err := config()
	if err == nil {
		var json1 jsonEditor
		err = json.Unmarshal(configBin, &json1)
		if err == nil && json1.Editor != "" {
			return json1.Editor
		}
	}
	editor, ok := os.LookupEnv("EDITOR")
	if !ok {
		return ""
	}
	return editor
}

func callEditor(draft []byte) ([]byte, error) {
	tempPath := filepath.Join(os.TempDir(), "htnblog-tmp.md")
	os.WriteFile(tempPath, draft, 0600)
	defer os.Remove(tempPath)

	editor := whichEditor()
	if editor == "" {
		return nil, errors.New(`Editor not found. Please set $EDITOR or { "editor":"(YOUR-EDITOR)}" on ~/.htnblog`)
	}
	cmd := exec.Command(editor, tempPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Editor `%s` aborted\n%s", editor, err)
	}
	return os.ReadFile(tempPath)
}

func newEntry(blog *htnblog.Blog) error {
	draft, err := callEditor([]byte{})
	if err != nil {
		return err
	}
	title, body, _ := bytes.Cut(draft, []byte{'\n'})
	return htnblog.Dump(blog.Post(strings.TrimSpace(string(title)), strings.TrimSpace(string(body))))
}

func editEntry(blog *htnblog.Blog) error {
	entries, err := blog.List()
	if err != nil {
		return err
	}
	if len(entries) <= 0 {
		return errors.New("no entries")
	}
	var buffer bytes.Buffer
	buffer.WriteString(entries[0].Title)
	buffer.WriteByte('\n')
	buffer.WriteString(entries[0].Content.Body)

	draft, err := callEditor(buffer.Bytes())
	if err != nil {
		return err
	}
	title, body, _ := bytes.Cut(draft, []byte{'\n'})
	entries[0].Title = strings.TrimSpace(string(title))
	entries[0].Content.Body = strings.TrimSpace(string(body))
	return htnblog.Dump(blog.Update(entries[0]))
}

func mains(args []string) error {
	auth, err := config()
	if err != nil {
		return err
	}
	blog, err := htnblog.NewFromJSON(auth)
	if err != nil {
		return err
	}
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, `Usage: htnblog {list|new|edit}`)
		fmt.Fprintln(os.Stderr, `  htnblog list ... show recent articles`)
		fmt.Fprintln(os.Stderr, `  htnblog new  ... create new draft`)
		fmt.Fprintln(os.Stderr, `  htnblog edit ... edit the latest article`)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, `Please set your editor to $EDITOR`)
		fmt.Fprintln(os.Stderr, ` or { "editor": "YOUR-EDITOR" } on ~/.htnblog`)
		return nil
	}
	switch args[0] {
	case "list":
		return list(blog)
	case "new":
		return newEntry(blog)
	case "edit":
		return editEntry(blog)
	default:
		return fmt.Errorf("%s: no such subcommand", args[0])
	}
}

func main() {
	if err := mains(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
