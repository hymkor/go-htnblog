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
}

func (feed *xmlFeed) NextUrl() string {
	return findLink("next", feed.Link)
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

func (B *Blog) List() ([]*XmlEntry, error) {
	var feed xmlFeed
	if err := B.get(B.EndPointUrl+"/entry", &feed); err != nil {
		return nil, err
	}
	return feed.Entry, nil
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
