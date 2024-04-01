package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/izziiyt/compaa/component"
	"github.com/izziiyt/compaa/handler"
)

var (
	rd    = flag.Int("d", 360, "recent days. used to determine log level")
	token = flag.String("t", "", "github token. recommend to set for sufficient github api rate limit")
)

func main() {
	flag.Parse()
	args := flag.Args()
	path := "."
	if len(args) > 0 {
		path = args[0]
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Println(path + " not found")
			return
		}
	}
	ctx := context.Background()
	wc := &component.DefaultWarnCondition
	wc.RecentDays = *rd
	r := NewRouter(*token)
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && excludedPatterns(d.Name()) {
			return filepath.SkipDir
		}
		if h := r.Route(d.Name()); h != nil {
			handler.Handle(ctx, h, path, wc)
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
}

func excludedPatterns(path string) bool {
	excludePatterns := []string{
		"node_modules",
		"vendor",
		".git",
		".vscode",
		".idea",
	}
	for _, p := range excludePatterns {
		if strings.Contains(path, p) {
			return true
		}
	}
	return false
}
