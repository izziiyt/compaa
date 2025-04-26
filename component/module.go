package component

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/izziiyt/compaa/sdk/gopkg"
	"github.com/izziiyt/compaa/sdk/npm"
	"github.com/izziiyt/compaa/sdk/pypi"
	"github.com/izziiyt/compaa/sdk/rubygem"
)

var moduleCache = sync.Map{}

type Module struct {
	Name     string
	Archived bool
	LastPush time.Time
	GHOrg    string
	GHRepo   string
	Err      error
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

func (m *Module) SyncWithNPM(ctx context.Context, cli *http.Client) *Module {
	if m.Err != nil {
		return m
	}
	v, err := npm.FetchLatestVersion(ctx, cli, m.Name)
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

func (m *Module) SyncWithGopkg(ctx context.Context, cli *http.Client) *Module {
	if m.Err != nil {
		return m
	}
	m.GHOrg, m.GHRepo, m.Err = gopkg.GetGitHub(ctx, cli, m.Name)
	return m
}

func (t *Module) OrgAndRepo() (string, string, error) {
	tokens := strings.Split(t.Name, "/")
	if len(tokens) < 3 {
		return "", "", fmt.Errorf("unexpected module name %v", t.Name)
	}
	return tokens[1], tokens[2], nil
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

func (t *Module) SyncWithPypi(ctx context.Context, cli *http.Client) *Module {
	if t.Err != nil {
		return t
	}
	r, err := pypi.GetPackage(ctx, cli, t.Name)
	if err != nil {
		t.Err = err
		return t
	}
	tokens := strings.Split(r.RepositoryURL, "/")
	t.GHOrg = tokens[3]
	t.GHRepo = tokens[4]
	return t
}

func (t *Module) SyncWithRubyGem(ctx context.Context, cli *http.Client) *Module {
	if t.Err != nil {
		return t
	}
	r, err := rubygem.GetGem(ctx, cli, t.Name)
	if err != nil {
		t.Err = err
		return t
	}
	var uri string
	if strings.Contains(r.SourceCodeURI, "github.com") {
		uri = r.SourceCodeURI
	} else if strings.Contains(r.DocumentationURI, "github.com") {
		uri = r.DocumentationURI
	} else if strings.Contains(r.HomepageURI, "github.com") {
		uri = r.HomepageURI
	}
	if uri == "" {
		t.Err = fmt.Errorf("github url not found in rubygem %v", r)
		return t
	}
	tokens := strings.Split(uri, "/")
	if len(tokens) < 4 {
		t.Err = fmt.Errorf("unexpected source code uri %v", uri)
		return t
	}
	t.GHOrg = tokens[3]
	t.GHRepo = tokens[4]
	return t
}

func (t *Module) Logging(wc *WarnCondition, logger Logger) {
	if logger == nil {
		logger = &DefaultLogger{}
	}

	if t.Err != nil {
		logger.Error("├ ERROR: %v %v\n", t.Name, t.Err)
		return
	}
	if wc.IfArchived && t.Archived {
		logger.Warn("├ WARN: %v is archived\n", t.Name)
		return
	}
	if t.LastPush.AddDate(0, 0, wc.RecentDays).Before(time.Now()) {
		logger.Warn("├ WARN: %v last push isn't recent (%v)\n", t.Name, t.LastPush)
		return
	}
	// logger.Info("├ INFO: pass %v last push is recent (%v)\n", t.Name, t.LastPush)
}
