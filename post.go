package htnblog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Blog struct {
	UserId      string `json:"userid"`
	EndPointUrl string `json:"endpointurl"`
	ApiKey      string `json:"apikey"`
	DebugPrint  io.Writer
}

func NewFromJSON(json1 []byte) (*Blog, error) {
	blog := &Blog{}
	err := json.Unmarshal(json1, blog)
	return blog, err
}

func drop(r io.ReadCloser) error {
	if _, err := io.Copy(io.Discard, r); err != nil {
		r.Close()
		return err
	}
	return r.Close()
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
		drop(res.Body)
		return nil, fmt.Errorf("(http.Client) Do: %w", err)
	}
	if len(res.Status) < 1 || res.Status[0] != '2' {
		drop(res.Body)
		return nil, fmt.Errorf("(http.Client) Do: Status: %s", res.Status)
	}
	return res.Body, nil
}

func (B *Blog) Post(title, content string) (io.ReadCloser, error) {
	entry := &XmlEntry{
		Title: title,
		Content: XmlContent{
			Type: "text/plain",
			Body: content,
		},
		Control: XmlControl{
			Draft: "yes",
		},
	}
	output, err := entry.Marshal()
	if err != nil {
		return nil, fmt.Errorf("Marshal: %w", err)
	}
	return B.request(http.MethodPost, B.EndPointUrl+"/entry", strings.NewReader(output))
}

func (B *Blog) Update(entry *XmlEntry) (io.ReadCloser, error) {
	output, err := entry.Marshal()
	if err != nil {
		return nil, fmt.Errorf("Marshal: %w", err)
	}
	if B.DebugPrint != nil {
		fmt.Fprintln(B.DebugPrint, "<<< SendData >>>")
		io.WriteString(B.DebugPrint, output)
		fmt.Fprintln(B.DebugPrint, "<<< /SendData >>>")
	}
	return B.request(http.MethodPut, entry.EditUrl(), strings.NewReader(output))
}

func Dump(r io.ReadCloser, err error) error {
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, r)
	return r.Close()
}
