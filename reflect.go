// surface 需要依赖 reflect 的部分

package surface

import (
	"reflect"
	"unsafe"
)

func (k Kind) String() string {
	return reflect.Kind(k).String()
}

type reflectValue struct {
	typ *Type //*rtype
	val unsafe.Pointer
	flag
}

func FromValue(v reflect.Value) Value {
	rv := (*reflectValue)(unsafe.Pointer(&v))
	return Value{
		rv.typ,
		sur{
			rv.val,
			rv.flag,
			unsafe.Pointer(rv.typ),
		},
	}
}

// Interface returns v's current value as an interface{}.
// It is equivalent to:
//	var i interface{} = (v's underlying value)
// It panics if the Value was obtained by accessing
// unexported struct fields.
func (s sur) Interface() (i interface{}) {
	if !s.CanInterface() {
		panic(&ValueError{"surface.Value.CanInterface", s.Kind()})
	}
	return s.ToValue().Interface()
}

func (s sur) ToValue() reflect.Value {
	v := reflect.Value{}
	rv := (*reflectValue)(unsafe.Pointer(&v))
	rv.typ = (*Type)(unsafe.Pointer(s.typ))
	rv.val = s.val
	rv.flag = s.flag
	return v
}

func (v Map) Index(key Value) Value {
	return FromValue(v.ToValue().MapIndex(key.ToValue()))
}

func (v Map) Keys() []Value {
	rv := v.ToValue().MapKeys()
	ret := make([]Value, len(rv))
	for i := 0; i < len(rv); i++ {
		ret[i] = FromValue(rv[i])
	}
	return ret
}
