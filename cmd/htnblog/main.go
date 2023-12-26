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
	"strconv"
	"strings"
	"sync"

	"github.com/hymkor/go-windows1x-virtualterminal"
	"github.com/hymkor/go-windows1x-virtualterminal/keyin"
	"github.com/hymkor/trash-go"

	"github.com/hymkor/go-htnblog"
)

var (
	flagRcFile  = flag.String("rc", "", "use the specified file instead of ~/.htnblog")
	flagMax     = flag.Int("n", 100, "fetch articles")
	flagUpdated = flag.String("updated", "", "(experimental) set the updated date like 2006-01-02T15:04:05-07:00")
	flagFirst   = flag.Bool("1", false, "Use the value of \"endpointurl1\" in the JSON setting")
	flagSecond  = flag.Bool("2", false, "Use the value of \"endpointurl2\" in the JSON setting")
	flagThrid   = flag.Bool("3", false, "Use the value of \"endpointurl3\" in the JSON setting")
	flagForce   = flag.Bool("f", false, "Delete without prompt")
)

type jsonEditor struct {
	UserId      string `json:"userid"`
	EndPointUrl string `json:"endpointurl"`
	ApiKey      string `json:"apikey"`
	Editor      string `json:"editor"`
	Url1        string `json:"endpointurl1"`
	Url2        string `json:"endpointurl2"`
	Url3        string `json:"endpointurl3"`
}

var config = sync.OnceValues(func() (*jsonEditor, error) {
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

	var json1 jsonEditor
	err = json.Unmarshal(bin, &json1)
	return &json1, err
})

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

func askYesNo() bool {
	if *flagForce {
		return true
	}
	if closer, err := keyin.Raw(); err == nil {
		defer closer()
	} else {
		return false
	}
	io.WriteString(os.Stdout, "\nAre you sure (Yes/[No]): ")
	key, err := keyin.Get()
	if err != nil {
		return false
	}
	return key == "y" || key == "Y"
}

func askYesNoEdit() (rune, error) {
	if closer, err := keyin.Raw(); err == nil {
		defer closer()
	}
	for {
		io.WriteString(os.Stdout, "Are you sure to post ? ([Yes]/No/Edit): ")
		key, err := keyin.Get()
		if err != nil {
			return 0, err
		}
		fmt.Println(key)
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
		return nil, errors.New(`Editor not found. Please set $EDITOR or { "editor":"(YOUR-EDITOR)}" on ~/.htnblog`)
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
		return errors.New("Your draft is empty. Posting is canceled.")
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
	// fmt.Fprintf(&buffer, "Draft: %s\n", entry.Control.Draft)
	fmt.Fprintf(&buffer, "---\n%s", entry.Content.Body)
	return buffer.Bytes()
}

func draftToEntry(draft []byte, entry *htnblog.XmlEntry) error {
	header, body, err := splitHeaderAndBody(bytes.NewReader(draft))
	if err != nil {
		return err
	}
	entry.Title = strings.Join(header["title"], " ")
	// entry.Control.Draft = strings.Join(header["draft"], " ")
	entry.Content.Body = string(body)
	return nil
}

var rxDateTime = regexp.MustCompile(`^\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d\+\d\d:\d\d$`)

func editEntry1(blog *htnblog.Blog, entry *htnblog.XmlEntry) error {
	if *flagUpdated != "" {
		if !rxDateTime.MatchString(*flagUpdated) {
			return fmt.Errorf("%s: invalid date/time format", *flagUpdated)
		}
		entry.Updated = *flagUpdated
		println(*flagUpdated)
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

func deleteEntry(blog *htnblog.Blog, args []string) error {
	entry, err := chooseEntry(blog, args)
	if err != nil {
		return err
	}
	os.Stdout.Write(entryToDraft(entry))
	if askYesNo() {
		fmt.Println("\n-> Deleted")
		return blog.Delete(entry.EditUrl())
	} else {
		fmt.Println("\n-> Canceled")
		return nil
	}
}

var version string

func mains(args []string) error {
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
		DebugPrint:  os.Stderr,
	}

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "%s %s-%s-%s by %s\n",
			filepath.Base(os.Args[0]),
			version, runtime.GOOS, runtime.GOARCH, runtime.Version())

		io.WriteString(os.Stderr, `
Usage: htnblog {options...} {list|new|type|edit}
  htnblog list                  ... show recent articles
  htnblog new                   ... create a new draft
  htnblog type   {URL|@0|@1|..} ... output the article to STDOUT
  htnblog edit   {URL|@0|@1|..} ... edit the article
  htnblog delete {URL|@0|@1|..} ... output the article to STDOUT and delete it
    The lines in the draft up to "---" are the header lines,
    and the rest is the article body.

Please write your setting on ~/.htnblog as below:
    {
        "userid":"(YOUR_USER_ID)",
        "endpointurl":"(END_POINT_URL used by default)",
        "apikey":"(YOUR API KEY)",
        "editor":"(YOUR EDITOR.THIS IS for cmd/htnblog/main.go)"
        "endpointurl1":"(END_POINT_URL used by option -1)",
        "endpointurl2":"(END_POINT_URL used by option -2)",
        "endpointurl3":"(END_POINT_URL used by option -3)",
    }

`)
		flag.PrintDefaults()
		return nil
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
	default:
		return fmt.Errorf("%s: no such subcommand", args[0])
	}
}

func main() {
	if closer, err := virtualterminal.EnableStdin(); err == nil {
		defer closer()
	}

	flag.Parse()
	if err := mains(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
