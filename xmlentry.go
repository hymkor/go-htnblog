package htnblog

import (
	"encoding/xml"
)

type XmlEntry struct {
	XMLName   xml.Name   `xml:"entry"`
	XMLNs     string     `xml:"xmlns,attr"`
	XMLNsApp  string     `xml:"xmlns:app,attr"`
	Title     string     `xml:"title"`
	Content   XmlContent `xml:"content"`
	Link      []XmlLink  `xml:"link"`
	Updated   string     `xml:"updated,omitempty"`
	Published string     `xml:"published,omitempty"`
	AppEdited string     `xml:"app:edited,omitempty"`
	Control   XmlControl `xml:"http://www.w3.org/2007/app control"`
}

type XmlControl struct {
	XMLName xml.Name `xml:"http://www.w3.org/2007/app control"`
	Draft   string   `xml:"http://www.w3.org/2007/app draft,omitempty"`
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

func findLink(rel string, links []XmlLink) string {
	for _, link := range links {
		if link.Rel == rel {
			return link.Href
		}
	}
	return ""
}

func (entry *XmlEntry) EditUrl() string {
	return findLink("edit", entry.Link)
}

func (entry *XmlEntry) AlternateUrl() string {
	return findLink("alternate", entry.Link)
}
