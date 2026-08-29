package service

import (
	"strconv"

	"github.com/cespare/xxhash/v2"
)

func shortHash(data []byte) string {
	return strconv.FormatUint(xxhash.Sum64(data), 36)
}
