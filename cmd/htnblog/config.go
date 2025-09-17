package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nyaosorg/go-readline-ny"

	"github.com/hymkor/htnblog-go/internal/defaulteditor"
	"github.com/hymkor/htnblog-go/internal/flag"
)

var (
	flagRcFile = flag.String("rc", "", "use the specified file instead of ~/.htnblog")
)

type configuration struct {
	UserId      string `json:"userid"`
	EndPointUrl string `json:"endpointurl"`
	ApiKey      string `json:"apikey"`
	Editor      string `json:"editor"`
	Url1        string `json:"endpointurl1"`
	Url2        string `json:"endpointurl2"`
	Url3        string `json:"endpointurl3"`
}

var getConfigPath = sync.OnceValues(func() (string, error) {
	if *flagRcFile != "" {
		return *flagRcFile, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".htnblog"), nil
})

var config = sync.OnceValues(func() (*configuration, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	bin, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", configPath, err)
	}

	var json1 configuration
	err = json.Unmarshal(bin, &json1)
	return &json1, err
})

func ask(prompt, defaults string) (string, error) {
	editor := &readline.Editor{
		Writer: os.Stderr,
		PromptWriter: func(w io.Writer) (int, error) {
			return io.WriteString(w, prompt)
		},
		Default: defaults,
	}
	answer, err := editor.ReadLine(context.Background())
	answer = strings.TrimSpace(answer)
	return answer, err
}

func nonZeroValue(a, b string) string {
	if a == "" {
		return b
	}
	return a
}

func initConfig() (*configuration, error) {
	json1, err := config()
	if err != nil {
		json1 = &configuration{}
	}
	json1.UserId, err = ask("Hatena-id ? ", json1.UserId)
	if err != nil {
		return nil, err
	}
	json1.ApiKey, err = ask("API-KEY ? ", json1.ApiKey)
	if err != nil {
		return nil, err
	}
	json1.Url1, err = ask("End Point URL 1 ? ", json1.Url1)
	if err != nil {
		return nil, err
	}
	json1.Url2, err = ask("End Point URL 2 ? ", json1.Url2)
	if err != nil {
		return nil, err
	}
	json1.Url3, err = ask("End Point URL 3 ? ", json1.Url3)
	if err != nil {
		return nil, err
	}
	json1.EndPointUrl, err = ask("End Point URL (default) ? ",
		nonZeroValue(json1.EndPointUrl, json1.Url1))
	if err != nil {
		return nil, err
	}
	json1.Editor, err = ask("Text Editor Path ? ",
		nonZeroValue(json1.Editor, defaulteditor.Find()))
	if err != nil {
		return nil, err
	}

	bin, err := json.MarshalIndent(json1, "", "\t")
	if err != nil {
		return nil, err
	}
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	err = writeFileWithBackup(configPath, bin, 0600)
	if err != nil {
		return nil, err
	}
	_, err = fmt.Fprintln(os.Stderr, "Saved configuration to", configPath)
	return json1, err
}
