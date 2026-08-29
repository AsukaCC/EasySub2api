package admin

import (
	"strings"

	"github.com/google/uuid"
)

func parseEntityID(raw string) (string, error) {
	parsed, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}
