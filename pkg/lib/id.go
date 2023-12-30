package lib

import (
	"github.com/anandvarma/namegen"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid"
)

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

type ShortIDGenerator struct {
}

func (g *ShortIDGenerator) GetId() (string, error) {
	id := shortuuid.New()
	return id, nil
}

type IDNameGenerator struct {
}

func (i *IDNameGenerator) GetId() (string, error) {
	name_schema := namegen.New()
	return name_schema.Get(), nil
}
