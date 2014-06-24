// Derived from Go's package reflect
// --------------------------------------------------------------------------
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Copyright 2014 The ZxxLang Authors. All rights reserved.

package surface

import (
	"unsafe"
)

const bigEndian = false // can be smarter if we find a big-endian machine
const ptrSize = unsafe.Sizeof((*byte)(nil))
const cannotSet = "cannot set value obtained from unexported struct field"

// Value is the reflection interface to a Go value.
//
// Not all methods apply to all kinds of values.  Restrictions,
// if any, are noted in the documentation for each method.
// Use the Kind method to find out the kind of value before
// calling kind-specific methods.  Calling a method
// inappropriate to the kind of type causes a run time panic.
//
// The zero Value represents no value.
// Its IsValid method returns false, its Kind method returns Invalid,
// its String method returns "<invalid Value>", and all other methods panic.
// Most functions and methods never return an invalid value.
// If one does, its documentation states the conditions explicitly.
//
// A Value can be used concurrently by multiple goroutines provided that
// the underlying Go value can be used concurrently for the equivalent
// direct operations.
type Value struct {
	// typ holds the type of the value represented by a Value.
	Type *Type
	sur
}

type Array struct {
	Type *ArrayType
	sur
}

type Chan struct {
	Type *ChanType
	sur
}
type Func struct {
	Type *FuncType
	sur
}
type Interface struct {
	Type *InterfaceType
	sur
	TargetType *Type
}
type Map struct {
	Type *MapType
	sur
}
type Ptr struct {
	Type *PtrType
	sur
}
type Slice struct {
	Type *SliceType
	sur
}
type Struct struct {
	Type *StructType
	sur
}

func (v Value) Array() Array {
	return Array{v.Type.Array(), v.sur}
}
func (v Value) Chan() Chan {
	return Chan{v.Type.Chan(), v.sur}
}
func (v Value) Func() Func {
	return Func{v.Type.Func(), v.sur}
}
func (v Value) Map() Map {
	return Map{v.Type.Map(), v.sur}
}
func (v Value) Ptr() Ptr {
	return Ptr{v.Type.Ptr(), v.sur}
}
func (v Value) Slice() Slice {
	return Slice{v.Type.Slice(), v.sur}
}
func (v Value) Struct() Struct {
	return Struct{v.Type.Struct(), v.sur}
}

func (v Value) PtrToThis() Value {
	return Value{v.Type.PtrToThis, v.sur}
}

// v.Type.Kind must be KInterface
func (v Value) Surface() Interface {
	var (
		typ *Type // TargetType
		val unsafe.Pointer
	)
	ifacetype := v.Type.Surface()
	if v.Type.NumMethod() == 0 {
		eface := (*EmptyInterface)(v.val)
		if eface.Type != nil {
			typ = eface.Type
			val = unsafe.Pointer(eface.word)
		}
	} else {
		iface := (*NonEmptyInterface)(v.val)
		if iface.ITab == nil {
			return Interface{}
		}
		typ = iface.ITab.TargetType
		val = unsafe.Pointer(iface.word)
	}
	fl := v.flag & flagRO

	fl |= flag(typ.Kind()) << flagKindShift
	if typ != nil && typ.Size > ptrSize {
		fl |= flagIndir
	}
	return Interface{ifacetype, sur{val, fl, v.typ}, typ}
}

func (v Interface) InterfaceData() [2]uintptr {
	// We treat this as a read operation, so we allow
	// it even for unexported data, because the caller
	// has to import "unsafe" to turn it into something
	// that can be abused.
	// Interface value is always bigger than a word; assume flagIndir.
	return *(*[2]uintptr)(v.val)
}

func (v Ptr) Elem() Value {
	val := v.val
	if v.flag&flagIndir != 0 {
		val = *(*unsafe.Pointer)(val)
	}
	// The returned value's address is v's value.
	if val == nil {
		return Value{}
	}

	tt := (*PtrType)(unsafe.Pointer(v.Type))
	typ := tt.Elem
	fl := v.flag&flagRO | flagIndir | flagAddr
	fl |= flag(typ.Kind() << flagKindShift)

	return Value{typ, sur{val, fl, unsafe.Pointer(tt.Elem)}}
}

// Indirect returns the value that v points to.
// If v is a nil pointer, Indirect returns a zero Value.
// If v is not a pointer, Indirect returns v.
func (v Value) Indirect() Value {
	if v.Kind() != KPtr {
		return v
	}
	return v.Ptr().Elem()
}

// v.Type.Elem.Kind must be KUint8
func (v Slice) Bytes() []byte {
	v.Type.Elem.mustBe(KUint8)
	return *(*[]byte)(v.val)
}

// v.Type.Elem.Kind must be KInt32
func (v Slice) Runes() []rune {
	v.Type.Elem.mustBe(KInt32)
	return *(*[]rune)(v.val)
}

// v.Type.Elem.Kind must be KUint8
func (v Array) Bytes() []byte {
	v.Type.Elem.mustBe(KUint8)
	return *(*[]byte)(v.val)
}

// v.Type.Elem.Kind must be KInt32
func (v Array) Runes() []rune {
	v.Type.Elem.mustBe(KInt32)
	return *(*[]rune)(v.val)
}

func (v Map) Len() int {
	return maplen(v.IWord())
}
func (v Slice) Len() int {
	return (*SliceHeader)(v.val).Len
}
func (v Array) Len() int {
	return int(v.Type.len)
}
func (v Chan) Len() int {
	return chanlen(v.IWord())
}

