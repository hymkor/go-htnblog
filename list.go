package htnblog

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

type xmlFeed struct {
	XMLName  xml.Name    `xml:"feed"`
	XMLNs    string      `xml:"xmlns,attr"`
	XMLNsApp string      `xml:"xmlns:app,attr"`
	Entry    []*XmlEntry `xml:"entry"`
	Link     []XmlLink   `xml:"link"`
	b        *Blog
}

func (feed *xmlFeed) nextUrl() string {
	return findLink("next", feed.Link)
}

func (feed *xmlFeed) ListNext() (*xmlFeed, error) {
	nextUrl := feed.nextUrl()
	if nextUrl == "" {
		return nil, io.EOF
	}
	var nextFeed xmlFeed
	if err := feed.b.get(nextUrl, &nextFeed); err != nil {
		return nil, err
	}
	nextFeed.b = feed.b
	return &nextFeed, nil
}

func (B *Blog) get(url string, v interface{}) error {
	body, err := B.request(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	defer body.Close()
	bin, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}
	err = xml.Unmarshal(bin, v)
	if err != nil {
		return fmt.Errorf("xml.Unmarshal: %w", err)
	}
	return nil
}

func (B *Blog) listFirst() (*xmlFeed, error) {
	var feed xmlFeed
	if err := B.get(B.EndPointUrl+"/entry", &feed); err != nil {
		return nil, err
	}
	feed.b = B
	return &feed, nil
}

func (B *Blog) List() ([]*XmlEntry, error) {
	if f, err := B.listFirst(); err != nil {
		return nil, err
	} else {
		return f.Entry, nil
	}
}

func (B *Blog) EachEntry(callback func(*XmlEntry) bool) error {
	f, err := B.listFirst()
	for err == nil {
		for _, entry := range f.Entry {
			if !callback(entry) {
				return nil
			}
		}
		f, err = f.ListNext()
	}
	if err == io.EOF {
		return nil
	}
	return err
}

func (B *Blog) Index(i int) *XmlEntry {
	f, err := B.listFirst()
	for err == nil {
		if i < len(f.Entry) {
			return f.Entry[i]
		}
		i -= len(f.Entry)
		f, err = f.ListNext()
	}
	return nil
}

func (B *Blog) Get(entryId string) (*XmlEntry, error) {
	var entry XmlEntry
	if err := B.get(B.EndPointUrl+"/entry/"+entryId, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func (B *Blog) Dump(w io.Writer) error {
	body, err := B.request(http.MethodGet, B.EndPointUrl+"/entry", nil)
	if err != nil {
		return err
	}
	defer body.Close()

	_, err = io.Copy(w, body)
	return err
}
