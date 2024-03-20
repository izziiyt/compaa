package component

import (
	"sync"
)

var (
	moduleCache   sync.Map
	languageCache sync.Map
)

type Component interface {
	Logging(wc *WarnCondition) error
	LoadCache() bool
	StoreCache()
}

func init() {
	moduleCache = sync.Map{}
	languageCache = sync.Map{}
}
