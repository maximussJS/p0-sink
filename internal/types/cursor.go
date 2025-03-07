package types

type Direction int32

const (
	DirectionForward      Direction = 0
	DirectionBackward     Direction = 1
	DirectionUnrecognized Direction = -1
)

type UpdateCursorPayload struct {
	LastCursor      string    `json:"lastCursor"`
	LastBlockNumber int64     `json:"lastBlockNumber"`
	BlocksSent      int64     `json:"blocksSent"`
	BytesSent       int64     `json:"bytesSent"`
	TimeSpent       int64     `json:"timeSpent"`
	CursorUpdatedAt int64     `json:"cursorUpdatedAt"`
	LastDirection   Direction `json:"lastDirection"`
}
