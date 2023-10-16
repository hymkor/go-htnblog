//go:build example

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/hymkor/go-htnblog"
)

func dump() error {
	auth, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	blog, err := htnblog.NewFromJSON(auth)
	if err != nil {
		return err
	}
	return blog.Dump(os.Stdout)
}

func main() {
	if err := dump(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
