package handler

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

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
	if gm.gpcli == nil {
		gm.gpcli = gopkg.NewClient(nil)
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
		fmt.Fprintf(&buf, "├ LookUp error: %v\n", err)
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
		if r.Indirect {
			continue
		}

		t := &component.Module{
			Name: r.Mod.Path,
		}

		if strings.HasPrefix(r.Mod.Path, "github.com") {
			t.GHOrg, t.GHRepo, err = t.OrgAndRepo()
			if err != nil {
				return nil, err
			}
		} else if strings.HasPrefix(r.Mod.Path, "go.uber") {
			t.GHOrg = "uber-go"
			t.GHRepo = strings.Split(t.Name, "/")[1]
		} else if strings.HasPrefix(r.Mod.Path, "gopkg.in") {
			t = t.SyncWithGopkg(ctx, h.gpcli)
		} else {
			continue
		}

		t = t.SyncWithGitHub(ctx, h.gcli)

		buf = append(buf, t)
	}

	return buf, err
}
