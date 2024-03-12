package component

type WarnCondition struct {
	IfArchived bool
	RecentDays int
}

var DefaultWarnCondition = WarnCondition{
	IfArchived: true,
	RecentDays: 180,
}
