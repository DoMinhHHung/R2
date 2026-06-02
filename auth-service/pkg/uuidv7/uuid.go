package uuidv7

import "github.com/google/uuid"

func New() string {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New().String()
	}
	return id.String()
}
