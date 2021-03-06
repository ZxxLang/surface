// Copyright 2014 The ZxxLang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package surface

import (
	"fmt"
	"github.com/achun/testing-want"
	"go/ast"
	"reflect"
	"testing"
)

func TestType(t *testing.T) {
	wt := want.T(t)
	for i, s := range testType {
		sv := ValueOf(s)
		wt.NotNil(sv.Type)
		want.Printf("0x%x,\n", sv.Type.Hash)
		wt.True(sv.Type.Hash == testTypeHash[i])
		rv := reflect.ValueOf(s)
		rt := rv.Type()
		st := sv.Type
		wt.True(FromValue(rv).Type.Hash == testTypeHash[i])
		wt.Equal(rt.Name(), st.Name())
		wt.Equal(rt.String(), st.String())
		wt.Equal(rt.PkgPath(), st.PkgPath())
	}
}

func TestStructAndFields(t *testing.T) {
	wt := want.T(t)
	for i, s := range testStruct {
		sv := ValueOf(s)
		wt.NotNil(sv.Type)
		st := sv.Type.Struct()

		rv := reflect.ValueOf(s)
		rt := rv.Type()

		//want.Printf("0x%x,\n", sv.Type.Hash)
		wt.True(sv.Type.Hash == testStructHash[i])
		wt.True(FromValue(rv).Type.Hash == testStructHash[i])
		wt.Equal(rt.Name(), sv.Type.Name())
		wt.Equal(rt.String(), sv.Type.String())
		wt.Equal(rt.PkgPath(), sv.Type.PkgPath())

		wt.Equal(rt.NumField(), st.NumField(), i, s)
		wt.Equal(rv.Interface(), sv.Interface())

		wt.Equal(rv.CanSet(), sv.CanSet())
		wt.Equal(rv.CanAddr(), sv.CanAddr())
		wt.Equal(rv.CanInterface(), sv.CanInterface())

		v := sv.Struct()
		for i := 0; i < st.NumField(); i++ {
			rf := rt.Field(i)
			sf := st.Fields[i]
			wt.Equal(rf.Name, sf.Name(), i, " ", rt.String(), " ", sv.Type.String())
			wt.Equal(rf.PkgPath, sf.PkgPath())
			wt.Equal(rf.Type.Kind().String(), sf.Type.Kind().String())

			rv := rv.Field(i)
			sv := v.Field(i)
			wt.Equal(rv.CanSet(), sv.CanSet())
			wt.Equal(rv.CanAddr(), sv.CanAddr())
			wt.Equal(rv.CanInterface(), sv.CanInterface())

			if !rv.CanInterface() {
				continue
				t := rv.Type()
				want.Printf("!CanInterface() \n\tPkgPath:%q\n\t   Name:%q\n\t String:%q\n",
					t.PkgPath(), t.Name(), t.String())
			}
			wt.Equal(rv.Interface(), sv.Interface(),
				"\nsv.IsIndir():    ", sv.IsIndir(),
				"\nrt.String():             ", rt.String(),
				"\nrf.Name:                 ", rf.Name,
				"\nrf.PkgPath:              ", rf.PkgPath,
				"\nrf.Type.Kind().String(): ", rf.Type.Kind().String(),
			)
		}
	}
}

func TestBuiltinType(t *testing.T) {
	wt := want.T(t)
	for i, s := range testBuiltinType {
		sv := ValueOf(s)
		wt.NotNil(sv.Type)
		//want.Printf("0x%x,\n", sv.Type.Hash)
		wt.True(sv.Type.Hash == testBuiltinHash[i])
		rv := reflect.ValueOf(s)
		rt := rv.Type()
		st := sv.Type
		wt.True(FromValue(rv).Type.Hash == testBuiltinHash[i])
		wt.Equal(rt.Name(), st.Name())
		wt.Equal(rt.String(), st.String())
		wt.Equal(rt.PkgPath(), st.PkgPath())

		wt.Equal(rv.CanSet(), sv.CanSet())
		wt.Equal(rv.CanAddr(), sv.CanAddr())
		wt.Equal(rv.CanInterface(), sv.CanInterface())
	}
}

var testType = [...]interface{}{
	EmptyInterface{},
	Type{},
	&EmptyInterface{},
	&Type{},
}

var testTypeHash = [...]uint32{
	0x71774d96, 0x790a307b, 0x4a00045b, 0x17531429}

