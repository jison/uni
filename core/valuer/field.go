package valuer

import (
	"fmt"
	"reflect"

	"github.com/jison/uni/internal/errors"
)

type structFieldValuer struct {
	fieldName string
}

type structField struct {
	name string
	val  reflect.Value
}

func (v *structFieldValuer) ValueOne(input Value) Value {
	// input of ValueOne can not be an Error value
	//if _, ok := input.AsError(); ok {
	//	return input
	//}

	rVal, ok := input.AsSingle()
	if !ok {
		return ErrorValue(errors.Bugf("input value of field should be single"))
	}

	elem := structField{
		name: v.fieldName,
		val:  rVal,
	}

	return SingleValue(reflect.ValueOf(elem))
}

func (v *structFieldValuer) String() string {
	return fmt.Sprintf("Field: %v", v.fieldName)
}

func (v *structFieldValuer) Clone() OneInputValuer {
	return &structFieldValuer{v.fieldName}
}

func (v *structFieldValuer) Equal(other interface{}) bool {
	o, ok := other.(*structFieldValuer)
	if !ok {
		return false
	}

	if v == nil || o == nil {
		return v == nil && o == nil
	}

	return v.fieldName == o.fieldName
}

func Field(fieldName string) Valuer {
	return &oneInputValuer{&structFieldValuer{fieldName}}
}
