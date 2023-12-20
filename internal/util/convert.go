package util

import "math/big"

func BigFloatToInt64(input *big.Float) int64 {
	intValue := new(big.Int)
	input.Int(intValue) // This truncates the decimal part
	result := intValue.Int64()
	return result
}
