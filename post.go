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

func (B *Blog) Post(title, content string) error {
	return B.post(http.MethodPost, B.EndPointUrl+"/entry", title, content)
}

func (B *Blog) post(method, endPointUrl, title, content string) error {
	entry := &xmlEntry{
		Title:   title,
		Author:  B.Author,
		Content: content,
	}
	output, err := entry.Marshal()
	if err != nil {
		return fmt.Errorf("Marshal: %w", err)
	}

	req, err := http.NewRequest(method, endPointUrl, strings.NewReader(output))
	if err != nil {
		return fmt.Errorf("http.NewRequest: %w", err)
	}
	req.SetBasicAuth(B.UserId, B.ApiKey)
	req.Header.Add("Content-Type", "application/x.atom+xml, application/xml, text/xml, */*")
	var client http.Client
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("(http.Client) Do: %w", err)
	}
	defer res.Body.Close()
	io.Copy(os.Stdout, res.Body)
	return nil
}
