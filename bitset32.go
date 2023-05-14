/*
Package bitset implements bitsets, a mapping
between non-negative integers and boolean values. It should be more
efficient than map[uint] bool.

It provides methods for setting, clearing, flipping, and testing
individual integers.

But it also provides set intersection, union, difference,
complement, and symmetric operations, as well as tests to
check whether any, all, or no bits are set, and querying a
bitset's current length and number of positive bits.

BitSets are expanded to the size of the largest set bit; the
memory allocation is approximately Max bits, where Max is
the largest set bit. BitSets are never shrunk. On creation,
a hint can be given for the number of bits that will be used.

Many of the methods, including Set,Clear, and Flip, return
a BitSet pointer, which allows for chaining.

Example use:

	import "bitset"
	var b BitSet
	b.Set(10).Set(11)
	if b.Test(1000) {
		b.Clear(1000)
	}
	if B.Intersection(bitset.New(100).Set(10)).Count() > 1 {
		fmt.Println("Intersection works.")
	}

As an alternative to BitSets, one should check out the 'big' package,
which provides a (less set-theoretical) view of bitsets.
*/
package bitset32

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/bits"
	"strconv"
)

// the wordSize of a bit set
const wordSize = uint(32)

// log2WordSize is lg(wordSize)
const log2WordSize = uint(5)

// allBits has every bit set
const allBits uint32 = 0xffffffff

// TODO BUGFIX
// default binary BigEndian
var binaryOrder binary.ByteOrder = binary.BigEndian

// A BitSet is a set of bits. The zero value of a BitSet is an empty set of length 0.
type BitSet32 struct {
	length uint
	set    []uint32
}

// Error is used to distinguish errors (panics) generated in this package.
type Error string

// safeSet will fixup b.set to be non-nil and return the field value
func (b *BitSet32) safeSet() []uint32 {
	if b.set == nil {
		b.set = make([]uint32, wordsNeeded(0))
	}
	return b.set
}

// SetBitsetFrom fills the bitset with an array of integers without creating a new BitSet instance
func (b *BitSet32) SetBitsetFrom(buf []uint32) {
	b.length = uint(len(buf)) * 32
	b.set = buf
}

// From is a constructor used to create a BitSet from an array of integers
func From(buf []uint32) *BitSet32 {
	return FromWithLength(uint(len(buf))*32, buf)
}

// FromWithLength constructs from an array of integers and length.
func FromWithLength(len uint, set []uint32) *BitSet32 {
	return &BitSet32{len, set}
}

// Bytes returns the bitset as array of integers
func (b *BitSet32) Bytes() []uint32 {
	return b.set
}

// wordsNeeded calculates the number of words needed for i bits
func wordsNeeded(i uint) int {
	if i > (Cap() - wordSize + 1) {
		return int(Cap() >> log2WordSize)
	}
	return int((i + (wordSize - 1)) >> log2WordSize)
}

// wordsNeededUnbound calculates the number of words needed for i bits, possibly exceeding the capacity.
// This function is useful if you know that the capacity cannot be exceeded (e.g., you have an existing bitmap).
func wordsNeededUnbound(i uint) int {
	return int((i + (wordSize - 1)) >> log2WordSize)
}

// wordsIndex calculates the index of words in a `uint64`
func wordsIndex(i uint) uint {
	return i & (wordSize - 1)
}

// New creates a new BitSet with a hint that length bits will be required
func New(length uint) (bset *BitSet32) {
	defer func() {
		if r := recover(); r != nil {
			bset = &BitSet32{
				0,
				make([]uint32, 0),
			}
		}
	}()

	bset = &BitSet32{
		length,
		make([]uint32, wordsNeeded(length)),
	}

	return bset
}

// Cap returns the total possible capacity, or number of bits
func Cap() uint {
	return ^uint(0)
}

// Len returns the number of bits in the BitSet.
// Note the difference to method Count, see example.
func (b *BitSet32) Len() uint {
	return b.length
}

