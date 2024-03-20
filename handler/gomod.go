package handler

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/google/go-github/v60/github"
	"github.com/izziiyt/compaa/component"
	"github.com/izziiyt/compaa/sdk/gopkg"
	"golang.org/x/mod/modfile"
)

type GoMod struct {
	wc   *component.WarnCondition
	gcli *github.Client
}

func NewGoMod(wc *component.WarnCondition, gcli *github.Client) *GoMod {
	return &GoMod{
		wc,
		gcli,
	}
}

func (h *GoMod) Handle(ctx context.Context, path string) {
	fmt.Printf("%v\n", path)

	ts, err := h.LookUp(path)
	if err != nil {
		color.Red("├ LookUp error: %v\n", err)
	}

	wg := &sync.WaitGroup{}
	done := make(chan struct{}, 10)
	for _, t := range ts {
		if ok := t.LoadCache(); ok {
			t.Logging(h.wc)
			continue
		}
		wg.Add(1)
		go func(ctx context.Context, t component.Component) {
			done <- struct{}{}
			if ok := t.LoadCache(); !ok {
				switch v := t.(type) {
				case *component.Module:
					if strings.HasPrefix(v.Name, "github.com") {
						v.GHOrg, v.GHRepo, v.Err = v.OrgAndRepo()
					} else if strings.HasPrefix(v.Name, "gopkg.in") {
						v = v.SyncWithGopkg(ctx)
					} else {
						v.GHOrg, v.GHRepo, v.Err = gopkg.GetRepoFromCustomDomain(ctx, v.Name)
					}

					v = v.SyncWithGitHub(ctx, h.gcli)
					v.StoreCache()
				case *component.Language:
					v = v.SyncWithEndOfLife(ctx)
					v.StoreCache()
				}
			}
			t.Logging(h.wc)
			<-done
			wg.Done()
		}(ctx, t)
	}
	wg.Wait()
}

func (h *GoMod) LookUp(path string) ([]component.Component, error) {
	var buf []component.Component
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	pf, err := modfile.Parse(path, b, nil)
	if err != nil {
		return nil, err
	}

	t := &component.Language{
		Name:    "go",
		Version: pf.Go.Version,
	}
	buf = append(buf, t)

	for _, r := range pf.Require {
		if r.Indirect {
			continue
		}

		t := &component.Module{
			Name: r.Mod.Path,
		}

		buf = append(buf, t)
	}

	return buf, err
}
