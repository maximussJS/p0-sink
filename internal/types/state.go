package types

type State struct {
	StreamId          string  `json:"streamId"`
	Status            string  `json:"status"`
	TerminationReason *string `json:"terminationReason"`
	LastCursor        string  `json:"lastCursor"`
	LastBlockNumber   *uint64 `json:"lastBlockNumber"`
	BlocksSent        int64   `json:"blocksSent"`
	BytesSent         int64   `json:"bytesSent"`
	TimeSpent         int64   `json:"timeSpent"`
	LastDirection     int     `json:"lastDirection"`
	StatusUpdatedAt   int64   `json:"statusUpdatedAt"`
	CursorUpdatedAt   int64   `json:"cursorUpdatedAt"`
}
