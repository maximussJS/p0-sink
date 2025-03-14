package types

type Destination struct {
	StreamId string                 `json:"streamId"`
	Type     string                 `json:"type"`
	Config   map[string]interface{} `json:"config"`
}
