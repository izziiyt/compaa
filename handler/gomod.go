package handler

import (
	"compaa/component"
	"compaa/eol"
	"context"
	"fmt"
	"github.com/google/go-github/v60/github"
	"golang.org/x/mod/modfile"
	"io"
	"os"
	"strings"
)

type GoMod struct {
	gcli *github.Client
	ecli *eol.Client
	wc   *component.WarnCondition
}

func NewGoMod(gcli *github.Client, ecli *eol.Client, wc *component.WarnCondition) *GoMod {
	gm := &GoMod{
		gcli: gcli,
		ecli: ecli,
		wc:   wc,
	}
	if gm.gcli == nil {
		gm.gcli = github.NewClient(nil)
	}
	if gm.ecli == nil {
		gm.ecli = eol.NewClient(nil)
	}
	if gm.wc == nil {
		gm.wc = &component.DefaultWarnCondition
	}
	return gm
}

func (h *GoMod) Handle(ctx context.Context, path string) {
	var buf strings.Builder
	defer func() { fmt.Print(buf.String()) }()

	fmt.Fprintf(&buf, "%v\n", path)

	ts, err := h.LookUp(ctx, path)
	if err != nil {
		fmt.Fprintf(&buf, "LookUp error: %v", err)
	}

	for _, t := range ts {
		t.Logging(&buf, h.wc)
	}
}

func (h *GoMod) LookUp(ctx context.Context, path string) ([]component.Component, error) {
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
	if t, err = t.SyncWithEndOfLife(ctx, h.ecli); err != nil {
		return nil, err
	}
	buf = append(buf, t)

	for _, r := range pf.Require {
		if strings.HasPrefix(r.Mod.Path, "golang.org") {
			continue
		}

		t := &component.Module{
			Name: r.Mod.Path,
		}
		t.GHOrg, t.GHRepo, err = t.OrgAndRepo()
		if err != nil {
			return nil, err
		}

		if t, err = t.SyncWithGitHub(ctx, h.gcli); err != nil {
			return nil, err
		}
		buf = append(buf, t)
	}

	return buf, err
}
