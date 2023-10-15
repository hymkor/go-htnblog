package htnblog

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Blog struct {
	UserId      string
	EndPointUrl string
	ApiKey      string
	Author      string
}

func (B *Blog) request(method, endPointUrl string, r io.Reader) (io.ReadCloser, error) {
	req, err := http.NewRequest(method, endPointUrl, r)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %w", err)
	}
	req.SetBasicAuth(B.UserId, B.ApiKey)
	req.Header.Add("Content-Type", "application/x.atom+xml, application/xml, text/xml, */*")
	var client http.Client
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("(http.Client) Do: %w", err)
	}
	return res.Body, nil
}

func (B *Blog) Post(title, content string) error {
	entry := &XmlEntry{
		Title:  title,
		Author: B.Author,
		Content: XmlContent{
			Type: "text/plain",
			Body: content,
		},
	}
	output, err := entry.Marshal()
	if err != nil {
		return fmt.Errorf("Marshal: %w", err)
	}
	r, err := B.request(http.MethodPost, B.EndPointUrl+"/entry", strings.NewReader(output))
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, r)
	r.Close()
	return nil
}

func (B *Blog) Update(entry *XmlEntry) error {
	output, err := entry.Marshal()
	if err != nil {
		return fmt.Errorf("Marshal: %w", err)
	}
	r, err := B.request(http.MethodPut, entry.EditUrl(), strings.NewReader(output))
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, r)
	r.Close()
	return nil
}
