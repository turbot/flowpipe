package util

import (
	"fmt"
	"os"

	"github.com/turbot/pipe-fittings/perr"
)

func EnsureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return perr.InternalWithMessage(fmt.Sprintf("error creating directory %s", dir))
		}
	}
	return nil
}
