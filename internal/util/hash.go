package util

import (
	"crypto/sha256"
	"encoding/hex"
)

func CalculateHash(s, salt string) string {
	// Concatenate the input string and the salt
	concatenated := s + salt

	// Create a new SHA-256 hash
	hasher := sha256.New()

	// Write the concatenated string to the hasher
	hasher.Write([]byte(concatenated))

	// Get the final hash value
	hashBytes := hasher.Sum(nil)

	// Convert the hash to a hexadecimal string
	hashString := hex.EncodeToString(hashBytes)

	return hashString
}
