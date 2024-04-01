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
	GCli *github.Client
}

func (h *PackageJSON) LookUp(path string) (buf []component.Component, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return
	}

	ps, err := parsePackageJSON(b)
	for _, p := range ps {
		t := &component.Module{
			Name: p.Name,
		}
		buf = append(buf, t)
	}

	return
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

func (h *PackageJSON) SyncWithSource(c component.Component, ctx context.Context) component.Component {
	switch v := c.(type) {
	case *component.Module:
		v = v.SyncWithNPM(ctx)
		v = v.SyncWithGitHub(ctx, h.GCli)
		return v
	default:
		return v
	}
}
