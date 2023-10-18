package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/hymkor/go-htnblog"
)

var flagRcFile = flag.String("rc", "", "use string instead of ~/.htnblog")

var config = sync.OnceValues(func() ([]byte, error) {
	var configPath string
	if *flagRcFile != "" {
		configPath = *flagRcFile
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		configPath = filepath.Join(home, ".htnblog")
	}
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
	for i, entry1 := range entries {
		fmt.Printf("@%d %s %s\n",
			i,
			url2id(entry1.EditUrl()),
			entry1.Title)
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
	draft, err := callEditor([]byte("Title: \n---\n\n"))
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
		if strings.HasPrefix(text, "---") {
			break
		}
		name, value, _ := strings.Cut(text, ":")
		name = strings.ToLower(name)
		header[name] = append(header[name], strings.TrimSpace(value))
	}
	body, err := io.ReadAll(br)
	return header, body, ignoreEof(err)
}

func entryToDraft(entry *htnblog.XmlEntry) []byte {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "Title: %s\n", entry.Title)
	fmt.Fprintf(&buffer, "---\n%s", entry.Content.Body)
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

var (
	rxDateTime = regexp.MustCompile(`^\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d\+\d\d:\d\d$`)
	updated    = flag.String("updated", "", "(hidden option) set update data")
)

func editEntry1(blog *htnblog.Blog, entry *htnblog.XmlEntry) error {
	if *updated != "" {
		if !rxDateTime.MatchString(*updated) {
			return fmt.Errorf("%s: invalid date/time format", *updated)
		}
		entry.Updated = *updated
		println(*updated)
	}
	draft, err := callEditor(entryToDraft(entry))
	if err != nil {
		return err
	}
	if err := draftToEntry(draft, entry); err != nil {
		return err
	}
	return htnblog.Dump(blog.Update(entry))
}

func url2id(url string) string {
	index := strings.LastIndexByte(url, '/')
	if index < 0 {
		return ""
	}
	return url[index+1:]
}

func editEntry(blog *htnblog.Blog, args []string) error {
	entries, err := blog.List()
	if err != nil {
		return err
	}
	if len(entries) <= 0 {
		return errors.New("no entries")
	}
	if len(args) <= 0 {
		return editEntry1(blog, entries[0])
	}
	if len(args[0]) == 2 && args[0][0] == '@' {
		nth := int(args[0][1] - '0')
		if nth >= 0 && nth < len(entries) {
			return editEntry1(blog, entries[nth])
		}
	}
	for _, entry1 := range entries {
		id := url2id(entry1.EditUrl())
		if id != "" && id == args[0] {
			return editEntry1(blog, entry1)
		}
	}
	return fmt.Errorf("%s: entry not found", args[0])
}

var version string

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
		fmt.Fprintf(os.Stderr, "%s %s-%s-%s by %s\n",
			filepath.Base(os.Args[0]),
			version, runtime.GOOS, runtime.GOARCH, runtime.Version())

		io.WriteString(os.Stderr, `
Usage: htnblog {list|new|edit}
  htnblog list ... show recent articles
  htnblog new  ... create new draft
  htnblog edit ... edit the latest article
    The lines in the draft up to "---" are the header lines,
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
		return editEntry(blog, args[1:])
	default:
		return fmt.Errorf("%s: no such subcommand", args[0])
	}
}

func main() {
	flag.Parse()
	if err := mains(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