// extendSet adds additional words to incorporate new bits if needed
func (b *BitSet32) extendSet(i uint) {
	if i >= Cap() {
		panic("You are exceeding the capacity")
	}
	nsize := wordsNeeded(i + 1)
	if b.set == nil {
		b.set = make([]uint32, nsize)
	} else if cap(b.set) >= nsize {
		b.set = b.set[:nsize] // fast resize
	} else if len(b.set) < nsize {
		newset := make([]uint32, nsize, 2*nsize) // increase capacity 2x
		copy(newset, b.set)
		b.set = newset
	}
	b.length = i + 1
}

// Test whether bit i is set.
func (b *BitSet32) Test(i uint) bool {
	if i >= b.length {
		return false
	}
	return b.set[i>>log2WordSize]&(1<<wordsIndex(i)) != 0
}

// Set bit i to 1, the capacity of the bitset is automatically
// increased accordingly.
// If i>= Cap(), this function will panic.
// Warning: using a very large value for 'i'
// may lead to a memory shortage and a panic: the caller is responsible
// for providing sensible parameters in line with their memory capacity.
func (b *BitSet32) Set(i uint) *BitSet32 {
	if i >= b.length { // if we need more bits, make 'em
		b.extendSet(i)
	}
	b.set[i>>log2WordSize] |= 1 << wordsIndex(i)
	return b
}

// Clear bit i to 0
func (b *BitSet32) Clear(i uint) *BitSet32 {
	if i >= b.length {
		return b
	}
	b.set[i>>log2WordSize] &^= 1 << wordsIndex(i)
	return b
}

// SetTo sets bit i to value.
// If i>= Cap(), this function will panic.
// Warning: using a very large value for 'i'
// may lead to a memory shortage and a panic: the caller is responsible
// for providing sensible parameters in line with their memory capacity.
func (b *BitSet32) SetTo(i uint, value bool) *BitSet32 {
	if value {
		return b.Set(i)
	}
	return b.Clear(i)
}

// Flip bit at i.
// If i>= Cap(), this function will panic.
// Warning: using a very large value for 'i'
// may lead to a memory shortage and a panic: the caller is responsible
// for providing sensible parameters in line with their memory capacity.
func (b *BitSet32) Flip(i uint) *BitSet32 {
	if i >= b.length {
		return b.Set(i)
	}
	b.set[i>>log2WordSize] ^= 1 << wordsIndex(i)
	return b
}

// FlipRange bit in [start, end).
// If end>= Cap(), this function will panic.
// Warning: using a very large value for 'end'
// may lead to a memory shortage and a panic: the caller is responsible
// for providing sensible parameters in line with their memory capacity.
func (b *BitSet32) FlipRange(start, end uint) *BitSet32 {
	if start >= end {
		return b
	}
	if end-1 >= b.length { // if we need more bits, make 'em
		b.extendSet(end - 1)
	}
	var startWord uint = start >> log2WordSize
	var endWord uint = end >> log2WordSize
	b.set[startWord] ^= ^(^uint32(0) << wordsIndex(start))
	for i := startWord; i < endWord; i++ {
		b.set[i] = ^b.set[i]
	}
	if end&(wordSize-1) != 0 {
		b.set[endWord] ^= ^uint32(0) >> wordsIndex(-end)
	}
	return b
}

// Shrink shrinks BitSet so that the provided value is the last possible
// set value. It clears all bits > the provided index and reduces the size
// and length of the set.
//
// Note that the parameter value is not the new length in bits: it is the
// maximal value that can be stored in the bitset after the function call.
// The new length in bits is the parameter value + 1. Thus it is not possible
// to use this function to set the length to 0, the minimal value of the length
// after this function call is 1.
//
// A new slice is allocated to store the new bits, so you may see an increase in
// memory usage until the GC runs. Normally this should not be a problem, but if you
// have an extremely large BitSet its important to understand that the old BitSet will
// remain in memory until the GC frees it.
func (b *BitSet32) Shrink(lastbitindex uint) *BitSet32 {
	length := lastbitindex + 1
	idx := wordsNeeded(length)
	if idx > len(b.set) {
		return b
	}
	shrunk := make([]uint32, idx)
	copy(shrunk, b.set[:idx])
	b.set = shrunk
	b.length = length
	lastWordUsedBits := length % 32
	if lastWordUsedBits != 0 {
		b.set[idx-1] &= allBits >> uint32(32-wordsIndex(lastWordUsedBits))
	}
	return b
}

