// Derived from Go's package reflect
// --------------------------------------------------------------------------
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Copyright 2014 The ZxxLang Authors. All rights reserved.

// Package reflect implements run-time reflection, allowing a program to
// manipulate objects with arbitrary types.  The typical use is to take a value
// with static type interface{} and extract its dynamic type information by
// calling TypeIface, which returns a Type.
//
// A call to ValueIface returns a Value representing the run-time data.
// Zero takes a Type and returns a Value representing a zero value
// for that type.
//
// See "The Laws of Reflection" for an introduction to reflection in Go:
// http://golang.org/doc/articles/laws_of_reflection.html

/**
surface 使用了极不常规的方法, 可能造成损失. 如果您担心请不要使用 surface.
surface 包主要的代码拷贝自 Go 官方 reflect 包. 只是做了简单的重组和字段导出处理,
这样是为了便于访问属性, 不能用于更改, 那将造成损失.
如果 surface 不能跟进 Go 内部实现变更也将造成损失.
*/
package surface

import (
	"runtime"
	"strconv"
	"unsafe"
)

type IWord unsafe.Pointer

// EmptyInterface is the header for an interface{} value.
type EmptyInterface struct {
	Type *Type
	word IWord
}

// NonEmptyInterface is the header for a interface value with methods.
type NonEmptyInterface struct {
	// see ../runtime/iface.c:/Itab
	ITab *ITab
	word IWord
}

type ITab struct {
	Type       *InterfaceType // static interface type
	TargetType *Type          // dynamic concrete type
	Link       unsafe.Pointer
	Bad        int32
	Unused     int32
	Fun        [100000]unsafe.Pointer // method table
}

type flag uintptr

const (
	flagRO flag = 1 << iota
	flagIndir
	flagAddr
	flagMethod
	flagKindShift        = iota
	flagKindWidth        = 5 // there are 27 kinds
	flagKindMask    flag = 1<<flagKindWidth - 1
	flagMethodShift      = flagKindShift + flagKindWidth
)

type sur struct {
	// val holds the 1-word representation of the value.
	// If flag's flagIndir bit is set, then val is a pointer to the data.
	// Otherwise val is a word holding the actual data.
	// When the data is smaller than a word, it begins at
	// the first byte (in the memory address sense) of val.
	// We use unsafe.Pointer so that the garbage collector
	// knows that val could be a pointer.
	val unsafe.Pointer

	// Non-pointer-valued data.  When the data is smaller
	// than a word, it begins at the first byte (in the memory
	// address sense) of this field.
	// Valid when flagIndir is not set and typ.pointers() is false.
	scalar uintptr // go1.3

	// flag holds metadata about the value.
	// The lowest bits are flag bits:
	//	- flagRO: obtained via unexported field, so read-only
	//	- flagIndir: val holds a pointer to the data
	//	- flagAddr: v.CanAddr is true (implies flagIndir)
	//	- flagMethod: v is a method value.
	// The next five bits give the Kind of the value.
	// This repeats typ.Kind() except for method values.
	// The remaining 23+ bits give a method number for method values.
	// If flag.kind() != Func, code can assume that flagMethod is unset.
	// If typ.size > ptrSize, code can assume that flagIndir is set.
	flag

	// A method value represents a curried method invocation
	// like r.Read for some receiver r.  The typ+val+flag bits describe
	// the receiver r, but the flag's Kind bits say Func (methods are
	// functions), and the top bits of the flag give the method number
	// in r's type's method table.

	typ unsafe.Pointer // unsafe.Pointer(*Type), 冗余数据, 为简化计算
}

// mustBe panics if f's kind is not expected.
// Making this a method on flag instead of on Value
// (and embedding flag in Value) means that we can write
// the very clear v.mustBe(Bool) and have it compile into
// v.flag.mustBe(Bool), which will only bother to copy the
// single important word for the receiver.
func (f flag) mustBe(expected Kind) {
	k := f.Kind()
	if k != expected {
		panic(&ValueError{methodName(), k})
	}
}

// mustBeExported panics if f records that the value was obtained using
// an unexported field.
func (f flag) mustBeExported() {
	if f == 0 {
		panic(&ValueError{methodName(), 0})
	}
	if f&flagRO != 0 {
		panic("surface: " + methodName() + " using value obtained using unexported field")
	}
}

