package main

import (
	"bytes"
	"context"
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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-tty"

	"github.com/nyaosorg/go-readline-ny"

	"github.com/hymkor/go-windows1x-virtualterminal"
	"github.com/hymkor/trash-go"

	"github.com/hymkor/go-htnblog"
)

var (
	flagRcFile = flag.String("rc", "", "use the specified file instead of ~/.htnblog")
	flagMax    = flag.Int("n", 100, "fetch articles")
	flagFirst  = flag.Bool("1", false, "Use the value of \"endpointurl1\" in the JSON setting")
	flagSecond = flag.Bool("2", false, "Use the value of \"endpointurl2\" in the JSON setting")
	flagThrid  = flag.Bool("3", false, "Use the value of \"endpointurl3\" in the JSON setting")
	flagForce  = flag.Bool("f", false, "Delete without prompt")
	flagDebug  = flag.Bool("debug", false, "Enable Debug Output")
)

type configuration struct {
	UserId      string `json:"userid"`
	EndPointUrl string `json:"endpointurl"`
	ApiKey      string `json:"apikey"`
	Editor      string `json:"editor"`
	Url1        string `json:"endpointurl1"`
	Url2        string `json:"endpointurl2"`
	Url3        string `json:"endpointurl3"`
}

var getConfigPath = sync.OnceValues(func() (string, error) {
	if *flagRcFile != "" {
		return *flagRcFile, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".htnblog"), nil
})

var config = sync.OnceValues(func() (*configuration, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	bin, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", configPath, err)
	}

	var json1 configuration
	err = json.Unmarshal(bin, &json1)
	return &json1, err
})

func ask(prompt, defaults string) (string, error) {
	editor := &readline.Editor{
		Writer: os.Stderr,
		PromptWriter: func(w io.Writer) (int, error) {
			return io.WriteString(w, prompt)
		},
		Default: defaults,
	}
	answer, err := editor.ReadLine(context.Background())
	answer = strings.TrimSpace(answer)
	return answer, err
}

