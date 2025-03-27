package types

import "time"

type BlockFunction struct {
	StreamId  string    `json:"streamId"`
	Code      *string   `json:"code,omitempty"`
	Enabled   bool      `json:"enabled"`
	UpdatedAt time.Time `json:"updatedAt"`
}
