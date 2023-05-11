package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"strconv"

	"github.com/denisbrodbeck/machineid"
	"github.com/turbot/flowpipe/fperr"
)

func NodeID(port string) (string, error) {
	// Unique ID for the machine
	mid, err := machineid.ID()
	if err != nil {
		return "", err
	}
	// Combine with the port, so we can have multiple nodes running
	// on the same machine (e.g. in different terminals).
	mpid := fmt.Sprintf("%s:%s", mid, port)
	// Return a short unique ID for the machine/port combination
	return Base36ID(mpid, ApplicationName(), 8)
}

// Base36IDPerMachine returns a base36 hash of the input string, using the
// machine ID as key. This means the same input will return the same hash on
// the same machine. Useful for unique IDs in a distributed system.
func Base36IDPerMachine(input string, length int) (string, error) {
	id, err := machineid.ID()
	if err != nil {
		return "", err
	}
	return Base36ID(id, input, length)
}

// Base36ID returns a base36 hash of the input string, using the provided
// key. Our approach is lossy using only part of the true hash, but good
// enough for our purposes to prevent clashes.
func Base36ID(key string, input string, length int) (string, error) {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(input))
	bs := fmt.Sprintf("%x", mac.Sum(nil))

	// Convert the first 16 chars of the hash from hex to base 36
	u1Hex := bs[0:16]
	u1, err := strconv.ParseUint(u1Hex, 16, 64)
	if err != nil {
		return "", fperr.InternalWithMessage("Unable to create hash.")
	}
	u1Base36 := strconv.FormatUint(u1, 36)

	// Either take the last {length} chars, or pad the result if needed
	if len(u1Base36) > length {
		return u1Base36[len(u1Base36)-length:], nil
	} else {
		return fmt.Sprintf("%0*s", length, u1Base36), nil
	}
}
