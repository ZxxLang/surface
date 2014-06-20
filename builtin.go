package surface

import (
	"unsafe"
)

func (v Value) Bool() bool {
	v.mustBe(KBool)
	if v.flag&flagIndir != 0 {
		return *(*bool)(v.val)
	}
	return *(*bool)(unsafe.Pointer(&v.val))
}

func (v Value) Int() int {
	v.mustBe(KInt)
	if v.flag&flagIndir != 0 {
		return *(*int)(v.val)
	}
	return *(*int)(unsafe.Pointer(&v.val))
}

func (v Value) Uint() uint {
	v.mustBe(KUint)
	if v.flag&flagIndir != 0 {
		return *(*uint)(v.val)
	}
	return *(*uint)(unsafe.Pointer(&v.val))
}

func (v Value) Int8() int8 {
	v.mustBe(KInt8)
	if v.flag&flagIndir != 0 {
		return *(*int8)(v.val)
	}
	return *(*int8)(unsafe.Pointer(&v.val))
}
func (v Value) Int16() int16 {
	v.mustBe(KInt16)
	if v.flag&flagIndir != 0 {
		return *(*int16)(v.val)
	}
	return *(*int16)(unsafe.Pointer(&v.val))
}
func (v Value) Int32() int32 {
	v.mustBe(KInt32)
	if v.flag&flagIndir != 0 {
		return *(*int32)(v.val)
	}
	return *(*int32)(unsafe.Pointer(&v.val))
}

func (v Value) Int64() int64 {
	k := v.Kind()
	var p unsafe.Pointer
	if v.flag&flagIndir != 0 {
		p = v.val
	} else {
		// The escape analysis is good enough that &v.val
		// does not trigger a heap allocation.
		p = unsafe.Pointer(&v.val)
	}
	switch k {
	case KInt:
		return int64(*(*int)(p))
	case KInt8:
		return int64(*(*int8)(p))
	case KInt16:
		return int64(*(*int16)(p))
	case KInt32:
		return int64(*(*int32)(p))
	case KInt64:
		return int64(*(*int64)(p))
	}
	panic(&ValueError{"rtype.Value.Int64", k})
}

func (v Value) Uint64() uint64 {
	k := v.Kind()
	var p unsafe.Pointer
	if v.flag&flagIndir != 0 {
		p = v.val
	} else {
		// The escape analysis is good enough that &v.val
		// does not trigger a heap allocation.
		p = unsafe.Pointer(&v.val)
	}
	switch k {
	case KUint:
		return uint64(*(*uint)(p))
	case KUint8:
		return uint64(*(*uint8)(p))
	case KUint16:
		return uint64(*(*uint16)(p))
	case KUint32:
		return uint64(*(*uint32)(p))
	case KUint64:
		return uint64(*(*uint64)(p))
	case KUintptr:
		return uint64(*(*uintptr)(p))
	}
	panic(&ValueError{"rtype.Value.Uint64", k})
}

func (v Value) Uint8() uint8 {
	v.mustBe(KUint8)
	if v.flag&flagIndir != 0 {
		return *(*uint8)(v.val)
	}
	return *(*uint8)(unsafe.Pointer(&v.val))
}

func (v Value) Uint16() uint16 {
	v.mustBe(KUint16)
	if v.flag&flagIndir != 0 {
		return *(*uint16)(v.val)
	}
	return *(*uint16)(unsafe.Pointer(&v.val))
}
func (v Value) Uint32() uint32 {
	v.mustBe(KUint32)
	if v.flag&flagIndir != 0 {
		return *(*uint32)(v.val)
	}
	return *(*uint32)(unsafe.Pointer(&v.val))
}

func (v Value) Uintptr() uintptr {
	v.mustBe(KUintptr)
	return *(*uintptr)(v.val)
}

func (v Value) Float32() float32 {
	v.mustBe(KFloat32)
	if v.flag&flagIndir != 0 {
		return *(*float32)(v.val)
	}
	return *(*float32)(unsafe.Pointer(&v.val))
}
func (v Value) Float64() float64 {
	k := v.Kind()
	switch k {
	case KFloat32:
		if v.flag&flagIndir != 0 {
			return float64(*(*float32)(v.val))
		}
		return float64(*(*float32)(unsafe.Pointer(&v.val)))
	case KFloat64:
		if v.flag&flagIndir != 0 {
			return *(*float64)(v.val)
		}
		return *(*float64)(unsafe.Pointer(&v.val))
	}
	panic(&ValueError{"rtype.Value.Float64", k})
}
func (v Value) Complex64() complex64 {
	v.mustBe(KComplex64)
	if v.flag&flagIndir != 0 {
		return *(*complex64)(v.val)
	}
	return *(*complex64)(unsafe.Pointer(&v.val))
}
func (v Value) Complex128() complex128 {
	v.mustBe(KComplex128)
	if v.flag&flagIndir != 0 {
		return *(*complex128)(v.val)
	}
	return *(*complex128)(unsafe.Pointer(&v.val))
}

func (v Value) String() string {
	v.mustBe(KString)
	if v.flag&flagIndir != 0 {
		return *(*string)(v.val)
	}
	return *(*string)(unsafe.Pointer(&v.val))
}

func (v Value) StringHeader() StringHeader {
	v.mustBe(KString)
	if v.flag&flagIndir != 0 {
		return *(*StringHeader)(unsafe.Pointer(v.val))
	}
	return *(*StringHeader)(unsafe.Pointer(&v.val))
}

func (v Value) Bytes() []byte {
	v.mustBe(KString)
	return *(*[]byte)(v.val)
}
