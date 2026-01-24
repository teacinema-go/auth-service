package valueobject

import (
	"fmt"

	"github.com/google/uuid"
)

type ID uuid.UUID

func NewID() (ID, error) {
	uuidV7, err := uuid.NewV7()
	if err != nil {
		return ID(uuidV7), fmt.Errorf("failed to generate UUID: %w", err)
	}

	return ID(uuidV7), nil
}

func NewIDFromString(data string) (ID, error) {
	u, err := uuid.Parse(data)
	if err != nil {
		return ID(u), fmt.Errorf("invalid UUID format: %w", err)
	}

	return ID(u), nil
}

func (id ID) ToUUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id ID) ToString() string {
	return id.ToUUID().String()
}
