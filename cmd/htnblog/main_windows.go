package main

import (
	"os"
	"path/filepath"
	"sync"
)

var osDefaultEditor = sync.OnceValue(func() string {
	return filepath.Join(os.Getenv("windir"), "system32", "notepad.exe")
})
