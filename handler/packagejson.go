package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/google/go-github/v60/github"
	"github.com/izziiyt/compaa/component"
	"github.com/izziiyt/compaa/sdk/eol"
	"github.com/izziiyt/compaa/sdk/npm"
)

type PackageJSON struct {
	gcli *github.Client
	ecli *eol.Client
	ncli *npm.Client
	wc   *component.WarnCondition
}

func NewPackageJSON(gcli *github.Client, wc *component.WarnCondition) *PackageJSON {
	gm := &PackageJSON{
		gcli: gcli,
		ecli: eol.NewClient(nil),
		ncli: npm.NewClient(nil),
		wc:   wc,
	}
	if gm.wc == nil {
		gm.wc = &component.DefaultWarnCondition
	}
	return gm
}

func (h *PackageJSON) Handle(ctx context.Context, path string) {
	fmt.Printf("%v\n", path)

	ts, err := h.LookUp(path)
	if err != nil {
		fmt.Printf("LookUp error: %v", err)
		return
	}

	wg := &sync.WaitGroup{}
	done := make(chan struct{}, 10)
	for _, t := range ts {
		wg.Add(1)
		go func(ctx context.Context, t component.Component) {
			done <- struct{}{}
			if ok := t.LoadCache(); !ok {
				switch v := t.(type) {
				case *component.Module:
					v = v.SyncWithNPM(ctx, h.ncli)
					v = v.SyncWithGitHub(ctx, h.gcli)
					v.StoreCache()
				}
			}
			t.StoreCache()
			t.Logging(h.wc)
			<-done
			wg.Done()
		}(ctx, t)
	}
	wg.Wait()
}

func (h *PackageJSON) LookUp(path string) ([]component.Component, error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	ps, err := parsePackageJSON(b)

	var buf []component.Component
	for _, p := range ps {
		t := &component.Module{
			Name: p.Name,
		}
		buf = append(buf, t)
	}
	return buf, nil
}

type pjJSON struct {
	DEV  bool
	Name string
}

func parsePackageJSON(b []byte) (ps []*pjJSON, err error) {
	j := struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}{}
	if err = json.Unmarshal(b, &j); err != nil {
		return
	}
	for k := range j.Dependencies {
		ps = append(ps, &pjJSON{DEV: false, Name: k})
	}
	for k := range j.DevDependencies {
		ps = append(ps, &pjJSON{DEV: true, Name: k})
	}
	return
}
