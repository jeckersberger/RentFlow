package httphandler

import (
	"github.com/google/uuid"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// parseUUID parses a string into a uuid.UUID, returning an AppError on failure.
func parseUUID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.Wrap(errors.ErrBadRequest, "Ungueltige UUID: "+raw)
	}
	return id, nil
}
