// Derived from Go's package reflect
// --------------------------------------------------------------------------
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Copyright 2014 The ZxxLang Authors. All rights reserved.

package surface

import (
	"reflect"
	"strconv"
	"unsafe"
)

type Kind uint

const (
	KInvalid Kind = iota
	KBool
	KInt
	KInt8
	KInt16
	KInt32
	KInt64
	KUint
	KUint8
	KUint16
	KUint32
	KUint64
	KUintptr
	KFloat32
	KFloat64
	KComplex64
	KComplex128
	KArray
	KChan
	KFunc
	KInterface
	KMap
	KPtr
	KSlice
	KString
	KStruct
	KUnsafePointer
)

func (k Kind) String() string {
	return reflect.Kind(k).String()
}

const (
	kindMask       Kind = 0x7f
	kindNoPointers Kind = 0x80
)

// Type is the common implementation of most values.
// It is embedded in other, public struct types, but always
// with a unique tag like `reflect:"array"` or `reflect:"ptr"`
// so that code cannot convert from, say, *arrayType to *ptrType.
type Type struct {
	Size          uintptr        // size in bytes
	Hash          uint32         // hash of type; avoids computation in hash tables
	_             uint8          // unused/padding
	Align         uint8          // alignment of variable with this type
	FieldAlign    uint8          // alignment of struct field with this type
	kind          uint8          // enumeration for C
	alg           *uintptr       // algorithm table (../runtime/runtime.h:/Alg)
	gc            unsafe.Pointer // garbage collection data
	string        *string        // string form; unnecessary but undeniably useful
	*uncommonType                // (relatively) uncommon fields
	PtrToThis     *Type          // type for pointer to this type, if used in binary or has methods
}

// uncommonType is present only for types with names or methods
// (if T is a named type, the uncommonTypes for T and *T have methods).
// Using a pointer to this struct reduces the overall size required
// to describe an unnamed type with no methods.
type uncommonType struct {
	name    *string  // name of type
	pkgPath *string  // import path; nil for built-in types like int, string
	Methods []Method // methods associated with type
}

// ChanDir represents a channel type's direction.
type ChanDir int

const (
	RecvDir ChanDir             = 1 << iota // <-chan
	SendDir                                 // chan<-
	BothDir = RecvDir | SendDir             // chan
)

// arrayType represents a fixed array type.
type ArrayType struct {
	Type  `reflect:"array"`
	Elem  *Type // array element type
	Slice *Type // slice type
	len   uintptr
}

// chanType represents a channel type.
type ChanType struct {
	Type `reflect:"chan"`
	Elem *Type   // channel element type
	Dir  uintptr // channel direction (ChanDir)
}

// funcType represents a function type.
type FuncType struct {
	Type      `reflect:"func"`
	DotDotDot bool    // last input parameter is ...
	In        []*Type // input parameter types
	Out       []*Type // output parameter types
}

// Method on non-interface type
type Method struct {
	name       *string        // name of method
	pkgPath    *string        // nil for exported Names; otherwise import path
	MethodType *FuncType      // method type (without receiver)
	FuncType   *FuncType      // .(*FuncType) underneath (with receiver)
	IfaceCall  unsafe.Pointer // fn used in interface call (one-word receiver)
	Call       unsafe.Pointer // fn used for normal method call
}

// imethod represents a method on an interface type
type IMethod struct {
	name    *string   // name of method
	pkgPath *string   // nil for exported Names; otherwise import path
	Type    *FuncType // .(*FuncType) underneath
}

// interfaceType represents an interface type.
type InterfaceType struct {
	Type    `reflect:"interface"`
	Methods []IMethod // sorted by hash
}

// mapType represents a map type.
type MapType struct {
	Type   `reflect:"map"`
	Key    *Type // map key type
	Elem   *Type // map element (value) type
	Bucket *Type // internal bucket structure
	HMap   *Type // internal map header
}

// ptrType represents a pointer type.
type PtrType struct {
	Type `reflect:"ptr"`
	Elem *Type // pointer element (pointed at) type
}

// StringHeader is the runtime representation of a string.
// It cannot be used safely or portably and its representation may
// change in a later release.
// Moreover, the Data field is not sufficient to guarantee the data
// it references will not be garbage collected, so programs must keep
// a separate, correctly typed pointer to the underlying data.
type StringHeader struct {
	Data uintptr
	Len  int
}

// SliceHeader is the runtime representation of a slice.
// It cannot be used safely or portably and its representation may
// change in a later release.
// Moreover, the Data field is not sufficient to guarantee the data
// it references will not be garbage collected, so programs must keep
// a separate, correctly typed pointer to the underlying data.
type SliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}

// sliceType represents a slice type.
type SliceType struct {
	Type `reflect:"slice"`
	Elem *Type // slice element type
}

// structType represents a struct type.
type StructType struct {
	Type   `reflect:"struct"`
	Fields []StructField // sorted by offset
}

// Struct field
type StructField struct {
	name    *string // nil for embedded fields
	pkgPath *string // nil for exported Names; otherwise import path
	Type    *Type   // type of field
	tag     *string // nil if no tag
	offset  uintptr // byte offset of field within struct
}

// A StructTag is the tag string in a struct field.
//
// By convention, tag strings are a concatenation of
// optionally space-separated key:"value" pairs.
// Each key is a non-empty string consisting of non-control
// characters other than space (U+0020 ' '), quote (U+0022 '"'),
// and colon (U+003A ':').  Each value is quoted using U+0022 '"'
// characters and Go string literal syntax.
type StructTag string

