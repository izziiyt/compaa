package component

type Component interface {
	Logging(wc *WarnCondition, logger Logger)
	LoadCache() bool
	StoreCache()
}
