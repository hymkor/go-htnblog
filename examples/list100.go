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
	i := 0
	return blog.EachEntry(func(entry1 *htnblog.XmlEntry) bool {
		i++
		fmt.Println(i, entry1.Title)
		fmt.Println(entry1.EditUrl())
		return i < 100
	})
}

func main() {
	if err := list(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
