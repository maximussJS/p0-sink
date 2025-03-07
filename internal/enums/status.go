package enums

type EStatus string

const (
	StatusStarting   EStatus = "starting"
	StatusRunning    EStatus = "running"
	StatusPausing    EStatus = "pausing"
	StatusPaused     EStatus = "paused"
	StatusTerminated EStatus = "terminated"
	StatusFinished   EStatus = "finished"
)