func nonZeroValue(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func initConfig() (*configuration, error) {
	json1, err := config()
	if err != nil {
		json1 = &configuration{}
	}
	json1.UserId, err = ask("Hatena-id ? ", json1.UserId)
	if err != nil {
		return nil, err
	}
	json1.ApiKey, err = ask("API-KEY ? ", json1.ApiKey)
	if err != nil {
		return nil, err
	}
	json1.Url1, err = ask("End Point URL 1 ? ", json1.Url1)
	if err != nil {
		return nil, err
	}
	json1.Url2, err = ask("End Point URL 2 ? ", json1.Url2)
	if err != nil {
		return nil, err
	}
	json1.Url3, err = ask("End Point URL 3 ? ", json1.Url3)
	if err != nil {
		return nil, err
	}
	json1.EndPointUrl, err = ask("End Point URL (default) ? ",
		nonZeroValue(json1.EndPointUrl, json1.Url1))
	if err != nil {
		return nil, err
	}
	json1.Editor, err = ask("Text Editor Path ? ",
		nonZeroValue(json1.Editor, os.Getenv("EDITOR"), osDefaultEditor()))
	if err != nil {
		return nil, err
	}

	bin, err := json.MarshalIndent(json1, "", "\t")
	if err != nil {
		return nil, err
	}
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(configPath, bin, 0644)
	if err != nil {
		return nil, err
	}
	_, err = fmt.Fprintln(os.Stderr, "Saved configuration to", configPath)
	return json1, err
}

func list(blog *htnblog.Blog) error {
	i := 0
	return blog.EachEntry(func(entry1 *htnblog.XmlEntry) bool {
		fmt.Printf("@%d %s %s\n",
			i,
			url2id(entry1.EditUrl()),
			entry1.Title)
		i++
		return i < *flagMax
	})
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

func askYesNo() (bool, error) {
	if *flagForce {
		return true, nil
	}
	tty1, err := tty.Open()
	if err != nil {
		return false, err
	}
	defer tty1.Close()

	io.WriteString(os.Stderr, "\nAre you sure (y/n): ")

	key, err := readline.GetKey(tty1)
	fmt.Fprintln(os.Stderr, key)
	if err != nil {
		return false, err
	}
	return key == "y" || key == "Y", nil
}

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

func newEntry(blog *htnblog.Blog) error {
	draft, err := callEditor([]byte("Title: \n---\n\n"))
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(draft)) == 0 {
		return errors.New("your draft is empty. Posting is canceled")
	}
	header, body := splitHeaderAndBody(draft)
	title := header["title"]
	res, err := blog.Post(title, strings.TrimSpace(string(body)), "yes")
	if res != nil {
		fmt.Fprintln(os.Stderr, res.Status)
	}
	return blog.DropResponse(res, err)
}

func chomp(text []byte) []byte {
	if len(text) > 0 && text[len(text)-1] == '\n' {
		text = text[:len(text)-1]
	}
	if len(text) > 0 && text[len(text)-1] == '\r' {
		text = text[:len(text)-1]
	}
	return text
}

func splitHeaderAndBody(source []byte) (map[string]string, []byte) {
	header := map[string]string{}
	for len(source) > 0 {
		var line []byte

		line, source, _ = bytes.Cut(source, []byte{'\n'})
		line = chomp(line)
		if bytes.HasPrefix(line, []byte{'-', '-', '-'}) {
			break
		}
		name, value, _ := strings.Cut(string(line), ":")
		name = strings.ToLower(name)
		value = strings.TrimSpace(value)
		if old, ok := header[name]; ok {
			header[name] = old + " " + value
		} else {
			header[name] = value
		}
	}
	return header, source
}

func entryToDraft(entry *htnblog.XmlEntry) []byte {
	var buffer bytes.Buffer
	fmt.Fprintln(&buffer, "Rem: Alternate-Url:", entry.AlternateUrl())
	fmt.Fprintln(&buffer, "Rem: App-Edited:", entry.AppEdited)
	fmt.Fprintln(&buffer, "Rem: Draft:", entry.Control.Draft)
	fmt.Fprintln(&buffer, "Rem: Edit-Url:", entry.EditUrl())
	fmt.Fprintln(&buffer, "Rem: Published:", entry.Published)
	fmt.Fprintln(&buffer, "Updated:", entry.Updated)
	fmt.Fprintln(&buffer, "Title:", entry.Title)
	fmt.Fprintln(&buffer, "---")
	fmt.Fprint(&buffer, entry.Content.Body)
	return buffer.Bytes()
}

func draftToEntry(draft []byte, entry *htnblog.XmlEntry) error {
	header, body := splitHeaderAndBody(draft)

	if val, ok := header["updated"]; ok && val != "" {
		if _, err := time.Parse(time.RFC3339, val); err != nil {
			return fmt.Errorf("updated: %s: %w", val, err)
		}
	}

	entry.Title = header["title"]
	entry.Updated = header["updated"]

	entry.Content.Body = string(body)
	return nil
}

var rxDateTime = regexp.MustCompile(`^\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d\+\d\d:\d\d$`)

func editEntry1(blog *htnblog.Blog, entry *htnblog.XmlEntry) error {
	draft := entryToDraft(entry)
	for {
		var err error
		draft, err = callEditor(draft)
		if err != nil {
			return err
		}
		err = draftToEntry(draft, entry)
		if err == nil {
			break
		}
		var buffer bytes.Buffer
		buffer.WriteString("Rem: Err: ")
		buffer.WriteString(err.Error())
		buffer.WriteByte('\n')
		buffer.Write(draft)
		draft = buffer.Bytes()
	}
	res, err := blog.Update(entry)
	if res != nil {
		fmt.Fprintln(os.Stderr, res.Status)
	}
	return blog.DropResponse(res, err)
}

func url2id(url string) string {
	index := strings.LastIndexByte(url, '/')
	if index < 0 {
		return ""
	}
	return url[index+1:]
}

func chooseEntry(blog *htnblog.Blog, args []string) (*htnblog.XmlEntry, error) {
	if len(args) <= 0 {
		entry := blog.Index(0)
		if entry == nil {
			return nil, errors.New("no entries")
		}
		return entry, nil
	}
	if len(args[0]) >= 2 && args[0][0] == '@' {
		nth, err := strconv.Atoi(args[0][1:])
		if err == nil {
			entry := blog.Index(nth)
			if entry != nil {
				return entry, nil
			}
		}
	}
	var result *htnblog.XmlEntry
	var err error

	if strings.HasPrefix(args[0], "http") {
		err = blog.EachEntry(func(entry1 *htnblog.XmlEntry) bool {
			url := entry1.AlternateUrl()
			if url != "" && url == args[0] {
				result = entry1
				return false
			}
			return true
		})
	} else {
		err = blog.EachEntry(func(entry1 *htnblog.XmlEntry) bool {
			id := url2id(entry1.EditUrl())
			if id != "" && id == args[0] {
				result = entry1
				return false
			}
			return true
		})
	}

	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("%s: entry not found", args[0])
	}
	return result, nil
}

