package valuer

import (
	"fmt"
	"reflect"
)

type constValuer struct {
	val reflect.Value
}

func (v *constValuer) Value(_ []Value) Value {
	return SingleValue(v.val)
}

func (v *constValuer) String() string {
	return fmt.Sprintf("Const: %+v(%+v)", v.val.Type(), v.val)
}

func (v *constValuer) Clone() Valuer {
	return &constValuer{v.val}
}

func (v *constValuer) Equal(other interface{}) bool {
	o, ok := other.(*constValuer)
	if !ok {
		return false
	}

	if v == nil || o == nil {
		return v == nil && o == nil
	}

	return v.val == o.val
}

func Const(val reflect.Value) Valuer {
	return &constValuer{val}
}
