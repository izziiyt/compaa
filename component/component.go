package component

import (
	"compaa/eol"
	"compaa/npm"
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/v60/github"
	"io"
	"strings"
	"sync"
	"time"
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
}

func (t *Module) OrgAndRepo() (string, string, error) {
	tokens := strings.Split(t.Name, "/")
	if len(tokens) < 2 {
		return "", "", errors.New("unexpected module format")
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

func (t *Module) SyncWithGitHub(ctx context.Context, cli *github.Client) (*Module, error) {
	_r, _, err := cli.Repositories.Get(ctx, t.GHOrg, t.GHRepo)
	if err != nil {
		return t, err
	}

	t.LastPush = _r.PushedAt.Time
	t.Archived = *_r.Archived

	moduleCache.Store(t.Name, t)

	return t, nil
}

func (t *Module) Logging(w io.Writer, wc *WarnCondition) error {
	if wc.IfArchived && t.Archived {
		_, err := fmt.Fprintf(w, "├ WARN: %v is archived\n", t.Name)
		return err
	}
	if t.LastPush.AddDate(0, 0, wc.RecentDays).Before(time.Now()) {
		_, err := fmt.Fprintf(w, "├ WARN: %v last push isn't recent (%v)\n", t.Name, t.LastPush)
		return err
	}
	_, err := fmt.Fprintf(w, "├ INFO: pass %v last push is recent (%v)\n", t.Name, t.LastPush)
	return err
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

func (m *Module) SyncWithNPM(ctx context.Context, cli *npm.Client) (*Module, error) {
	v, err := cli.FetchLatestVersion(ctx, m.Name)
	if err != nil {
		return m, err
	}
	if v.Repository.Type != "git" {
		return m, fmt.Errorf("not git %v", v)
	}
	tokens := strings.Split(v.Repository.Url, "/")
	m.GHOrg = tokens[3]
	m.GHRepo = strings.Split(tokens[4], ".")[0]
	return m, nil
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
	_, err := fmt.Fprintf(w, "├ INFO: pass %v%v\n", t.Name, t.Version)
	return err
}
