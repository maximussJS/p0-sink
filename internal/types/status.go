package types

import "p0-sink/internal/enums"

type StatusResponse struct {
	Status enums.EStatus `json:"status"`
}
