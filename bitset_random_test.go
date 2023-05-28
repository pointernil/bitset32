package bitset32

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/bits-and-blooms/bitset"
)

var opc int
var opcT int

var ft string = "%v|pos:%9X|opc:%9X|result:%v\n"
var rt string = "%v|opc:%9d|pass:%9d\n"

var opNum = 100
var bitTestNum = 10000
var randNum = math.MaxInt32 / 2

func TestBitSet(t *testing.T) {
	var b32 = New(1)
	var b64 = bitset.New(1)
	res := true
	pos := uint(0)
	rand.Seed(time.Now().Unix())
	for j := 0; j < 1; j++ {
		// Test, Set,
		for i := 0; i < opNum; i++ {
			pos = uint(rand.Intn(randNum))
			b32 = b32.Set(uint(pos))
			b64 = b64.Set(uint(pos))
			res = b32.Test(pos) == b64.Test(pos)
			opc++
			if res {
				opcT++
			} else {
				t.Log(ft, time.Now(), pos, opc, res)
			}
		}
		// Clear
		for i := 0; i < opNum; i++ {
			pos = uint(rand.Intn(randNum))
			b32 = b32.Clear(uint(pos))
			b64 = b64.Clear(uint(pos))
			res = b32.Test(pos) == b64.Test(pos)
			opc++
			if res {
				opcT++
			} else {
				t.Logf(ft, time.Now().Unix(), pos, opc, res)
			}
		}
		// SetTo = Set + Clear
		for i := 0; i < opNum; i++ {
			pos = uint(rand.Intn(randNum))
			value := rand.Intn(randNum)%2 == 1
			b32 = b32.SetTo(uint(pos), value)
			b64 = b64.SetTo(uint(pos), value)
			res = b32.Test(pos) == b64.Test(pos)
			opc++
			if res {
				opcT++
			} else {
				t.Logf(ft, time.Now().Unix(), pos, opc, res)
			}
		}
		// Flip
		for i := 0; i < opNum; i++ {
			pos = uint(rand.Intn(randNum))
			b64 = b64.Flip(pos)
			b32 = b32.Flip(pos)
			res = isSameBitset(b32, b64)
			opc++
			if res {
				opcT++
			} else {
				t.Logf(ft, time.Now().Unix(), pos, opc, res)
			}
		}
		// Flip Range
		for i := 0; i < opNum; i++ {
			start, end := uint(rand.Intn(randNum)), uint(rand.Intn(randNum))
			if start > end {
				start, end = end, start
			}
			b64 = b64.FlipRange(start, end)
			b32 = b32.FlipRange(start, end)
			res = isSameBitset(b32, b64)
			opc++
			if res {
				opcT++
			} else {
				t.Logf(ft, time.Now().Unix(), pos, opc, res)
			}
		}
		// InsertAt
		for i := 0; i < opNum; i++ {
			pos = uint(rand.Intn(randNum))
			b64 = b64.InsertAt(pos)
			b32 = b32.InsertAt(pos)
			res = isSameBitset(b32, b64)
			opc++
			if res {
				opcT++
			} else {
				t.Logf(ft, time.Now().Unix(), pos, opc, res)
			}
		}
		// DeleteAt
		for i := 0; i < opNum; i++ {
			pos = uint(rand.Intn(randNum))
			if b64.Len() < pos || b32.Len() < pos {
				continue
			}
			b64 = b64.DeleteAt(pos)
			b32 = b32.DeleteAt(pos)
			res = isSameBitset(b32, b64)
			opc++
			if res {
				opcT++
			} else {
				t.Logf(ft, time.Now().Unix(), pos, opc, res)
			}
		}
		// Compact, Shrink
		for i := 0; i < opNum; i++ {
			b32 = b32.Compact()
			b64 = b64.Compact()
			res = isSameBitset(b32, b64)
			opc++
			if res {
				opcT++
			} else {
				t.Logf(ft, time.Now().Unix(), pos, opc, res)
			}
		}
	}
	// Compact
	b32 = b32.Compact()
	b64 = b64.Compact()
	res = isSameBitset(b32, b64)
	t.Log("Compact:", res)
	bs64 := &BitSet64{b64}
	t.Log("Max Count:", b32.MaxConsecutiveOne(0, b32.Len()), bs64.MaxConsecutiveOne(0, b64.Len()))
	t.Log("String:", b32.String() == b64.String())
	t.Logf(rt, time.Now().Unix(), opc, opcT)
}

func isSameBitset(b32 *BitSet32, b64 *bitset.BitSet) bool {
	if b32.Len() != b64.Len() {
		return false
	}
	for i := 0; i < bitTestNum; i++ {
		pos := uint(rand.Intn(randNum))
		if b32.Test(pos) != b64.Test(pos) {
			return false
		}
	}
	return true
}

/*
Running tool: C:\support\go\bin\go.exe test -timeout 30s -run ^TestBitSet$ bitset -v

=== RUN   TestBitSet
    d:\workspace\DataStruct\go\bitset\bitset_test.go:139: Max Count: 23067608 23067608
    d:\workspace\DataStruct\go\bitset\bitset_test.go:140: true
    d:\workspace\DataStruct\go\bitset\bitset_test.go:141: 1678626066|opc:      800|pass:      800
--- PASS: TestBitSet (23.61s)
PASS
ok  	bitset	24.243s
*/
