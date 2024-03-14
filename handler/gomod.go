package handler

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/google/go-github/v60/github"
	"github.com/izziiyt/compaa/component"
	"github.com/izziiyt/compaa/sdk/eol"
	"github.com/izziiyt/compaa/sdk/gopkg"
	"golang.org/x/mod/modfile"
)

type GoMod struct {
	gcli  *github.Client
	ecli  *eol.Client
	gpcli *gopkg.Client
	wc    *component.WarnCondition
}

func NewGoMod(gcli *github.Client, wc *component.WarnCondition) *GoMod {
	gm := &GoMod{
		gcli:  gcli,
		ecli:  eol.NewClient(nil),
		gpcli: gopkg.NewClient(nil),
		wc:    wc,
	}
	if gm.wc == nil {
		gm.wc = &component.DefaultWarnCondition
	}
	return gm
}

func (h *GoMod) Handle(ctx context.Context, path string) {
	fmt.Printf("%v\n", path)

	ts, err := h.LookUp(path)
	if err != nil {
		fmt.Printf("├ LookUp error: %v\n", err)
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
						v = v.SyncWithGopkg(ctx, h.gpcli)
					} else {
						v.GHOrg, v.GHRepo, v.Err = gopkg.GetRepoFromCustomDomain(ctx, v.Name)
					}

					v = v.SyncWithGitHub(ctx, h.gcli)
					v.StoreCache()
				case *component.Language:
					v = v.SyncWithEndOfLife(ctx, h.ecli)
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
	defer f.Close()
	if err != nil {
		return nil, err
	}

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
