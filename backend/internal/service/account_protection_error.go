package service

import (
	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"strings"
)

func IsAccountProtectionError(err error) bool {
	reason := infraerrors.Reason(err)
	return strings.HasPrefix(reason, "MODE1_") || strings.HasPrefix(reason, "PROTECTION_") || strings.HasPrefix(reason, "REQUEST_INTEGRITY_")
}
