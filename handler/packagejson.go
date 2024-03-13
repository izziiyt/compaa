package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

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

func NewPackageJSON(gcli *github.Client, ecli *eol.Client, ncli *npm.Client, wc *component.WarnCondition) *PackageJSON {
	gm := &PackageJSON{
		gcli: gcli,
		ecli: ecli,
		ncli: ncli,
		wc:   wc,
	}
	if gm.gcli == nil {
		gm.gcli = github.NewClient(nil)
	}
	if gm.ecli == nil {
		gm.ecli = eol.NewClient(nil)
	}
	if gm.ncli == nil {
		gm.ncli = npm.NewClient(nil)
	}
	if gm.wc == nil {
		gm.wc = &component.DefaultWarnCondition
	}
	return gm
}

func (h *PackageJSON) Handle(ctx context.Context, path string) {
	var buf strings.Builder
	defer func() { fmt.Print(buf.String()) }()

	fmt.Fprintf(&buf, "%v\n", path)

	ts, err := h.LookUp(ctx, path)
	if err != nil {
		fmt.Fprintf(&buf, "LookUp error: %v", err)
		return
	}

	for _, t := range ts {
		t.Logging(&buf, h.wc)
	}
}

func (h *PackageJSON) LookUp(ctx context.Context, path string) ([]component.Component, error) {
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

	ps, err := parsePackageJSON(b)

	for _, p := range ps {
		t := &component.Module{
			Name: p.Name,
		}
		t = t.SyncWithNPM(ctx, h.ncli)
		t = t.SyncWithGitHub(ctx, h.gcli)
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
	// for k := range j.DevDependencies {
	// 	ps = append(ps, &pjJSON{DEV: true, Name: k})
	// }
	return
}
