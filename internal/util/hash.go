package util

import (
	"github.com/turbot/pipe-fittings/utils"
)

func CalculateHash(s, salt string) (string, error) {
	return utils.Base36Hash(s+salt, 16)
}
