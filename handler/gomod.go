package handler

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/izziiyt/compaa/component"
	"github.com/izziiyt/compaa/sdk/gopkg"
	"golang.org/x/mod/modfile"
)

type GoMod struct {
	GCli *github.Client
}

func (h *GoMod) LookUp(path string) (buf []component.Component, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return
	}

	pf, err := modfile.Parse(path, b, nil)
	if err != nil {
		return
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

	return
}

func (h *GoMod) SyncWithSource(c component.Component, ctx context.Context) component.Component {
	switch v := c.(type) {
	case *component.Module:
		if strings.HasPrefix(v.Name, "github.com") {
			v.GHOrg, v.GHRepo, v.Err = v.OrgAndRepo()
		} else if strings.HasPrefix(v.Name, "gopkg.in") {
			v = v.SyncWithGopkg(ctx)
		} else {
			v.GHOrg, v.GHRepo, v.Err = gopkg.GetRepoFromCustomDomain(ctx, v.Name)
		}

		v = v.SyncWithGitHub(ctx, h.GCli)
		return v
	case *component.Language:
		v = v.SyncWithEndOfLife(ctx)
		return v
	default:
		return v
	}
}
