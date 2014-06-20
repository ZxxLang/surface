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
	"unsafe"
)

type IWord unsafe.Pointer

// emptyInterface is the header for an interface{} value.
type emptyInterface struct {
	Type *Type
	word IWord
}

// nonEmptyInterface is the header for a interface value with methods.
type nonEmptyInterface struct {
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
