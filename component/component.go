package component

type Component interface {
	Logging(wc *WarnCondition)
	LoadCache() bool
	StoreCache()
}
