package main

import (
	"fmt"
	"github.com/gobwas/glob"
)

func main() {
	pattern := "*.go"
	testFiles := []string{"main.go", "src/helper.go", "helper.go", "test.js"}
	
	g := glob.MustCompile(pattern, '/')
	
	for _, file := range testFiles {
		matches := g.Match(file)
		fmt.Printf("Pattern '%s' matches '%s': %v\n", pattern, file, matches)
	}
}