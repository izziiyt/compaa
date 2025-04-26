package component

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/izziiyt/compaa/sdk/eol"
)

var languageCache = sync.Map{}

type Language struct {
	Name    string
	Version string
	EOL     bool
	EOLDate time.Time
	Err     error
}

func (t *Language) SyncWithEndOfLife(ctx context.Context, cli *http.Client) *Language {
	splited := strings.Split(t.Version, ".")
	cd, err := eol.SingleCycleDetail(ctx, cli, t.Name, strings.Join(splited[0:2], "."))
	if err != nil {
		t.Err = err
		return t
	}

	t.EOL = cd.EOL
	t.EOLDate = cd.EOLDate

	return t
}

func (t *Language) Logging(wc *WarnCondition, logger Logger) {
	if logger == nil {
		logger = &DefaultLogger{}
	}

	if t.Err != nil {
		logger.Error("├ ERROR: %v %v\n", t.Name, t.Err)
		return
	}
	if wc.IfArchived && t.EOL {
		logger.Warn("├ WARN: %v%v is EOL\n", t.Name, t.Version)
		return
	}
	if !t.EOLDate.IsZero() && time.Now().AddDate(0, 0, wc.RecentDays).After(t.EOLDate) {
		logger.Warn("├ WARN: %v%v EOL is recent\n", t.Name, t.Version)
		return
	}
	// logger.Info("├ INFO: pass %v%v\n", t.Name, t.Version)
}

func (t *Language) LoadCache() bool {
	v, ok := languageCache.Load(t.Name)
	if ok {
		_v := v.(*Language)
		t.Name = _v.Name
		t.Version = _v.Version
		t.EOL = _v.EOL
		t.EOLDate = _v.EOLDate
		t.Err = _v.Err
	}
	return ok
}

func (t *Language) StoreCache() {
	languageCache.Store(t.Name+t.Version, t)
}
