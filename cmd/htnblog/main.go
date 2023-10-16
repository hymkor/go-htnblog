package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	draft, err := callEditor([]byte("Title: \n\n\n"))
	if err != nil {
		return err
	}

	header, body, err := splitHeaderAndBody(bytes.NewReader(draft))
	if err != nil {
		return err
	}
	title := strings.Join(header["title"], " ")
	return htnblog.Dump(blog.Post(title, strings.TrimSpace(string(body))))
}

func ignoreEof(err error) error {
	if err == io.EOF {
		return nil
	}
	return err
}

func chomp(text string) string {
	if len(text) > 0 && text[len(text)-1] == '\n' {
		text = text[:len(text)-1]
	}
	if len(text) > 0 && text[len(text)-1] == '\r' {
		text = text[:len(text)-1]
	}
	return text
}

func splitHeaderAndBody(r io.Reader) (map[string][]string, []byte, error) {
	br := bufio.NewReader(r)
	header := map[string][]string{}
	for {
		text, err := br.ReadString('\n')
		if err != nil {
			return header, []byte{}, ignoreEof(err)
		}
		text = chomp(text)
		if text == "" {
			break
		}
		name, value, _ := strings.Cut(text, ": ")
		name = strings.ToLower(name)
		header[name] = append(header[name], value)
	}
	body, err := io.ReadAll(br)
	return header, body, ignoreEof(err)
}

func entryToDraft(entry *htnblog.XmlEntry) []byte {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "Title: %s\n", entry.Title)
	fmt.Fprintf(&buffer, "\n%s", entry.Content.Body)
	return buffer.Bytes()
}

func draftToEntry(draft []byte, entry *htnblog.XmlEntry) error {
	header, body, err := splitHeaderAndBody(bytes.NewReader(draft))
	if err != nil {
		return err
	}
	entry.Title = strings.Join(header["title"], " ")
	entry.Content.Body = string(body)
	return nil
}

func editEntry(blog *htnblog.Blog) error {
	entries, err := blog.List()
	if err != nil {
		return err
	}
	if len(entries) <= 0 {
		return errors.New("no entries")
	}

	draft, err := callEditor(entryToDraft(entries[0]))
	if err != nil {
		return err
	}
	if err := draftToEntry(draft, entries[0]); err != nil {
		return err
	}
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
		io.WriteString(os.Stderr,
			`Usage: htnblog {list|new|edit}
  htnblog list ... show recent articles
  htnblog new  ... create new draft
  htnblog edit ... edit the latest article
    The lines in the draft up to the first blank line are the header lines,
    and the rest is the article body.

Please write your setting on ~/.htnblog as below:
    {
        "userid":"(YOUR_USER_ID)",
        "endpointurl":"(END_POINT_URL)",
        "apikey":"(YOUR API KEY)",
        "author":"(YOUR NAME)",
        "editor":"(YOUR EDITOR.THIS IS for cmd/htnblog/main.go)"
    }
`)
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
