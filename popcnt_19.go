package bitset32

import "math/bits"

func popcntSlice(s []uint32) uint64 {
	var cnt int
	for _, x := range s {
		cnt += bits.OnesCount32(x)
	}
	return uint64(cnt)
}

func popcntMaskSlice(s, m []uint32) uint64 {
	var cnt int
	for i := range s {
		cnt += bits.OnesCount32(s[i] &^ m[i])
	}
	return uint64(cnt)
}

func popcntAndSlice(s, m []uint32) uint64 {
	var cnt int
	for i := range s {
		cnt += bits.OnesCount32(s[i] & m[i])
	}
	return uint64(cnt)
}

func popcntOrSlice(s, m []uint32) uint64 {
	var cnt int
	for i := range s {
		cnt += bits.OnesCount32(s[i] | m[i])
	}
	return uint64(cnt)
}

func popcntXorSlice(s, m []uint32) uint64 {
	var cnt int
	for i := range s {
		cnt += bits.OnesCount32(s[i] ^ m[i])
	}
	return uint64(cnt)
}
