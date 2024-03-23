The flexible "flag" package
===========================

The standard "flag" package stops parsing just before the first non-flag argument.

This flexible-"flag" package continues parsing even when it finds the first non-flag argument.

For example:

`go run example.go a b c -n 100` is equivalent to `go run example.go -n 100 a b c` when this new "flag" package is used.

```example.go
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
```

**go run example.go a b c -n 100**

```go run example.go a b c -n 100|
*flagN==100
flag.Args() == []string{"a", "b", "c"}
```

**go run example.go -n 100 a b c**

```go run example.go -n 100 a b c|
*flagN==100
flag.Args() == []string{"a", "b", "c"}
```

### Supported functions

```
func Bool(name string, defaults bool, usage string) *bool
func Int(name string, defaults int, usage string) *int
func String(name, defaults, usage string) *string

func Args() []string
func Parse()
func PrintDefaults()
```