// mustBeAssignable panics if f records that the value is not assignable,
// which is to say that either it was obtained using an unexported field
// or it is not addressable.
func (f flag) mustBeAssignable() {
	if f == 0 {
		panic(&ValueError{methodName(), KInvalid})
	}
	// Assignable if addressable and not read-only.
	if f&flagRO != 0 {
		panic("surface: " + methodName() + " using value obtained using unexported field")
	}
	if f&flagAddr == 0 {
		panic("surface: " + methodName() + " using unaddressable value")
	}
}

func (f flag) CanSet() bool {
	return f&(flagAddr|flagRO) == flagAddr
}
func (f flag) CanAddr() bool {
	return f&flagAddr != 0
}
func (f flag) CanInterface() bool {
	if f == 0 {
		return false
	}
	return f&flagRO == 0
}

func (f flag) Exported() bool {
	return f != 0 && f&flagRO == 0
}

func (f flag) Kind() Kind {
	return Kind((f >> flagKindShift) & flagKindMask)
}
func (f flag) IsIndir() bool {
	return f&flagIndir != 0
}
func (f flag) IsValid() bool {
	return f != 0
}

func (s sur) IsNil() bool {
	if s.val == nil {
		return true
	}
	f := s.flag
	switch f.Kind() {
	case KChan, KFunc, KMap, KPtr:
		if f&flagMethod != 0 {
			return false
		}
		if f&flagIndir != 0 {
			return *(*unsafe.Pointer)(s.val) == nil
		}
	case KInterface, KSlice:
		// Both interface and slice are nil if first word is 0.
		// Both are always bigger than a word; assume flagIndir.
		return *(*unsafe.Pointer)(s.val) == nil
	}
	return false
}

func (s sur) IWord() IWord {
	typ := (*Type)(unsafe.Pointer(s.typ))
	if s.flag&flagIndir != 0 && typ.Size <= ptrSize {
		// Have indirect but want direct word.
		return loadIword(s.val, typ.Size)
	}
	return IWord(s.val)
}

// loadIword loads n bytes at p from memory into an iword.
func loadIword(p unsafe.Pointer, n uintptr) IWord {
	// Run the copy ourselves instead of calling memmove
	// to avoid moving w to the heap.
	var w IWord
	switch n {
	default:
		panic("surface: internal error: loadIword of " + strconv.Itoa(int(n)) + "-byte value")
	case 0:
	case 1:
		*(*uint8)(unsafe.Pointer(&w)) = *(*uint8)(p)
	case 2:
		*(*uint16)(unsafe.Pointer(&w)) = *(*uint16)(p)
	case 3:
		*(*[3]byte)(unsafe.Pointer(&w)) = *(*[3]byte)(p)
	case 4:
		*(*uint32)(unsafe.Pointer(&w)) = *(*uint32)(p)
	case 5:
		*(*[5]byte)(unsafe.Pointer(&w)) = *(*[5]byte)(p)
	case 6:
		*(*[6]byte)(unsafe.Pointer(&w)) = *(*[6]byte)(p)
	case 7:
		*(*[7]byte)(unsafe.Pointer(&w)) = *(*[7]byte)(p)
	case 8:
		*(*uint64)(unsafe.Pointer(&w)) = *(*uint64)(p)
	}
	return w
}

// methodName returns the name of the calling method,
// assumed to be two stack frames above.
func methodName() string {
	pc, _, _, _ := runtime.Caller(2)
	f := runtime.FuncForPC(pc)
	if f == nil {
		return "unknown method"
	}
	return f.Name()
}

// TODO: This will have to go away when
// the new gc goes in.
func memmove(adst, asrc unsafe.Pointer, n uintptr) {
	dst := uintptr(adst)
	src := uintptr(asrc)
	switch {
	case src < dst && src+n > dst:
		// byte copy backward
		// careful: i is unsigned
		for i := n; i > 0; {
			i--
			*(*byte)(unsafe.Pointer(dst + i)) = *(*byte)(unsafe.Pointer(src + i))
		}
	case (n|src|dst)&(ptrSize-1) != 0:
		// byte copy forward
		for i := uintptr(0); i < n; i++ {
			*(*byte)(unsafe.Pointer(dst + i)) = *(*byte)(unsafe.Pointer(src + i))
		}
	default:
		// word copy forward
		for i := uintptr(0); i < n; i += ptrSize {
			*(*uintptr)(unsafe.Pointer(dst + i)) = *(*uintptr)(unsafe.Pointer(src + i))
		}
	}
}
