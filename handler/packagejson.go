package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/google/go-github/v60/github"
	"github.com/izziiyt/compaa/component"
)

type PackageJSON struct {
	wc   *component.WarnCondition
	gcli *github.Client
}

func NewPackageJSON(wc *component.WarnCondition, gcli *github.Client) *PackageJSON {
	return &PackageJSON{
		wc,
		gcli,
	}
}

func (h *PackageJSON) Handle(ctx context.Context, path string) {
	fmt.Printf("%v\n", path)

	ts, err := h.LookUp(path)
	if err != nil {
		color.Red("LookUp error: %v", err)
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
					v = v.SyncWithNPM(ctx)
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
	if err != nil {
		return nil, err
	}
	defer f.Close()

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
