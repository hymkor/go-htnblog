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
}

func (B *Blog) List() ([]*XmlEntry, error) {
	body, err := B.request(http.MethodGet, B.EndPointUrl+"/entry", nil)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	feedBin, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}
	var feed xmlFeed
	err = xml.Unmarshal(feedBin, &feed)
	if err != nil {
		return nil, fmt.Errorf("xml.Unmarshal: %w", err)
	}
	return feed.Entry, nil
}