// Get returns the value associated with key in the tag string.
// If there is no such key in the tag, Get returns the empty string.
// If the tag does not have the conventional format, the value
// returned by Get is unspecified.
func (tag StructTag) Get(key string) string {
	for tag != "" {
		// skip leading space
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// scan to colon.
		// a space or a quote is a syntax error
		i = 0
		for i < len(tag) && tag[i] != ' ' && tag[i] != ':' && tag[i] != '"' {
			i++
		}
		if i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// scan quoted string to find value
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		if key == name {
			value, _ := strconv.Unquote(qvalue)
			return value
		}
	}
	return ""
}

func (t *Type) IsNil() bool {
	return t == nil
}

func (t *Type) String() string {
	if t == nil {
		return "nil"
	}
	return *t.string
}

func (u *uncommonType) IsBuiltin() bool {
	return u == nil || u.pkgPath == nil
}
func (u *uncommonType) Name() string {
	if u == nil {
		return ""
	}
	return *u.name
}

func (u *uncommonType) PkgPath() string {
	if u == nil {
		return ""
	}
	return *u.pkgPath
}
func (u *uncommonType) NumMethod() int {
	if u == nil {
		return 0
	}
	return len(u.Methods)
}

func (u *InterfaceType) NumMethod() int {
	if u == nil {
		return 0
	}
	return len(u.Methods)
}

func (u Method) Name() string {
	return *u.name
}

func (u Method) PkgPath() string {
	if u.pkgPath == nil {
		return ""
	}
	return *u.pkgPath
}
func (u Method) Exported() bool {
	return u.pkgPath == nil
}

func (u IMethod) Name() string {
	return *u.name
}
func (u IMethod) PkgPath() string {
	if u.pkgPath == nil {
		return ""
	}
	return *u.pkgPath
}
func (u IMethod) Exported() bool {
	return u.pkgPath == nil
}

func (u StructField) Name() string {
	if u.name == nil {
		return ""
	}
	return *u.name
}

func (u StructField) PkgPath() string {
	if u.pkgPath == nil {
		return ""
	}
	return *u.pkgPath
}

func (u StructField) Tag() StructTag {
	if u.tag == nil {
		return ""
	}
	return StructTag(*u.tag)
}
func (u StructField) Embedded() bool {
	return u.name == nil
}
func (u StructField) Exported() bool {
	return u.pkgPath == nil
}
func (u StructField) HasTag() bool {
	return u.tag != nil
}

func TypeOf(i interface{}) *Type {
	ei := *(*emptyInterface)(unsafe.Pointer(&i))
	return ei.Type
}

func (t *Type) Kind() Kind {
	if t == nil {
		return KInvalid
	}
	return Kind(t.kind) & kindMask
}

func (t *Type) mustBe(expected Kind) {
	if t.Kind() != expected {
		panic(&ValueError{methodName(), t.Kind()})
	}
}

// Array Chan Func Interface Map Ptr Slice Struct
func (t *Type) Array() *ArrayType {
	t.mustBe(KArray)
	return (*ArrayType)(unsafe.Pointer(t))
}

func (t *Type) Chan() *ChanType {
	t.mustBe(KChan)
	return (*ChanType)(unsafe.Pointer(t))
}

func (t *Type) Func() *FuncType {
	t.mustBe(KFunc)
	return (*FuncType)(unsafe.Pointer(t))
}

func (t *Type) Interface() *InterfaceType {
	if t.Kind() == KPtr {
		t = t.Ptr().Elem
	}
	t.mustBe(KInterface)
	return (*InterfaceType)(unsafe.Pointer(t))
}

func (t *Type) Map() *MapType {
	t.mustBe(KMap)
	return (*MapType)(unsafe.Pointer(t))
}

func (t *Type) Ptr() *PtrType {
	t.mustBe(KPtr)
	return (*PtrType)(unsafe.Pointer(t))
}

func (t *Type) Slice() *SliceType {
	t.mustBe(KSlice)
	return (*SliceType)(unsafe.Pointer(t))
}

func (t *Type) Struct() *StructType {
	t.mustBe(KStruct)
	return (*StructType)(unsafe.Pointer(t))
}

func (t *StructType) NumField() int {
	return len(t.Fields)
}

func (v *Type) Implements(t *InterfaceType) bool {
	if v == nil || t == nil {
		return false
	}
	if t.NumMethod() == 0 {
		return true
	}
	if v.NumMethod() == 0 {
		return false
	}
	i := 0
	if v.Kind() == KInterface {
		for j := 0; j < len(v.Methods); j++ {
			tm := &t.Methods[i]
			vm := &v.Methods[j]
			if vm.name == tm.name && vm.pkgPath == tm.pkgPath && vm.FuncType == tm.Type {
				if i++; i >= len(t.Methods) {
					return true
				}
			}
		}
		return false
	}
	for j := 0; j < len(v.Methods); j++ {
		tm := &t.Methods[i]
		vm := &v.Methods[j]
		if vm.name == tm.name && vm.pkgPath == tm.pkgPath && vm.MethodType == tm.Type {
			if i++; i >= len(t.Methods) {
				return true
			}
		}
	}
	return false
}
func (v *Type) Indirect() *Type {
	if v.Kind() == KPtr {
		return v.Ptr().Elem
	}
	return v
}