var testBuiltinType = [...]interface{}{
	true, false,
	uint8(0),
	uint16(0),
	uint32(0),
	uint64(0),
	int8(0),
	int16(0),
	int32(0),
	int64(0),
	float32(0),
	float64(0),
	complex64(0),
	complex128(0),
	"",
	' ',
	rune(0),
}
var testBuiltinHash = [...]uint32{
	0x13ff06c5,
	0x13ff06c5,
	0x663e425f,
	0xeff20ea0,
	0xd04ae83d,
	0x86318d2e,
	0xcc06c027,
	0xecd580ce,
	0xbbad4102,
	0x963f9bff,
	0xb0c23ed3,
	0x2ea27ffb,
	0x7925028c,
	0xb31a546d,
	0xe0ff5cb4,
	0xbbad4102,
	0xbbad4102,
}
var testStruct = [...]interface{}{
	ast.Comment{},
	ast.CommentGroup{},
	ast.Field{},
	ast.FieldList{},
	ast.BadExpr{},
	ast.Ident{},
	ast.Ellipsis{},
	ast.BasicLit{},
	ast.FuncLit{},
	ast.CompositeLit{},
	ast.ParenExpr{},
	ast.SelectorExpr{},
	ast.IndexExpr{},
	ast.SliceExpr{},
	ast.TypeAssertExpr{},
	ast.CallExpr{},
	ast.StarExpr{},
	ast.UnaryExpr{},
	ast.BinaryExpr{},
	ast.KeyValueExpr{},
	ast.ArrayType{},
	ast.StructType{},
	ast.FuncType{},
	ast.InterfaceType{},
	ast.MapType{},
	ast.ChanType{},
	ast.BadStmt{},
	ast.DeclStmt{},
	ast.EmptyStmt{},
	ast.LabeledStmt{},
	ast.ExprStmt{},
	ast.SendStmt{},
	ast.IncDecStmt{},
	ast.AssignStmt{},
	ast.GoStmt{},
	ast.DeferStmt{},
	ast.ReturnStmt{},
	ast.BranchStmt{},
	ast.BlockStmt{},
	ast.IfStmt{},
	ast.CaseClause{},
	ast.SwitchStmt{},
	ast.TypeSwitchStmt{},
	ast.CommClause{},
	ast.SelectStmt{},
	ast.ForStmt{},
	ast.RangeStmt{},
	ast.ImportSpec{},
	ast.ValueSpec{},
	ast.TypeSpec{},
	ast.BadDecl{},
	ast.GenDecl{},
	ast.FuncDecl{},
	ast.File{},
	ast.Package{},
	ast.Scope{},
	ast.Object{},

	reflect.Method{},
	reflect.StructField{},
	reflect.Value{},
	reflect.ValueError{},
	reflect.StringHeader{},
	reflect.SliceHeader{},
	reflect.SelectCase{},
}
var testStructHash = [...]uint32{
	0xf48243e5,
	0x9b7f98fd,
	0xbb565db3,
	0x70d6b046,
	0xaa2c6af9,
	0xc37b242a,
	0x871de7b6,
	0x788b934a,
	0x5be4e626,
	0x8f9625de,
	0x7e040fe8,
	0xa7f68b74,
	0x263d6039,
	0x39e3f92b,
	0x50985bd1,
	0x755ac548,
	0x89e1f214,
	0xa704b411,
	0xaf14e630,
	0xd8970a4d,
	0x77d5f5b,
	0x82528451,
	0x23afcb71,
	0x22edc278,
	0x3c88d9b7,
	0xd2399adc,
	0xa0f371c7,
	0x77c5c71d,
	0xcfe0f5d5,
	0x67e0a4b1,
	0x10bfd036,
	0xc032340d,
	0x7d77a37b,
	0xb654cffe,
	0xd36409d5,
	0x57e55264,
	0x57e74f2,
	0x49a0e155,
	0x22711cf1,
	0xaa17a95a,
	0x7f44a9b1,
	0x913d0d61,
	0x9650428a,
	0x91cdd9cc,
	0x9091d65e,
	0xf1df114a,
	0xb3db98a6,
	0x841d816b,
	0x2ac98620,
	0xda9e7a05,
	0x62432c5c,
	0x8171bbf2,
	0x28a17a52,
	0x7cc853e3,
	0xb21305e7,
	0xb3fba30f,
	0x86511633,
	0xe6850a2a,
	0x2474018,
	0x500c1abc,
	0x91d8c392,
	0xe6f3830b,
	0xfde08b8a,
	0x914cd10,
}

