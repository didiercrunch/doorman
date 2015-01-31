package shared

type DoormanUpdater struct {
	Id            string    `json:"id"`
	Timestamp     int64     `json:"timestamp"`
	Probabilities []float64 `json:"probabilities"`
}

type UpdateHandlerFunc func(m *DoormanUpdater) error