func (v Slice) Cap() int {
	return (*SliceHeader)(v.val).Cap
}
func (v Array) Cap() int {
	return int(v.Type.len)
}
func (v Chan) Cap() int {
	return int(chancap(v.IWord()))
}

func (v Array) Index(i int) Value {
	tt := v.Type
	if i < 0 || i > v.Len() {
		panic("rtype: array index out of range")
		return Value{}
	}
	typ := tt.Elem
	fl := v.flag & (flagRO | flagIndir | flagAddr) // bits same as overall array
	fl |= flag(typ.Kind()) << flagKindShift
	offset := uintptr(i) * typ.Size

	var val unsafe.Pointer
	switch {
	case fl&flagIndir != 0:
		// Indirect.  Just bump pointer.
		val = unsafe.Pointer(uintptr(v.val) + offset)
	case bigEndian:
		// Direct.  Discard leading bytes.
		val = unsafe.Pointer(uintptr(v.val) << (offset * 8))
	default:
		// Direct.  Discard leading bytes.
		val = unsafe.Pointer(uintptr(v.val) >> (offset * 8))
	}
	return Value{typ, sur{val, fl, unsafe.Pointer(tt.Elem)}}

}
func (v Slice) Index(i int) Value {
	// Element flag same as Elem of Ptr.
	// Addressable, indirect, possibly read-only.
	fl := flagAddr | flagIndir | v.flag&flagRO
	s := (*SliceHeader)(v.val)
	if i < 0 || i >= s.Len {
		panic("surface: slice index out of range")
	}
	typ := v.Type.Elem
	fl |= flag(typ.Kind()) << flagKindShift
	val := unsafe.Pointer(s.Data + uintptr(i)*typ.Size)
	return Value{typ, sur{val, fl, unsafe.Pointer(v.Type.Elem)}}
}

func (v Struct) Field(i int) Value {
	if i < 0 || i >= v.Type.NumField() {
		panic("surface: Field index out of range")
	}
	return v.field(i)
}
func (v Struct) field(i int) Value {
	field := v.Type.Fields[i]
	typ := field.Type

	// Inherit permission bits from v.
	fl := v.flag & (flagRO | flagIndir | flagAddr)
	// Using an unexported field forces flagRO.
	if field.pkgPath != nil {
		fl |= flagRO
	}
	fl |= flag(typ.Kind()) << flagKindShift

	var val unsafe.Pointer
	switch {
	case fl&flagIndir != 0:
		// Indirect.  Just bump pointer.
		val = unsafe.Pointer(uintptr(v.val) + field.offset)
	case bigEndian:
		// Direct.  Discard leading bytes.
		val = unsafe.Pointer(uintptr(v.val) << (field.offset * 8))
	default:
		// Direct.  Discard leading bytes.
		val = unsafe.Pointer(uintptr(v.val) >> (field.offset * 8))
	}

	return Value{typ, sur{val, fl, unsafe.Pointer(field.Type)}}
}

// FieldByIndex returns the nested field corresponding to index.
// It panics if v's Kind is not struct.
func (v Struct) FieldByIndex(index []int) Value {
	if len(index) == 0 {
		return Value{}
	}
	rv := v.Field(index[0])
	for _, x := range index[1:] {
		rv = rv.Indirect().Struct().Field(x)
	}
	return rv
}

// FieldByName returns the struct field with the given name.
// It returns the zero Value if no field was found.
// It panics if v's Kind is not struct.
func (v Struct) FieldByName(name string) Value {
	for i, f := range v.Type.Fields {
		if f.Name() == name {
			return v.field(i)
		}
	}
	return Value{}
}

// ValueOf returns a new Value initialized to the concrete value
// stored in the interface i.  ValueOf(nil) returns the zero Value.
func ValueOf(i interface{}) Value {
	if i == nil {
		return Value{}
	}

	// TODO(rsc): Eliminate this terrible hack.
	// In the call to packValue, eface.typ doesn't escape,
	// and eface.word is an integer.  So it looks like
	// i (= eface) doesn't escape.  But really it does,
	// because eface.word is actually a pointer.
	// escapes(i)

	// For an interface value with the noAddr bit set,
	// the representation is identical to an empty interface.
	eface := *(*EmptyInterface)(unsafe.Pointer(&i))
	typ := eface.Type
	fl := flag(typ.Kind()) << flagKindShift
	if typ.Size > ptrSize {
		fl |= flagIndir
	}
	return Value{
		typ,
		sur{
			unsafe.Pointer(eface.word),
			fl,
			unsafe.Pointer(eface.Type),
		},
	}
}

// Dummy annotation marking that the value x escapes,
// for use in cases where the reflect code is so clever that
// the compiler cannot follow.
func escapes(x interface{}) {
	if dummy.b {
		dummy.x = x
	}
}

var dummy struct {
	b bool
	x interface{}
}

// A ValueError occurs when a Value method is invoked on
// a Value that does not support it.  Such cases are documented
// in the description of each method.
type ValueError struct {
	Method string
	Kind   Kind
}

func (e *ValueError) Error() string {
	if e.Kind == 0 {
		return "surface: call of " + e.Method + " on zero Value"
	}
	return "surface: call of " + e.Method + " on " + e.Kind.String() + " Value"
}
