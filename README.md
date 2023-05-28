# bitset32

[![Test](https://github.com/bits-and-blooms/bitset/workflows/Test/badge.svg)](https://github.com/pointernil/bitset32/actions?query=workflow%3ATest)

[zh_CN 简体中文](./README_zh_CN.md)

## Description

Package bitset32 modified from `"github.com/bits-and-blooms/bitset"`
implements bitset with uint32. Both packages are used in the same way.

If not necessary, it is highly recommended to use 
["github.com/bits-and-blooms/bitset"](https://github.com/bits-and-blooms/bitset).

## Go version
```
go version go1.19.4 windows/amd64
```

## Install
```
go get github.com/pointernil/bitset32
```

## Testing
```
go test
go test -cover
```

## Usage
```
package main

import (
	"fmt"
	"math/rand"

	"github.com/pointernil/bitset32"
)

func main() {
	fmt.Printf("Hello from BitSet!\n")
	var b bitset32.BitSet32
	// play some Go Fish
	for i := 0; i < 100; i++ {
		card1 := uint(rand.Intn(52))
		card2 := uint(rand.Intn(52))
		b.Set(card1)
		if b.Test(card2) {
			fmt.Println("Go Fish!")
		}
		b.Clear(card1)
	}

	// Chaining
	b.Set(10).Set(11)

	for i, e := b.NextSet(0); e; i, e = b.NextSet(i + 1) {
		fmt.Println("The following bit is set:", i)
	}
	if b.Intersection(bitset32.New(100).Set(10)).Count() == 1 {
		fmt.Println("Intersection works.")
	} else {
		fmt.Println("Intersection doesn't work???")
	}
}
```