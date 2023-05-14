//go:build go1.9
// +build go1.9

package bitset32

import "math/bits"

func trailingZeroes32(v uint32) uint {
	return uint(bits.TrailingZeros32(v))
}
