# bitset32 

[![Test](https://github.com/bits-and-blooms/bitset/workflows/Test/badge.svg)](https://github.com/pointernil/bitset32/actions?query=workflow%3ATest)

[en English](./README.md)

## 简介

包 `bitset32` 修改自 `"github.com/bits-and-blooms/bitset"`，底层使用uint32存数据。`bitset32` 与 `bitset` 用法一致。

如非必要，请使用 ["github.com/bits-and-blooms/bitset"](https://github.com/bits-and-blooms/bitset)。

## Golang版本
```
go version go1.19.4 windows/amd64
```

## 安装
```
go get github.com/pointernil/bitset32
```

## 测试
```
go test
go test -cover
```

## 使用示意
```
package main

import (
	"fmt"
	"math/rand"

	"github.com/pointernil/bitset32"
)

func main() {
	fmt.Printf("! \n")
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