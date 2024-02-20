package util

import "math/big"

const RFC3389WithMS = "2006-01-02T15:04:05.000Z07:00"

func BigFloatToInt64(input *big.Float) int64 {
	intValue := new(big.Int)
	input.Int(intValue) // This truncates the decimal part
	result := intValue.Int64()
	return result
}

func SafeDeref[T any](ptr *T) T {
	if ptr != nil {
		return *ptr
	}
	var zeroVal T
	return zeroVal
}
