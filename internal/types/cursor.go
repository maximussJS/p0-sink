package types

import (
	"p0-sink/internal/utils/direction"
	pb "p0-sink/proto"
	"time"
)

type UpdateCursorPayload struct {
	LastCursor      string       `json:"lastCursor"`
	LastBlockNumber uint64       `json:"lastBlockNumber"`
	BlocksSent      int64        `json:"blocksSent"`
	BytesSent       int64        `json:"bytesSent"`
	TimeSpent       int64        `json:"timeSpent"`
	CursorUpdatedAt int64        `json:"cursorUpdatedAt"`
	LastDirection   pb.Direction `json:"lastDirection"`
}

func NewUpdateCursorPayload(batch ProcessedBatch, state State, timeSpentDelta int64) *UpdateCursorPayload {
	return &UpdateCursorPayload{
		LastCursor:      batch.Cursor,
		LastBlockNumber: batch.LastBlockNumber(),
		BlocksSent:      state.BlocksSent + int64(batch.NumBlocks()),
		BytesSent:       state.BytesSent + int64(batch.BilledBytes),
		TimeSpent:       state.TimeSpent + timeSpentDelta,
		CursorUpdatedAt: time.Now().Unix(),
		LastDirection:   direction.ArrowToDirection(batch.Direction),
	}
}
