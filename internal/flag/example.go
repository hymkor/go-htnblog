//go:build ignore

package main

import (
	"fmt"

	"github.com/hymkor/go-htnblog/internal/flag"
)

var flagN = flag.Int("n", 10, "fetch articles")

func main() {
	flag.Parse()

	fmt.Printf("*flagN==%#v\n", *flagN)
	fmt.Printf("flag.Args() == %#v\n", flag.Args())
}
