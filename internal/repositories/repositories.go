package repositories

type txKeyType string

const (
	TxKey txKeyType = "repositories.tx_key"
)

type LockType string

const (
	LockTypeNoLock             LockType = "NO_LOCK"
	LockTypeLockForUpdate      LockType = "LOCK_FOR_UPDATE"
	LockTypeLockForNoKeyUpdate LockType = "LOCK_FOR_NO_KEY_UPDATE"
)

type CascadeBehavior string

const (
	CascadeBehaviorNoWait     CascadeBehavior = "NO_WAIT"
	CascadeBehaviorSkipLocked CascadeBehavior = "SKIP_LOCKED"
)

type LockControlParameters struct {
	Type     LockType
	Behavior CascadeBehavior
}

type QueryParameters struct {
	SortBy []SortField
	Limit  int
	Offset int
}

type SortField struct {
	Name  string
	Order SortOrder
}

type SortOrder string

const (
	SortOrderAscending  SortOrder = "ASC"
	SortOrderDescending SortOrder = "DESC"
)
