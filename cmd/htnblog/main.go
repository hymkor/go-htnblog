package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-tty"
	"github.com/toqueteos/webbrowser"

	"github.com/nyaosorg/go-readline-ny"

	"github.com/hymkor/go-windows1x-virtualterminal"

	"github.com/hymkor/go-htnblog"
	"github.com/hymkor/go-htnblog/internal/flag"
)

var (
	flagMax    = flag.Int("n", 10, "fetch articles")
	flagFirst  = flag.Bool("1", false, "Use the value of \"endpointurl1\" in the JSON setting")
	flagSecond = flag.Bool("2", false, "Use the value of \"endpointurl2\" in the JSON setting")
	flagThrid  = flag.Bool("3", false, "Use the value of \"endpointurl3\" in the JSON setting")
	flagForce  = flag.Bool("f", false, "Delete without prompt")
	flagDebug  = flag.Bool("debug", false, "Enable Debug Output")
)

func list(blog *htnblog.Blog) error {
	i := 0
	iter, err := blog.Iterator()
	if err != nil {
		return err
	}
	for entry1 := range iter {
		var prefix strings.Builder
		if strings.EqualFold(entry1.Control.Draft, "yes") {
			prefix.WriteString("<draft> ")
		}
		if len(entry1.Category) > 0 {
			for _, category := range entry1.Category {
				fmt.Fprintf(&prefix, "[%s]", category.Term)
			}
			prefix.WriteByte(' ')
		}

		fmt.Printf("@%d %s %s%s\n",
			i,
			entry1.EntryId(),
			prefix.String(),
			entry1.Title)
		i++
		if i >= *flagMax {
			break
		}
	}
	return nil
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

func reportUrls(res *http.Response, err error) error {
	if err != nil {
		return err
	}
	defer res.Body.Close()
	bin, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if *flagDebug {
		os.Stderr.Write(bin)
	}
	var entry htnblog.XmlEntry
	err = xml.Unmarshal(bin, &entry)
	if err != nil {
		return err
	}
	if url, err := entry.UrlForBrowserToEdit(); err == nil {
		fmt.Fprintln(os.Stderr, "UrlToEdit:", url)
	}
	if url := entry.AlternateUrl(); url != "" {
		fmt.Fprintln(os.Stderr, "Alternate-Url:", url)
	}
	return nil
}

func newEntry(blog *htnblog.Blog) error {
	var buffer bytes.Buffer
	fmt.Fprintln(&buffer, "Rem: EndPointUrl:", blog.EndPointUrl)
	buffer.WriteString("Category: \nTitle: \n---\n\n")
	draft, err := callEditor(buffer.Bytes())
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(draft)) == 0 {
		return errors.New("your draft is empty. Posting is canceled")
	}
	var entry htnblog.XmlEntry
	if err := draftToEntry(draft, &entry); err != nil {
		return err
	}
	res, err := blog.Add(&entry)
	if res != nil {
		fmt.Fprintln(os.Stderr, res.Status)
	}
	return reportUrls(res, err)
}

func fromStdin(blog *htnblog.Blog) error {
	draft, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	var entry htnblog.XmlEntry
	if err := draftToEntry(draft, &entry); err != nil {
		return err
	}
	res, err := blog.Add(&entry)
	if res != nil {
		fmt.Fprintln(os.Stderr, res.Status)
	}
	return reportUrls(res, err)
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
	fmt.Fprintln(&buffer, "Rem: Url-To-Edit:", entry.UrlToEdit())
	fmt.Fprintln(&buffer, "Rem: Published:", entry.Published)
	fmt.Fprint(&buffer, "Category:")
	for _, c := range entry.Category {
		fmt.Fprintf(&buffer, " %s", c.Term)
	}
	fmt.Fprintln(&buffer)
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

	category := header["category"]
	entry.Category = make([]*htnblog.XmlCategory, 0, 4)
	already := make(map[string]struct{})
	for {
		var term string
		term, category, _ = strings.Cut(category, " ")
		if term != "" {
			termL := strings.ToLower(term)
			if _, ok := already[termL]; !ok {
				entry.Category = append(entry.Category,
					&htnblog.XmlCategory{Term: term})
				already[termL] = struct{}{}
			}
		}
		if category == "" {
			break
		}
	}
	entry.Content.Body = string(body)
	return nil
}

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
	return reportUrls(res, err)
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
	iter, err := blog.Iterator()
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(args[0], "http") {
		for entry1 := range iter {
			url := entry1.AlternateUrl()
			if url != "" && url == args[0] {
				return entry1, nil
			}
		}
	} else {
		for entry1 := range iter {
			id := entry1.EntryId()
			if id != "" && id == args[0] {
				return entry1, nil
			}
		}
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

func browseEntry(blog *htnblog.Blog, args []string) error {
	entry, err := chooseEntry(blog, args)
	if err != nil {
		return err
	}
	url, err := entry.UrlForBrowserToEdit()
	if err != nil {
		return err
	}
	return webbrowser.Open(url)
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
  htnblog browse  {URL|@0|@1|..} ... open the edit page in a web browser

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
	case "from-stdin":
		return fromStdin(blog)
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
	case "browse":
		return browseEntry(blog, args[1:])
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
