//go:build example

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/hymkor/go-htnblog"
)

func list() error {
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
		return err
	}
	for _, entry1 := range entries {
		fmt.Println(entry1.Title)
		fmt.Println(entry1.EditUrl())
	}
	return nil
}

func main() {
	if err := list(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
