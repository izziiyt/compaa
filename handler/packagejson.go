package handler

import (
	"context"
	"encoding/json"
	"io"
	"os"

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

func (h *PackageJSON) SyncWithSource(ctx context.Context, c component.Component) component.Component {
	switch v := c.(type) {
	case *component.Module:
		v = v.SyncWithNPM(ctx)
		v = v.SyncWithGitHub(ctx, h.gcli)
		return v
	default:
		return v
	}
}
