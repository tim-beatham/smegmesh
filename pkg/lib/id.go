package lib

import "github.com/google/uuid"

// IdGenerator generates unique ids
type IdGenerator interface {
	// GetId generates a unique ID or an error if something went wrong
	GetId() (string, error)
}

type UUIDGenerator struct {
}

func (g *UUIDGenerator) GetId() (string, error) {
	id := uuid.New()
	return id.String(), nil
}
