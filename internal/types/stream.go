package types

import (
	"encoding/json"
	"fmt"
	"time"
)

type Stream struct {
	Id                string                 `json:"id"`
	Name              string                 `json:"name"`
	TeamId            string                 `json:"teamId"`
	DestinationEntity Destination            `json:"destinationEntity"`
	DestinationType   string                 `json:"destinationType"`
	State             State                  `json:"state"`
	DatasetId         string                 `json:"datasetId"`
	Dataset           Dataset                `json:"dataset"`
	NetworkId         string                 `json:"networkId"`
	Network           Network                `json:"network"`
	FromBlock         int64                  `json:"fromBlock"`
	ToBlock           int64                  `json:"toBlock"`
	LagFromRealtime   int32                  `json:"lagFromRealtime"`
	BlockFunction     *BlockFunction         `json:"blockFunction,omitempty"`
	MaxBatchSize      int                    `json:"maxBatchSize"`
	Flags             map[string]interface{} `json:"flags"`
	CreatedAt         time.Time              `json:"createdAt"`
	UpdatedAt         time.Time              `json:"updatedAt"`
	DeletedAt         *time.Time             `json:"deletedAt"` // pointer since it can be null
}

func (s *Stream) String() string {
	out, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Sprintf("Stream{ error: %v }", err)
	}
	return string(out)
}
