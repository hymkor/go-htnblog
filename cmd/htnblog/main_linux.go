package main

import (
	"sync"
)

var osDefaultEditor = sync.OnceValue(func() string {
	return "/etc/alternatives/editor"
})