// Compact shrinks BitSet to so that we preserve all set bits, while minimizing
// memory usage. Compact calls Shrink.
func (b *BitSet32) Compact() *BitSet32 {
	idx := len(b.set) - 1
	for ; idx >= 0 && b.set[idx] == 0; idx-- {
	}
	newlength := uint((idx + 1) << log2WordSize)
	if newlength >= b.length {
		return b // nothing to do
	}
	if newlength > 0 {
		return b.Shrink(newlength - 1)
	}
	// TODO: FIX
	// We preserve one word
	return b.Shrink(31)
}

// InsertAt takes an index which indicates where a bit should be
// inserted. Then it shifts all the bits in the set to the left by 1, starting
// from the given index position, and sets the index position to 0.
//
// Depending on the size of your BitSet, and where you are inserting the new entry,
// this method could be extremely slow and in some cases might cause the entire BitSet
// to be recopied.
func (b *BitSet32) InsertAt(idx uint) *BitSet32 {
	insertAtElement := idx >> log2WordSize

	// if length of set is a multiple of wordSize we need to allocate more space first
	if b.isLenExactMultiple() {
		b.set = append(b.set, uint32(0))
	}

	var i uint
	for i = uint(len(b.set) - 1); i > insertAtElement; i-- {
		// all elements above the position where we want to insert can simply by shifted
		b.set[i] <<= 1

		// we take the most significant bit of the previous element and set it as
		// the least significant bit of the current element
		// TODO: FIX
		b.set[i] |= (b.set[i-1] & 0x80000000) >> 31
	}

	// generate a mask to extract the data that we need to shift left
	// within the element where we insert a bit
	dataMask := uint32(1)<<uint32(wordsIndex(idx)) - 1

	// extract that data that we'll shift
	data := b.set[i] & (^dataMask)

	// set the positions of the data mask to 0 in the element where we insert
	b.set[i] &= dataMask

	// shift data mask to the left and insert its data to the slice element
	b.set[i] |= data << 1

	// add 1 to length of BitSet
	b.length++

	return b
}

