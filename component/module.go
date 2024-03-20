package component

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/go-github/v60/github"
	"github.com/izziiyt/compaa/sdk/gopkg"
	"github.com/izziiyt/compaa/sdk/npm"
)

func (m *Module) SyncWithNPM(ctx context.Context, cli *npm.Client) *Module {
	if m.Err != nil {
		return m
	}
	v, err := cli.FetchLatestVersion(ctx, m.Name)
	if err != nil {
		m.Err = err
		return m
	}
	if v.Repository.Type != "git" {
		m.Err = fmt.Errorf("github url not found in npm %v", v)
		return m
	}
	// for "git+https://github.com/gregberge/svgr.git#main" pattern
	if tokens := strings.Split(v.Repository.Url, "#"); len(tokens) > 1 {
		v.Repository.Url = tokens[0]
	}
	// for "git+https://github.com/node ./bin/swc-project/pkgs.git" pattern
	if tokens := strings.Split(v.Repository.Url, " "); len(tokens) > 1 {
		tokens = strings.Split(tokens[1], "/")
		v.Repository.Url = "git+https://github.com/" + strings.Join(tokens[len(tokens)-2:], "/")
	}
	tokens := strings.Split(v.Repository.Url, "/")
	m.GHOrg = tokens[3]
	m.GHRepo = strings.TrimSuffix(tokens[4], ".git")
	return m
}

func (m *Module) SyncWithGopkg(ctx context.Context, cli *gopkg.Client) *Module {
	if m.Err != nil {
		return m
	}
	m.GHOrg, m.GHRepo, m.Err = cli.GetGitHub(ctx, m.Name)
	return m
}

type Module struct {
	Name     string
	Archived bool
	LastPush time.Time
	GHOrg    string
	GHRepo   string
	Err      error
}

func (t *Module) OrgAndRepo() (string, string, error) {
	tokens := strings.Split(t.Name, "/")
	if len(tokens) < 3 {
		return "", "", fmt.Errorf("unexpected module name %v", t.Name)
	}
	return tokens[1], tokens[2], nil
}

func (t *Module) LoadCache() bool {
	v, ok := moduleCache.Load(t.Name)
	if ok {
		_v := v.(*Module)
		t.Name = _v.Name
		t.Archived = _v.Archived
		t.LastPush = _v.LastPush
		t.GHOrg = _v.GHOrg
		t.GHRepo = _v.GHRepo
		t.Err = _v.Err
	}
	return ok
}

func (t *Module) StoreCache() {
	moduleCache.Store(t.Name, t)
}

func (t *Module) SyncWithGitHub(ctx context.Context, cli *github.Client) *Module {
	if t.Err != nil {
		return t
	}
	_r, _, err := cli.Repositories.Get(ctx, t.GHOrg, t.GHRepo)
	if err != nil {
		t.Err = err
		return t
	}
	t.LastPush = _r.PushedAt.Time
	t.Archived = *_r.Archived

	return t
}

func (t *Module) Logging(wc *WarnCondition) {
	if t.Err != nil {
		color.Red("├ ERROR: %v %v\n", t.Name, t.Err)
		return
	}
	if wc.IfArchived && t.Archived {
		color.Yellow("├ WARN: %v is archived\n", t.Name)
		return
	}
	if t.LastPush.AddDate(0, 0, wc.RecentDays).Before(time.Now()) {
		color.Yellow("├ WARN: %v last push isn't recent (%v)\n", t.Name, t.LastPush)
		return
	}
	// _, err := fmt.Fprintf(w, "├ INFO: pass %v last push is recent (%v)\n", t.Name, t.LastPush)
	// return err
}
