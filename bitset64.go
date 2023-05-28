package bitset32

import (
	bitset64 "github.com/bits-and-blooms/bitset"
)

type BitSet64 struct {
	*bitset64.BitSet
}

// TODO: TestFunc
func (b *BitSet64) MaxConsecutiveOne(start, end uint) uint {
	return b.continueMaxCount(start, end, true)
}

func (b *BitSet64) MaxConsecutiveZero(start, end uint) uint {
	return b.continueMaxCount(start, end, false)
}

func (b *BitSet64) continueMaxCount(start, end uint, flag bool) uint {
	flag = !flag
	if end > b.Len() {
		end = b.Len()
	}
	if start >= b.Len() {
		return 0
	}
	if start > end {
		return 0
	}
	rt, sum := uint(0), uint(0)
	for i := start; i < end; i++ {
		if xor(flag, b.Test(i)) {
			sum++
			continue
		}
		if sum > rt {
			rt = sum
		}
		sum = 0
	}
	if sum > rt {
		rt = sum
	}
	return rt
}
