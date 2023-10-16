package htnblog

import (
	"encoding/xml"
)

type XmlEntry struct {
	XMLName  xml.Name   `xml:"entry"`
	XMLNs    string     `xml:"xmlns,attr"`
	XMLNsApp string     `xml:"xmlns:app,attr"`
	Title    string     `xml:"title"`
	Author   string     `xml:"author>name"`
	Content  XmlContent `xml:content"`
	IsDraft  string     `xml:"app:control>app:draft,omitempty"`
	Link     []XmlLink  `xml:"link"`
}

type XmlContent struct {
	XMLName xml.Name `xml:"content"`
	Body    string   `xml:",cdata"`
	Type    string   `xml:"type,attr"`
}

// <link rel="edit" href="https://blog.hatena.ne.jp/{はてなID}/ブログID}/atom/entry/2500000000"/>

type XmlLink struct {
	XMLName xml.Name `xml:"link"`
	Rel     string   `xml:"rel,attr"`
	Href    string   `xml:"href,attr"`
}

func (entry *XmlEntry) Marshal() (string, error) {
	entry.XMLNs = "http://www.w3.org/2005/Atom"
	entry.XMLNsApp = "http://www.w3.org/2007/app"

	result, err := xml.MarshalIndent(entry, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(result), nil
}

func (entry *XmlEntry) EditUrl() string {
	for _, link := range entry.Link {
		if link.Rel == "edit" {
			return link.Href
		}
	}
	return ""
}
