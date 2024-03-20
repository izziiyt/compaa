package component

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/izziiyt/compaa/sdk/eol"
)

type Language struct {
	Name    string
	Version string
	EOL     bool
	EOLDate time.Time
	Err     error
}

func (t *Language) SyncWithEndOfLife(ctx context.Context, cli *eol.Client) *Language {
	splited := strings.Split(t.Version, ".")
	cd, err := cli.SingleCycleDetail(ctx, t.Name, strings.Join(splited[0:2], "."))
	if err != nil {
		t.Err = err
		return t
	}

	t.EOL = cd.EOL
	t.EOLDate = cd.EOLDate

	return t
}

func (t *Language) Logging(wc *WarnCondition) error {
	if t.Err != nil {
		_, err := fmt.Printf("├ ERROR: %v %v\n", t.Name, t.Err)
		return err
	}
	if wc.IfArchived && t.EOL {
		_, err := fmt.Printf("├ WARN: %v%v is EOL\n", t.Name, t.Version)
		return err
	}
	if !t.EOLDate.IsZero() && time.Now().AddDate(0, 0, wc.RecentDays).After(t.EOLDate) {
		_, err := fmt.Printf("├ WARN: %v%v EOL is recent\n", t.Name, t.Version)
		return err
	}
	// _, err := fmt.Fprintf(w, "├ INFO: pass %v%v\n", t.Name, t.Version)
	// return err
	return nil
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