func editEntry(blog *htnblog.Blog, args []string) error {
	entry, err := chooseEntry(blog, args)
	if err != nil {
		return err
	}
	return editEntry1(blog, entry)
}

func typeEntry(blog *htnblog.Blog, args []string) error {
	entry, err := chooseEntry(blog, args)
	if err != nil {
		return err
	}
	os.Stdout.Write(entryToDraft(entry))
	return nil
}

func publishEntry(blog *htnblog.Blog, args []string) error {
	entry, err := chooseEntry(blog, args)
	if err != nil {
		return err
	}
	entry.Control.Draft = "no"
	res, err := blog.Update(entry)
	if res != nil {
		fmt.Fprintln(os.Stderr, res.Status)
	}
	return blog.DropResponse(res, err)
}

func unpublishEntry(blog *htnblog.Blog, args []string) error {
	entry, err := chooseEntry(blog, args)
	if err != nil {
		return err
	}
	entry.Control.Draft = "yes"
	res, err := blog.Update(entry)
	if res != nil {
		fmt.Fprintln(os.Stderr, res.Status)
	}
	return blog.DropResponse(res, err)
}

func deleteEntry(blog *htnblog.Blog, args []string) error {
	entry, err := chooseEntry(blog, args)
	if err != nil {
		return err
	}
	os.Stdout.Write(entryToDraft(entry))
	ans, err := askYesNo()
	if err != nil {
		return err
	}
	if ans {
		res, err := blog.Delete(entry)
		if res != nil {
			fmt.Fprintln(os.Stderr, res.Status)
		}
		return blog.DropResponse(res, err)
	} else {
		fmt.Fprintln(os.Stderr, "Canceled")
		return nil
	}
}

var version string

func mains(args []string) error {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "%s %s-%s-%s by %s\n",
			filepath.Base(os.Args[0]),
			version, runtime.GOOS, runtime.GOARCH, runtime.Version())

		io.WriteString(os.Stderr, `
Usage: htnblog {options...} {init|list|new|type|edit}
  htnblog init                   ... edit configuration
  htnblog list                   ... show recent articles
  htnblog new                    ... create a new draft
  htnblog type    {URL|@0|@1|..} ... output the article to STDOUT
  htnblog edit    {URL|@0|@1|..} ... edit the article
  htnblog delete  {URL|@0|@1|..} ... output the article to STDOUT and delete it
  htnblog publish {URL|@0|@1|..} ... set false the draft flag of the article
    The lines in the draft up to "---" are the header lines,
    and the rest is the article body.

`)
		flag.PrintDefaults()
		return nil
	}

	if args[0] == "init" {
		_, err := initConfig()
		return err
	}

	json1, err := config()
	if err != nil {
		return err
	}
	endp := json1.EndPointUrl
	if *flagFirst {
		if json1.Url1 == "" {
			return errors.New("-1: field \"endpointurl1\" is not set")
		}
		endp = json1.Url1
	}
	if *flagSecond {
		if json1.Url2 == "" {
			return errors.New("-2: field \"endpointurl2\" is not set")
		}
		endp = json1.Url2
	}
	if *flagThrid {
		if json1.Url3 == "" {
			return errors.New("-3: field \"endpointurl3\" is not set")
		}
		endp = json1.Url3
	}

	blog := &htnblog.Blog{
		UserId:      json1.UserId,
		EndPointUrl: endp,
		ApiKey:      json1.ApiKey,
	}
	if *flagDebug {
		blog.DebugPrint = os.Stderr
	}

	switch args[0] {
	case "list":
		return list(blog)
	case "new":
		return newEntry(blog)
	case "edit":
		return editEntry(blog, args[1:])
	case "type":
		return typeEntry(blog, args[1:])
	case "delete":
		return deleteEntry(blog, args[1:])
	case "publish":
		return publishEntry(blog, args[1:])
	case "unpublish":
		return unpublishEntry(blog, args[1:])
	default:
		return fmt.Errorf("%s: no such subcommand", args[0])
	}
}

func main() {
	if closer, err := virtualterminal.EnableStdout(); err == nil {
		defer closer()
	}

	flag.Parse()
	if err := mains(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
