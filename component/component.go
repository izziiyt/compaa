package component

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/izziiyt/compaa/sdk/eol"
	"github.com/izziiyt/compaa/sdk/gopkg"
	"github.com/izziiyt/compaa/sdk/npm"
)

var (
	moduleCache   sync.Map
	languageCache sync.Map
)

type Component interface {
	Logging(w io.Writer, wc *WarnCondition) error
}

func init() {
	moduleCache = sync.Map{}
	languageCache = sync.Map{}
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

func (t *Module) UseCache() (*Module, bool) {
	v, ok := moduleCache.Load(t.Name)
	if ok {
		return v.(*Module), true
	}
	return t, false
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

	moduleCache.Store(t.Name, t)

	return t
}

func (t *Module) Logging(w io.Writer, wc *WarnCondition) error {
	if t.Err != nil {
		_, err := fmt.Fprintf(w, "├ ERROR: %v %v\n", t.Name, t.Err)
		return err
	}
	if wc.IfArchived && t.Archived {
		_, err := fmt.Fprintf(w, "├ WARN: %v is archived\n", t.Name)
		return err
	}
	if t.LastPush.AddDate(0, 0, wc.RecentDays).Before(time.Now()) {
		_, err := fmt.Fprintf(w, "├ WARN: %v last push isn't recent (%v)\n", t.Name, t.LastPush)
		return err
	}
	// _, err := fmt.Fprintf(w, "├ INFO: pass %v last push is recent (%v)\n", t.Name, t.LastPush)
	// return err
	return nil
}

type Language struct {
	Name    string
	Version string
	EOL     bool
	EOLDate time.Time
}

func (t *Language) SyncWithEndOfLife(ctx context.Context, cli *eol.Client) (*Language, error) {
	if _t, ok := languageCache.Load(t.Name + t.Version); ok {
		return _t.(*Language), nil
	}
	splited := strings.Split(t.Version, ".")
	cd, err := cli.SingleCycleDetail(ctx, t.Name, strings.Join(splited[0:2], "."))
	if err != nil {
		return t, err
	}

	t.EOL = cd.EOL
	t.EOLDate = cd.EOLDate

	languageCache.Store(t.Name+t.Version, t)

	return t, err
}

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
	tokens := strings.Split(v.Repository.Url, "/")
	m.GHOrg = tokens[3]
	m.GHRepo = strings.Split(tokens[4], ".")[0]
	return m
}

func (m *Module) SyncWithGopkg(ctx context.Context, cli *gopkg.Client) *Module {
	if m.Err != nil {
		return m
	}
	var err error
	m.GHOrg, m.GHRepo, err = cli.GetGitHub(ctx, m.Name)
	if err != nil {
		m.Err = err
	}
	return m
}

func (t *Language) Logging(w io.Writer, wc *WarnCondition) error {
	if wc.IfArchived && t.EOL {
		_, err := fmt.Fprintf(w, "├ WARN: %v%v is EOL\n", t.Name, t.Version)
		return err
	}
	if !t.EOLDate.IsZero() && time.Now().AddDate(0, 0, wc.RecentDays).After(t.EOLDate) {
		_, err := fmt.Fprintf(w, "├ WARN: %v%v EOL is recent\n", t.Name, t.Version)
		return err
	}
	// _, err := fmt.Fprintf(w, "├ INFO: pass %v%v\n", t.Name, t.Version)
	// return err
	return nil
}
