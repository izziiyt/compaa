package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

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
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	ctx := context.Background()
	wc := &component.DefaultWarnCondition
	wc.RecentDays = *rd
	r := NewRouter(*token, wc)
	wg := &sync.WaitGroup{}
	done := make(chan struct{}, 10)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		h := r.Route(e.Name())
		if h == nil {
			continue
		}
		wg.Add(1)
		go func(ctx context.Context, path string) {
			done <- struct{}{}
			h.Handle(ctx, path)
			<-done
			wg.Done()
		}(ctx, path+"/"+e.Name())
	}
	wg.Wait()
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
		gomod:       handler.NewGoMod(gcli, nil, wc),
		packagejson: handler.NewPackageJSON(gcli, nil, nil, wc),
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
