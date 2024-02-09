package util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/filepaths"
	"log/slog"
	"os"
	"path/filepath"
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

func GetModSaltOrDefault() (string, error) {
	c := cache.GetCache()
	if ms, exists := c.Get("mod_salt"); exists {
		if modSalt, ok := ms.(string); ok {
			return modSalt, nil
		} else {
			return modSalt, fmt.Errorf("mod specific salt not a string")
		}
	}

	return GetGlobalSalt()
}

func GetGlobalSalt() (string, error) {
	c := cache.GetCache()
	if s, exists := c.Get("salt"); exists {
		if salt, ok := s.(string); ok {
			return salt, nil
		} else {
			return salt, fmt.Errorf("salt not a string")
		}
	}
	globalSaltPath := filepath.Join(filepaths.GlobalInternalDir(), "salt")
	return CreateFlowpipeSalt(globalSaltPath, 32)
}
