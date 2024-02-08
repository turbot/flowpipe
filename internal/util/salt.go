package util

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"os"
)

// Assumes that the dir exists
//
// The function creates the salt if it does not exist, or it returns the existing
// salt if it's already there
func CreateFlowpipeSalt(filename string, length int) (string, error) {
	// Check if the salt file exists
	if _, err := os.Stat(filename); err == nil {
		// If the file exists, read the salt from it
		slog.Debug("Salt file exists, reading from it", "filename", filename, "length", length)
		saltBytes, err := os.ReadFile(filename)
		if err != nil {
			return "", err
		}
		return string(saltBytes), nil
	}

	slog.Debug("Salt file does not exist, creating a new one", "filename", filename, "length", length)
	// If the file does not exist, generate a new salt
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	// Encode the salt as a hexadecimal string
	saltHex := hex.EncodeToString(salt)

	// Write the salt to the file
	err = os.WriteFile(filename, []byte(saltHex), 0600)
	if err != nil {
		return "", err
	}

	return saltHex, nil
}
