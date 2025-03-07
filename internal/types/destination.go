package types

type Destination struct {
	StreamID string                 `json:"streamId"`
	Type     string                 `json:"type"`
	Config   map[string]interface{} `json:"config"`
}
