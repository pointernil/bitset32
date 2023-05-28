package bitset32

// MaxConsecutiveOne
func (b *BitSet32) MaxConsecutiveOne(start, end uint) uint {
	return b.consecutiveMaxCount(start, end, true)
}

// MaxConsecutiveZero
func (b *BitSet32) MaxConsecutiveZero(start, end uint) uint {
	return b.consecutiveMaxCount(start, end, false)
}

func (b *BitSet32) consecutiveMaxCount(start, end uint, flag bool) uint {
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

func xor(a, b bool) bool {
	return (a || b) && !(a && b)
}
