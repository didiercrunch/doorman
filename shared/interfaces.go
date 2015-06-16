package shared

import "math/big"

type DoormanUpdater struct {
	Id            string     `json:"id"`
	Timestamp     int64      `json:"timestamp"`
	Probabilities []*big.Rat `json:"probabilities"`
}

type UpdateHandlerFunc func(m *DoormanUpdater) error
