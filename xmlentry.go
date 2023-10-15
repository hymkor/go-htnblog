package htnblog

import (
	"encoding/xml"
)

type xmlEntry struct {
	XMLName  xml.Name   `xml:"entry"`
	XMLNs    string     `xml:"xmlns,attr"`
	XMLNsApp string     `xml:"xmlns:app,attr"`
	Title    string     `xml:"title"`
	Author   string     `xml:"author>name"`
	Content  xmlContent `xml:content"`
	IsDraft  string     `xml:"app:control>app:draft"`
	Link     []xmlLink  `xml:"link"`
}

type xmlContent struct {
	XMLName xml.Name `xml:"content"`
	Body    string   `xml:",cdata"`
	Type    string   `xml:"type,attr"`
}

// <link rel="edit" href="https://blog.hatena.ne.jp/{はてなID}/ブログID}/atom/entry/2500000000"/>

type xmlLink struct {
	XMLName xml.Name `xml:"link"`
	Rel     string   `xml:"rel,attr"`
	Href    string   `xml:"href,attr"`
}

func (entry *xmlEntry) Marshal() (string, error) {
	entry.IsDraft = "yes"
	entry.XMLNs = "http://www.w3.org/2005/Atom"
	entry.XMLNsApp = "http://www.w3.org/2007/app"

	result, err := xml.MarshalIndent(entry, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(result), nil
}

func (entry *xmlEntry) EditUrl() string {
	for _, link := range entry.Link {
		if link.Rel == "edit" {
			return link.Href
		}
	}
	return ""
}
