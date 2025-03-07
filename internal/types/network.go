package types

type Network struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	ShortName          string `json:"shortName"`
	LogoUrl            string `json:"logoUrl"`
	BlockStreamGrpcUrl string `json:"blockStreamGrpcUrl"`
	Priority           int    `json:"priority"`
	Enabled            bool   `json:"enabled"`
	Hidden             bool   `json:"hidden"`
	OrchestratorUrl    string `json:"orchestratorUrl"`
}