// String creates a string representation of the Bitmap
func (b *BitSet32) String() string {
	// follows code from https://github.com/RoaringBitmap/roaring
	var buffer bytes.Buffer
	start := []byte("{")
	buffer.Write(start)
	counter := 0
	i, e := b.NextSet(0)
	for e {
		counter = counter + 1
		// to avoid exhausting the memory
		if counter > 0x40000 {
			buffer.WriteString("...")
			break
		}
		buffer.WriteString(strconv.FormatInt(int64(i), 10))
		i, e = b.NextSet(i + 1)
		if e {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("}")
	return buffer.String()
}

// DeleteAt deletes the bit at the given index position from
// within the bitset
// All the bits residing on the left of the deleted bit get
// shifted right by 1
// The running time of this operation may potentially be
// relatively slow, O(length)
func (b *BitSet32) DeleteAt(i uint) *BitSet32 {
	// the index of the slice element where we'll delete a bit
	deleteAtElement := i >> log2WordSize

	// generate a mask for the data that needs to be shifted right
	// within that slice element that gets modified
	dataMask := ^((uint32(1) << wordsIndex(i)) - 1)

	// extract the data that we'll shift right from the slice element
	data := b.set[deleteAtElement] & dataMask

	// set the masked area to 0 while leaving the rest as it is
	b.set[deleteAtElement] &= ^dataMask

	// shift the previously extracted data to the right and then
	// set it in the previously masked area
	b.set[deleteAtElement] |= (data >> 1) & dataMask

	// loop over all the consecutive slice elements to copy each
	// lowest bit into the highest position of the previous element,
	// then shift the entire content to the right by 1
	for i := int(deleteAtElement) + 1; i < len(b.set); i++ {
		b.set[i-1] |= (b.set[i] & 1) << 31
		b.set[i] >>= 1
	}

	b.length = b.length - 1

	return b
}

// NextSet returns the next bit set from the specified index,
// including possibly the current index
// along with an error code (true = valid, false = no set bit found)
// for i,e := v.NextSet(0); e; i,e = v.NextSet(i + 1) {...}
//
// Users concerned with performance may want to use NextSetMany to
// retrieve several values at once.
func (b *BitSet32) NextSet(i uint) (uint, bool) {
	x := int(i >> log2WordSize)
	if x >= len(b.set) {
		return 0, false
	}
	w := b.set[x]
	w = w >> wordsIndex(i)
	if w != 0 {
		return i + uint(bits.TrailingZeros32(w)), true
	}
	x = x + 1
	for x < len(b.set) {
		if b.set[x] != 0 {
			return uint(x)*wordSize + uint(bits.TrailingZeros32(b.set[x])), true
		}
		x = x + 1

	}
	return 0, false
}

// NextSetMany returns many next bit sets from the specified index,
// including possibly the current index and up to cap(buffer).
// If the returned slice has len zero, then no more set bits were found
//
//	buffer := make([]uint, 256) // this should be reused
//	j := uint(0)
//	j, buffer = bitmap.NextSetMany(j, buffer)
//	for ; len(buffer) > 0; j, buffer = bitmap.NextSetMany(j,buffer) {
//	 for k := range buffer {
//	  do something with buffer[k]
//	 }
//	 j += 1
//	}
//
// It is possible to retrieve all set bits as follow:
//
//	indices := make([]uint, bitmap.Count())
//	bitmap.NextSetMany(0, indices)
//
// However if bitmap.Count() is large, it might be preferable to
// use several calls to NextSetMany, for performance reasons.
func (b *BitSet32) NextSetMany(i uint, buffer []uint) (uint, []uint) {
	myanswer := buffer
	capacity := cap(buffer)
	x := int(i >> log2WordSize)
	if x >= len(b.set) || capacity == 0 {
		return 0, myanswer[:0]
	}
	skip := wordsIndex(i)
	word := b.set[x] >> skip
	myanswer = myanswer[:capacity]
	size := int(0)
	for word != 0 {
		r := uint(bits.TrailingZeros32(word))
		t := word & ((^word) + 1)
		myanswer[size] = r + i
		size++
		if size == capacity {
			goto End
		}
		word = word ^ t
	}
	x++
	for idx, word := range b.set[x:] {
		for word != 0 {
			r := uint(bits.TrailingZeros32(word))
			t := word & ((^word) + 1)
			myanswer[size] = r + (uint(x+idx) << 6)
			size++
			if size == capacity {
				goto End
			}
			word = word ^ t
		}
	}
End:
	if size > 0 {
		return myanswer[size-1], myanswer[:size]
	}
	return 0, myanswer[:0]
}

// NextClear returns the next clear bit from the specified index,
// including possibly the current index
// along with an error code (true = valid, false = no bit found i.e. all bits are set)
func (b *BitSet32) NextClear(i uint) (uint, bool) {
	x := int(i >> log2WordSize)
	if x >= len(b.set) {
		return 0, false
	}
	w := b.set[x]
	w = w >> wordsIndex(i)
	wA := allBits >> wordsIndex(i)
	index := i + uint(bits.TrailingZeros32(^w))
	if w != wA && index < b.length {
		return index, true
	}
	x++
	for x < len(b.set) {
		index = uint(x)*wordSize + uint(bits.TrailingZeros32(^b.set[x]))
		if b.set[x] != allBits && index < b.length {
			return index, true
		}
		x++
	}
	return 0, false
}

// ClearAll clears the entire BitSet
func (b *BitSet32) ClearAll() *BitSet32 {
	if b != nil && b.set != nil {
		for i := range b.set {
			b.set[i] = 0
		}
	}
	return b
}

// wordCount returns the number of words used in a bit set
func (b *BitSet32) wordCount() int {
	return wordsNeededUnbound(b.length)
}

// Clone this BitSet
func (b *BitSet32) Clone() *BitSet32 {
	c := New(b.length)
	if b.set != nil { // Clone should not modify current object
		copy(c.set, b.set)
	}
	return c
}

// Copy into a destination BitSet using the Go array copy semantics:
// the number of bits copied is the minimum of the number of bits in the current
// BitSet (Len()) and the destination Bitset.
// We return the number of bits copied in the destination BitSet.
func (b *BitSet32) Copy(c *BitSet32) (count uint) {
	if c == nil {
		return
	}
	if b.set != nil { // Copy should not modify current object
		copy(c.set, b.set)
	}
	count = c.length
	if b.length < c.length {
		count = b.length
	}
	// Cleaning the last word is needed to keep the invariant that other functions, such as Count, require
	// that any bits in the last word that would exceed the length of the bitmask are set to 0.
	c.cleanLastWord()
	return
}

// CopyFull copies into a destination BitSet such that the destination is
// identical to the source after the operation, allocating memory if necessary.
func (b *BitSet32) CopyFull(c *BitSet32) {
	if c == nil {
		return
	}
	c.length = b.length
	if len(b.set) == 0 {
		if c.set != nil {
			c.set = c.set[:0]
		}
	} else {
		if cap(c.set) < len(b.set) {
			c.set = make([]uint32, len(b.set))
		} else {
			c.set = c.set[:len(b.set)]
		}
		copy(c.set, b.set)
	}
}

// Count (number of set bits).
// Also known as "popcount" or "population count".
func (b *BitSet32) Count() uint {
	if b != nil && b.set != nil {
		return uint(popcntSlice(b.set))
	}
	return 0
}

// Equal tests the equivalence of two BitSets.
// False if they are of different sizes, otherwise true
// only if all the same bits are set
func (b *BitSet32) Equal(c *BitSet32) bool {
	if c == nil || b == nil {
		return c == b
	}
	if b.length != c.length {
		return false
	}
	if b.length == 0 { // if they have both length == 0, then could have nil set
		return true
	}
	wn := b.wordCount()
	for p := 0; p < wn; p++ {
		if c.set[p] != b.set[p] {
			return false
		}
	}
	return true
}

func panicIfNull(b *BitSet32) {
	if b == nil {
		panic(Error("BitSet must not be null"))
	}
}

// Difference of base set and other set
// This is the BitSet equivalent of &^ (and not)
func (b *BitSet32) Difference(compare *BitSet32) (result *BitSet32) {
	panicIfNull(b)
	panicIfNull(compare)
	result = b.Clone() // clone b (in case b is bigger than compare)
	l := int(compare.wordCount())
	if l > int(b.wordCount()) {
		l = int(b.wordCount())
	}
	for i := 0; i < l; i++ {
		result.set[i] = b.set[i] &^ compare.set[i]
	}
	return
}

// DifferenceCardinality computes the cardinality of the differnce
func (b *BitSet32) DifferenceCardinality(compare *BitSet32) uint {
	panicIfNull(b)
	panicIfNull(compare)
	l := int(compare.wordCount())
	if l > int(b.wordCount()) {
		l = int(b.wordCount())
	}
	cnt := uint64(0)
	cnt += popcntMaskSlice(b.set[:l], compare.set[:l])
	cnt += popcntSlice(b.set[l:])
	return uint(cnt)
}

// InPlaceDifference computes the difference of base set and other set
// This is the BitSet equivalent of &^ (and not)
func (b *BitSet32) InPlaceDifference(compare *BitSet32) {
	panicIfNull(b)
	panicIfNull(compare)
	l := int(compare.wordCount())
	if l > int(b.wordCount()) {
		l = int(b.wordCount())
	}
	for i := 0; i < l; i++ {
		b.set[i] &^= compare.set[i]
	}
}

// Convenience function: return two bitsets ordered by
// increasing length. Note: neither can be nil
func sortByLength(a *BitSet32, b *BitSet32) (ap *BitSet32, bp *BitSet32) {
	if a.length <= b.length {
		ap, bp = a, b
	} else {
		ap, bp = b, a
	}
	return
}

// Intersection of base set and other set
// This is the BitSet equivalent of & (and)
func (b *BitSet32) Intersection(compare *BitSet32) (result *BitSet32) {
	panicIfNull(b)
	panicIfNull(compare)
	b, compare = sortByLength(b, compare)
	result = New(b.length)
	for i, word := range b.set {
		result.set[i] = word & compare.set[i]
	}
	return
}

// IntersectionCardinality computes the cardinality of the union
func (b *BitSet32) IntersectionCardinality(compare *BitSet32) uint {
	panicIfNull(b)
	panicIfNull(compare)
	b, compare = sortByLength(b, compare)
	cnt := popcntAndSlice(b.set, compare.set)
	return uint(cnt)
}

// InPlaceIntersection destructively computes the intersection of
// base set and the compare set.
// This is the BitSet equivalent of & (and)
func (b *BitSet32) InPlaceIntersection(compare *BitSet32) {
	panicIfNull(b)
	panicIfNull(compare)
	l := int(compare.wordCount())
	if l > int(b.wordCount()) {
		l = int(b.wordCount())
	}
	for i := 0; i < l; i++ {
		b.set[i] &= compare.set[i]
	}
	for i := l; i < len(b.set); i++ {
		b.set[i] = 0
	}
	if compare.length > 0 {
		if compare.length-1 >= b.length {
			b.extendSet(compare.length - 1)
		}
	}
}

// Union of base set and other set
// This is the BitSet equivalent of | (or)
func (b *BitSet32) Union(compare *BitSet32) (result *BitSet32) {
	panicIfNull(b)
	panicIfNull(compare)
	b, compare = sortByLength(b, compare)
	result = compare.Clone()
	for i, word := range b.set {
		result.set[i] = word | compare.set[i]
	}
	return
}

// UnionCardinality computes the cardinality of the uniton of the base set
// and the compare set.
func (b *BitSet32) UnionCardinality(compare *BitSet32) uint {
	panicIfNull(b)
	panicIfNull(compare)
	b, compare = sortByLength(b, compare)
	cnt := popcntOrSlice(b.set, compare.set)
	if len(compare.set) > len(b.set) {
		cnt += popcntSlice(compare.set[len(b.set):])
	}
	return uint(cnt)
}

// InPlaceUnion creates the destructive union of base set and compare set.
// This is the BitSet equivalent of | (or).
func (b *BitSet32) InPlaceUnion(compare *BitSet32) {
	panicIfNull(b)
	panicIfNull(compare)
	l := int(compare.wordCount())
	if l > int(b.wordCount()) {
		l = int(b.wordCount())
	}
	if compare.length > 0 && compare.length-1 >= b.length {
		b.extendSet(compare.length - 1)
	}
	for i := 0; i < l; i++ {
		b.set[i] |= compare.set[i]
	}
	if len(compare.set) > l {
		for i := l; i < len(compare.set); i++ {
			b.set[i] = compare.set[i]
		}
	}
}

// SymmetricDifference of base set and other set
// This is the BitSet equivalent of ^ (xor)
func (b *BitSet32) SymmetricDifference(compare *BitSet32) (result *BitSet32) {
	panicIfNull(b)
	panicIfNull(compare)
	b, compare = sortByLength(b, compare)
	// compare is bigger, so clone it
	result = compare.Clone()
	for i, word := range b.set {
		result.set[i] = word ^ compare.set[i]
	}
	return
}

// SymmetricDifferenceCardinality computes the cardinality of the symmetric difference
func (b *BitSet32) SymmetricDifferenceCardinality(compare *BitSet32) uint {
	panicIfNull(b)
	panicIfNull(compare)
	b, compare = sortByLength(b, compare)
	cnt := popcntXorSlice(b.set, compare.set)
	if len(compare.set) > len(b.set) {
		cnt += popcntSlice(compare.set[len(b.set):])
	}
	return uint(cnt)
}

// InPlaceSymmetricDifference creates the destructive SymmetricDifference of base set and other set
// This is the BitSet equivalent of ^ (xor)
func (b *BitSet32) InPlaceSymmetricDifference(compare *BitSet32) {
	panicIfNull(b)
	panicIfNull(compare)
	l := int(compare.wordCount())
	if l > int(b.wordCount()) {
		l = int(b.wordCount())
	}
	if compare.length > 0 && compare.length-1 >= b.length {
		b.extendSet(compare.length - 1)
	}
	for i := 0; i < l; i++ {
		b.set[i] ^= compare.set[i]
	}
	if len(compare.set) > l {
		for i := l; i < len(compare.set); i++ {
			b.set[i] = compare.set[i]
		}
	}
}

// Is the length an exact multiple of word sizes?
func (b *BitSet32) isLenExactMultiple() bool {
	return wordsIndex(b.length) == 0
}

// Clean last word by setting unused bits to 0
func (b *BitSet32) cleanLastWord() {
	if !b.isLenExactMultiple() {
		b.set[len(b.set)-1] &= allBits >> (wordSize - wordsIndex(b.length))
	}
}

// Complement computes the (local) complement of a bitset (up to length bits)
func (b *BitSet32) Complement() (result *BitSet32) {
	panicIfNull(b)
	result = New(b.length)
	for i, word := range b.set {
		result.set[i] = ^word
	}
	result.cleanLastWord()
	return
}

// All returns true if all bits are set, false otherwise. Returns true for
// empty sets.
func (b *BitSet32) All() bool {
	panicIfNull(b)
	return b.Count() == b.length
}

// None returns true if no bit is set, false otherwise. Returns true for
// empty sets.
func (b *BitSet32) None() bool {
	panicIfNull(b)
	if b != nil && b.set != nil {
		for _, word := range b.set {
			if word > 0 {
				return false
			}
		}
	}
	return true
}

// Any returns true if any bit is set, false otherwise
func (b *BitSet32) Any() bool {
	panicIfNull(b)
	return !b.None()
}

// IsSuperSet returns true if this is a superset of the other set
func (b *BitSet32) IsSuperSet(other *BitSet32) bool {
	for i, e := other.NextSet(0); e; i, e = other.NextSet(i + 1) {
		if !b.Test(i) {
			return false
		}
	}
	return true
}

// IsStrictSuperSet returns true if this is a strict superset of the other set
func (b *BitSet32) IsStrictSuperSet(other *BitSet32) bool {
	return b.Count() > other.Count() && b.IsSuperSet(other)
}

// DumpAsBits dumps a bit set as a string of bits
func (b *BitSet32) DumpAsBits() string {
	if b.set == nil {
		return "."
	}
	buffer := bytes.NewBufferString("")
	i := len(b.set) - 1
	for ; i >= 0; i-- {
		fmt.Fprintf(buffer, "%064b.", b.set[i])
	}
	return buffer.String()
}

// BinaryStorageSize returns the binary storage requirements
func (b *BitSet32) BinaryStorageSize() int {
	nWords := b.wordCount()
	return binary.Size(uint64(0)) + binary.Size(b.set[:nWords])
}

// WriteTo writes a BitSet to a stream
func (b *BitSet32) WriteTo(stream io.Writer) (int64, error) {
	length := uint64(b.length)

	// Write length
	err := binary.Write(stream, binaryOrder, length)
	if err != nil {
		return 0, err
	}

	// Write set
	// current implementation of bufio.Writer is more memory efficient than
	// binary.Write for large set
	writer := bufio.NewWriter(stream)
	var item = make([]byte, binary.Size(uint32(0))) // for serializing one uint32
	nWords := b.wordCount()
	for i := range b.set[:nWords] {
		binaryOrder.PutUint32(item, b.set[i])
		if nn, err := writer.Write(item); err != nil {
			return int64(i*binary.Size(uint32(0)) + nn), err
		}
	}

	err = writer.Flush()
	return int64(b.BinaryStorageSize()), err
}

// ReadFrom reads a BitSet from a stream written using WriteTo
func (b *BitSet32) ReadFrom(stream io.Reader) (int64, error) {
	var length uint64

	// Read length first
	err := binary.Read(stream, binaryOrder, &length)
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return 0, err
	}
	newset := New(uint(length))

	if uint64(newset.length) != length {
		return 0, errors.New("unmarshalling error: type mismatch")
	}

	var item [4]byte
	nWords := wordsNeeded(uint(length))
	reader := bufio.NewReader(io.LimitReader(stream, 4*int64(nWords)))
	for i := 0; i < nWords; i++ {
		if _, err := io.ReadFull(reader, item[:]); err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return 0, err
		}
		newset.set[i] = binaryOrder.Uint32(item[:])
	}

	*b = *newset
	return int64(b.BinaryStorageSize()), nil
}