func TestFuncType(t *testing.T) {
	wt := want.T(t)
	for i, s := range testFuncType {
		sv := ValueOf(s)
		wt.NotNil(sv.Type)
		//want.Printf("0x%x,\n", sv.Type.Hash)
		wt.True(sv.Type.Hash == testFuncHash[i])
		rv := reflect.ValueOf(s)
		rt := rv.Type()
		st := sv.Type
		wt.True(FromValue(rv).Type.Hash == testFuncHash[i])
		wt.Equal(rt.Name(), st.Name())
		wt.Equal(rt.String(), st.String())
		wt.Equal(rt.PkgPath(), st.PkgPath())

		wt.Equal(rv.CanSet(), sv.CanSet())
		wt.Equal(rv.CanAddr(), sv.CanAddr())
		wt.Equal(rv.CanInterface(), sv.CanInterface())
	}
}

var testFuncType = [...]interface{}{
	ValueOf,
	fmt.Print,
	fmt.Println,
	fmt.Printf,
	fmt.Scanf,
}
var testFuncHash = [...]uint32{
	0x26955e73,
	0x81799c9a,
	0x81799c9a,
	0xd9fb8597,
	0xd9fb8597,
}

func TestInterfaceType(t *testing.T) {
	wt := want.T(t)
	for i, s := range testInterfaceType {
		sv := ValueOf(s)
		wt.NotNil(sv.Type)
		//want.Printf("0x%x,\n", sv.Type.Hash)
		wt.True(sv.Type.Hash == testInterfaceHash[i])
		rv := reflect.ValueOf(s)
		rt := rv.Type()
		st := sv.Type
		wt.True(FromValue(rv).Type.Hash == testInterfaceHash[i])
		wt.Equal(rt.Name(), st.Name())
		wt.Equal(rt.String(), st.String())
		wt.Equal(rt.PkgPath(), st.PkgPath())
		wt.Equal(rv.CanSet(), sv.CanSet())
		wt.Equal(rv.CanAddr(), sv.CanAddr())
		wt.Equal(rv.CanInterface(), sv.CanInterface())
	}
}

type stringer interface {
	String() string
}

var testInterfaceType = [...]interface{}{
	stringer(&ast.Ident{}),
	fmt.Stringer(&ast.Ident{}),
	stringer(&reflect.Value{}),
	fmt.Stringer(&reflect.Value{}),
}

var testInterfaceHash = [...]uint32{
	0x2f3b734e,
	0x2f3b734e,
	0xf764ad0,
	0xf764ad0,
}

func TestMethodType(t *testing.T) {
	ai := ast.Ident{}
	rv := reflect.Value{}
	pai := &ai
	prv := &rv
	var testMethodType = [...]interface{}{
		ai.String,
		pai.String,
		rv.String,
		prv.String,
	}
	var testMethodHash = [...]uint32{
		0x1ecb6da2,
		0x1ecb6da2,
		0x1ecb6da2,
		0x1ecb6da2,
	}

	wt := want.T(t)
	for i, s := range testMethodType {
		sv := ValueOf(s)
		wt.NotNil(sv.Type)
		//want.Printf("0x%x,\n", sv.Type.Hash)
		wt.True(sv.Type.Hash == testMethodHash[i])
		rv := reflect.ValueOf(s)
		rt := rv.Type()
		st := sv.Type
		wt.True(FromValue(rv).Type.Hash == testMethodHash[i])
		wt.Equal(rt.Name(), st.Name())
		wt.Equal(rt.String(), st.String())
		wt.Equal(rt.PkgPath(), st.PkgPath())
		wt.Equal(rv.CanSet(), sv.CanSet())
		wt.Equal(rv.CanAddr(), sv.CanAddr())
		wt.Equal(rv.CanInterface(), sv.CanInterface())
	}
}

func TestMap(t *testing.T) {
	wt := want.T(t)

	m := map[string]interface{}{
		"1":    1,
		"2":    2,
		"true": true,
		"nil":  nil,
		"func": TestMap,
	}
	sv := ValueOf(m).Map()
	rv := reflect.ValueOf(m)
	wt.Equal(rv.Len(), sv.Len())
	wt.Equal(FromValue(rv).Map().Len(), sv.Len())
	for _, key := range []string{"1", "2", "true", "nil", "func"} {
		rv := rv.MapIndex(reflect.ValueOf(key))
		sv := sv.Index(ValueOf(key))
		wt.Equal(rv.CanSet(), sv.CanSet())
		wt.Equal(rv.CanAddr(), sv.CanAddr())
		wt.Equal(rv.CanInterface(), sv.CanInterface())
		wt.Equal(rv.Interface(), sv.Interface())
	}
}
