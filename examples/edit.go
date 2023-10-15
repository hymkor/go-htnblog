//go:build example

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/hymkor/go-htnblog"
)

func mains() error {
	auth, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	blog, err := htnblog.NewFromJSON(auth)
	if err != nil {
		return err
	}
	entries, err := blog.List()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	if len(entries) <= 0 {
		return errors.New("no entries")
	}
	entries[0].Content.Body += time.Now().Format("\n編集 2006-01-02 15:04:05")
	return htnblog.Dump(blog.Update(entries[0]))
}

func main() {
	if err := mains(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
