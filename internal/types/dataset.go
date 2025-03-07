package types

type Dataset struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	Priority  *int   `json:"priority"`
	Enabled   bool   `json:"enabled"`
	Hidden    bool   `json:"hidden"`
	NetworkId string `json:"networkId"`
	IsPaid    bool   `json:"isPaid"`
}
