package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/izziiyt/compaa/component"
	"github.com/izziiyt/compaa/handler"
)

var (
	rd    = flag.Int("d", 360, "recent days. used to determine log level")
	token = flag.String("t", "", "github token. recommended to set for sufficient github api rate limit, or set GITHUB_TOKEN env var")
)

func main() {
	flag.Parse()
	args := flag.Args()
	path := "."
	if len(args) > 0 {
		path = args[0]
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, path+" not found")
			os.Exit(1)
		}
	}
	ctx := context.Background()
	wc := &component.DefaultWarnCondition
	wc.RecentDays = *rd
	if *token == "" {
		*token = os.Getenv("GITHUB_TOKEN")
	}
	if *token == "" {
		fmt.Println("WARN: recommended to use github token. see `compaa -h`")
	}
	transport := NewCacheTransport(http.DefaultTransport)
	defer transport.Close()
	r := NewRouter(*token, transport)
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && excludedPatterns(d.Name()) {
			return filepath.SkipDir
		}
		if h := r.Route(d.Name()); h != nil {
			handler.Handle(h, ctx, path, wc)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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
