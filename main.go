package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/izziiyt/compaa/component"
	"github.com/izziiyt/compaa/handler"
)

var (
	rd    = flag.Int("d", 180, "recent days. used to determine log level")
	token = flag.String("t", "", "github token. recommend to set for sufficient github api rate limit")
)

func main() {
	flag.Parse()
	args := flag.Args()
	path := "."
	if len(args) > 0 {
		path = args[0]
	}
	ctx := context.Background()
	wc := &component.DefaultWarnCondition
	wc.RecentDays = *rd
	r := NewRouter(*token, wc)
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && strings.Contains(d.Name(), "node_modules") {
			return filepath.SkipDir
		}
		h := r.Route(path)
		if h == nil {
			return nil
		}
		h.Handle(ctx, path)
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
}

type Router struct {
	gomod       *handler.GoMod
	packagejson *handler.PackageJSON
}

func NewRouter(ghtoken string, wc *component.WarnCondition) *Router {
	gcli := github.NewClient(nil)
	if ghtoken != "" {
		gcli = gcli.WithAuthToken(ghtoken)
	}
	return &Router{
		gomod:       handler.NewGoMod(gcli, wc),
		packagejson: handler.NewPackageJSON(gcli, wc),
	}
}

func (r *Router) Route(path string) handler.Handler {
	if strings.Contains(path, "go.mod") {
		return r.gomod
	}
	if strings.Contains(path, "package.json") {
		return r.packagejson
	}
	return nil
}
