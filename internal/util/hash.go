package util

import (
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/utils"
)

func CalculateHash(s, salt string) (string, error) {
	return utils.Base36Hash(s+salt, 13)
}

func CalculateHashFromGlobalSalt(s string) (string, error) {
	globalSalt, err := GetGlobalSalt()
	if err != nil {
		return "", perr.InternalWithMessage("unable to obtain global salt")
	}

	hashedValue, err := utils.Base36Hash(s+globalSalt, 13)
	if err != nil {
		return "", perr.InternalWithMessage("unable to calculate hash")
	}

	return hashedValue, nil
}
