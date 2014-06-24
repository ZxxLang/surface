// Derived from Go's package runtime
// --------------------------------------------------------------------------
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Copyright 2014 The ZxxLang Authors. All rights reserved.

// +build !race

// 这些代码转化自 package runtime

package surface

// Via /pkg/runtime/runtime.h
type _Lock struct {
	key uintptr
}

// Via /pkg/runtime/chan.c
type _SudoG struct {
	_           *uintptr // g and selgen constitute
	selgen      uint32   // a weak pointer to g
	link        *_SudoG
	releasetime int64
	elem        *byte // data element
}

type _WaitQ struct {
	first *_SudoG
	last  *_SudoG
}

// Via /pkg/runtime/chan.c
type _Hchan struct {
	qcount   uint
	dataqsiz uint
	elemsize uint16
	pad      uint16
	closed   bool
	alg      *uintptr
	sendx    uint
	recvx    uint
	recvq    _WaitQ
	sendq    _WaitQ
	_Lock
}

// Via /pkg/runtime/hashmap.c
type _Hmap struct {
	count      uint
	flags      uint32
	hash0      uint32
	B          uint8
	keysize    uint8
	valuesize  uint8
	bucketsize uint8
	buckets    *byte
	oldbuckets *byte
	nevacuate  uintptr
}

func maplen(m IWord) int {
	if m == nil {
		return 0
	}
	p := (*_Hmap)(m)
	return int(p.count)
}

func chancap(ch IWord) int {
	if ch == nil {
		return 0
	}
	p := (*_Hchan)(ch)
	return int(p.dataqsiz)
}

func chanlen(ch IWord) int {
	if ch == nil {
		return 0
	}
	p := (*_Hchan)(ch)
	return int(p.qcount)
}
