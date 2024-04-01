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
	gcli *github.Client
}

func NewGoMod(gcli *github.Client) *GoMod {
	return &GoMod{
		gcli,
	}
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

		v = v.SyncWithGitHub(ctx, h.gcli)
		return v
	case *component.Language:
		v = v.SyncWithEndOfLife(ctx)
		return v
	default:
		return v
	}
}
