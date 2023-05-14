package bitset32

// FIXME: too slow
func (b *BitSet32) MaxConsecutiveOne() uint {
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
