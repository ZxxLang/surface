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
	"reflect"
	"runtime"
	"strconv"
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

type sur struct {
	// val holds the 1-word representation of the value.
	// If flag's flagIndir bit is set, then val is a pointer to the data.
	// Otherwise val is a word holding the actual data.
	// When the data is smaller than a word, it begins at
	// the first byte (in the memory address sense) of val.
	// We use unsafe.Pointer so that the garbage collector
	// knows that val could be a pointer.
	val unsafe.Pointer

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

type reflectValue struct {
	typ *uintptr //*rtype
	val unsafe.Pointer
	flag
}

func toValue(typ *uintptr, sur sur) reflect.Value {
	v := reflect.Value{}
	rv := (*reflectValue)(unsafe.Pointer(&sur.val))
	rv.typ = typ
	rv.val = sur.val
	rv.flag = sur.flag
	return v
}

func ValueFrom(v reflect.Value) Value {
	rv := (*reflectValue)(unsafe.Pointer(&v))
	return Value{
		(*Type)(unsafe.Pointer(rv.typ)),
		sur{rv.val, rv.flag},
	}
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

// methodReceiver returns information about the receiver
// described by v. The Value v may or may not have the
// flagMethod bit set, so the kind cached in v.flag should
// not be used.
func methodReceiver(op string, v Value, methodIndex int) (t *FuncType, fn unsafe.Pointer, rcvr IWord) {
	i := methodIndex
	if v.Type.Kind() == KInterface {
		tt := (*InterfaceType)(unsafe.Pointer(v.Type))
		if i < 0 || i >= len(tt.Methods) {
			panic("reflect: internal error: invalid method index")
		}
		m := &tt.Methods[i]
		if m.pkgPath != nil {
			panic("reflect: " + op + " of unexported method")
		}
		t = m.Type
		iface := (*nonEmptyInterface)(v.val)
		if iface.ITab == nil {
			panic("reflect: " + op + " of method on nil interface value")
		}
		fn = unsafe.Pointer(&iface.ITab.Fun[i])
		rcvr = iface.word
	} else {
		ut := v.Type.uncommonType
		if ut == nil || i < 0 || i >= len(ut.Methods) {
			panic("reflect: internal error: invalid method index")
		}
		m := &ut.Methods[i]
		if m.pkgPath != nil {
			panic("reflect: " + op + " of unexported method")
		}
		fn = unsafe.Pointer(&m.Call)
		t = m.MethodType
		rcvr = v.IWord()
	}
	return
}

// loadIword loads n bytes at p from memory into an iword.
func loadIword(p unsafe.Pointer, n uintptr) IWord {
	// Run the copy ourselves instead of calling memmove
	// to avoid moving w to the heap.
	var w IWord
	switch n {
	default:
		panic("reflect: internal error: loadIword of " + strconv.Itoa(int(n)) + "-byte value")
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

func valueIWord(f flag, size uintptr, val unsafe.Pointer) IWord {
	if f&flagIndir != 0 && size <= ptrSize {
		// Have indirect but want direct word.
		return loadIword(val, size)
	}
	return IWord(val)
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
		panic("reflect: " + methodName() + " using value obtained using unexported field")
	}
}

func (f flag) CanAddr() bool {
	return f&flagAddr != 0
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
		panic("reflect: " + methodName() + " using value obtained using unexported field")
	}
	if f&flagAddr == 0 {
		panic("reflect: " + methodName() + " using unaddressable value")
	}
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

//func (f flag) ReadOnly() bool {
//	return f&flagRO != 0
//}

func (f flag) CanSet() bool {
	return f&(flagAddr|flagRO) == flagAddr
}

func (v Value) IWord() IWord {
	return valueIWord(v.flag, v.Type.Size, v.val)
}
func (v Array) IWord() IWord {
	return valueIWord(v.flag, v.Type.Size, v.val)
}
func (v Chan) IWord() IWord {
	return valueIWord(v.flag, v.Type.Size, v.val)
}
func (v Func) IWord() IWord {
	return valueIWord(v.flag, v.Type.Size, v.val)
}
func (v Interface) IWord() IWord {
	return valueIWord(v.flag, v.Type.Size, v.val)
}
func (v Map) IWord() IWord {
	return valueIWord(v.flag, v.Type.Size, v.val)
}
func (v Ptr) IWord() IWord {
	return valueIWord(v.flag, v.Type.Size, v.val)
}
func (v Slice) IWord() IWord {
	return valueIWord(v.flag, v.Type.Size, v.val)
}
func (v Struct) IWord() IWord {
	return valueIWord(v.flag, v.Type.Size, v.val)
}

func (v Value) Reflect() reflect.Value {
	return toValue((*uintptr)(unsafe.Pointer(v.Type)), v.sur)
}
func (v Array) Reflect() reflect.Value {
	return toValue((*uintptr)(unsafe.Pointer(v.Type)), v.sur)
}
func (v Chan) Reflect() reflect.Value {
	return toValue((*uintptr)(unsafe.Pointer(v.Type)), v.sur)
}
func (v Func) Reflect() reflect.Value {
	return toValue((*uintptr)(unsafe.Pointer(v.Type)), v.sur)
}
func (v Interface) Reflect() reflect.Value {
	return toValue((*uintptr)(unsafe.Pointer(v.Type)), v.sur)
}
func (v Map) Reflect() reflect.Value {
	return toValue((*uintptr)(unsafe.Pointer(v.Type)), v.sur)
}
func (v Ptr) Reflect() reflect.Value {
	return toValue((*uintptr)(unsafe.Pointer(v.Type)), v.sur)
}
func (v Slice) Reflect() reflect.Value {
	return toValue((*uintptr)(unsafe.Pointer(v.Type)), v.sur)
}
func (v Struct) Reflect() reflect.Value {
	return toValue((*uintptr)(unsafe.Pointer(v.Type)), v.sur)
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
	return Value{v.Type.PtrToThis, sur{unsafe.Pointer(&v.val), v.flag}}
}

func (v Value) Interface() Interface {
	var (
		typ *Type // TargetType
		val unsafe.Pointer
	)
	if v.Type.NumMethod() == 0 {
		eface := (*emptyInterface)(v.val)
		if eface.Type != nil {
			typ = eface.Type
			val = unsafe.Pointer(eface.word)
		}
	} else {
		iface := (*nonEmptyInterface)(v.val)
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
	return Interface{v.Type.Interface(), sur{val, v.flag}, typ} /// ???
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
	return Value{typ, sur{val, fl}}
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

func (v Slice) Bytes() []byte {
	v.Type.Elem.mustBe(KUint8)
	return *(*[]byte)(v.val)
}

func (v Slice) Runes() []rune {
	v.Type.Elem.mustBe(KInt32)
	return *(*[]rune)(v.val)
}

func (v Array) Bytes() []byte {
	v.Type.Elem.mustBe(KUint8)
	return *(*[]byte)(v.val)
}
func (v Array) Runes() []rune {
	v.Type.Elem.mustBe(KInt32)
	return *(*[]rune)(v.val)
}

// ------------------------ to FooType -----------------------
// Len returns v's length.
// It panics if v's Kind is not Array, Chan, Map, Slice, or String.
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
	return Value{typ, sur{val, fl}}
}
func (v Slice) Index(i int) Value {
	// Element flag same as Elem of Ptr.
	// Addressable, indirect, possibly read-only.
	fl := flagAddr | flagIndir | v.flag&flagRO
	s := (*SliceHeader)(v.val)
	if i < 0 || i >= s.Len {
		panic("reflect: slice index out of range")
	}
	typ := v.Type.Elem
	fl |= flag(typ.Kind()) << flagKindShift
	val := unsafe.Pointer(s.Data + uintptr(i)*typ.Size)
	return Value{typ, sur{val, fl}}
}
func (v Map) Index(key Value) Value {
	return ValueFrom(v.Reflect().MapIndex(key.Reflect()))
}

func (v Map) Keys() []Value {
	rv := v.Reflect().MapKeys()
	ret := make([]Value, len(rv))
	for i := 0; i < len(rv); i++ {
		ret[i] = ValueFrom(rv[i])
	}
	return ret
}

func (v Struct) Field(i int) Value {
	if i < 0 || i >= v.Type.NumField() {
		panic("reflect: Field index out of range")
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

	return Value{typ, sur{val, fl}}
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
	eface := *(*emptyInterface)(unsafe.Pointer(&i))
	typ := eface.Type
	fl := flag(typ.Kind()) << flagKindShift
	if typ.Size > ptrSize {
		fl |= flagIndir
	}
	return Value{typ, sur{unsafe.Pointer(eface.word), fl}}
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
		return "reflect: call of " + e.Method + " on zero Value"
	}
	return "reflect: call of " + e.Method + " on " + e.Kind.String() + " Value"
}
