package bitset32

import (
	bitset64 "github.com/bits-and-blooms/bitset"
)

type BitSet64 struct {
	*bitset64.BitSet
}

// FIXME: too slow
func (b *BitSet64) MaxConsecutiveOne() uint {
	rt := uint(0)
	sum := uint(0)
	for i := uint(0); i <= b.Len(); i++ {
		if b.Test(i) {
			sum++
			continue
		}
		if rt < sum {
			rt = sum
		}
		sum = 0
	}

	return uint(rt)
}
