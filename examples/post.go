//go:build example

package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/hymkor/go-htnblog"
)

func post() error {
	auth, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	blog, err := htnblog.NewFromJSON(auth)
	if err != nil {
		return err
	}
	return htnblog.Dump(blog.Post(time.Now().Format("投稿 2006-01-02 15:04:05"), "本文を書く", "yes"))
}

func main() {
	if err := post(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
