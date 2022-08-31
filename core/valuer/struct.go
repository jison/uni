package valuer

import (
	"fmt"
	"reflect"

	"github.com/jison/uni/internal/errors"
	"github.com/jison/uni/internal/reflecting"
)

type structValuer struct {
	structType reflect.Type
}

func (v *structValuer) Value(inputs []Value) Value {
	if v.structType == nil {
		return ErrorValue(errors.Newf("struct type is nil"))
	}

	fields := map[string]reflect.Value{}
	errs := errors.Empty()
	for _, inVal := range inputs {
		if err, ok := inVal.AsError(); ok {
			errs = errs.AddErrors(err)
			continue
		}
		rVal, isSingle := inVal.AsSingle()
		if !isSingle {
			errs = errs.AddErrors(errors.Bugf("input of struct should be structField"))
			continue
		}

		if !rVal.CanInterface() {
			errs = errs.AddErrors(errors.Newf("%+v .CanInterface() is false", inVal))
			continue
		}

		if field, ok := rVal.Interface().(structField); ok {
			if _, exist := fields[field.name]; exist {
				errs = errs.AddErrors(errors.Bugf("duplicate field %v of struct element", field.name))
				continue
			}

			fields[field.name] = field.val
		} else {
			errs = errs.AddErrors(errors.Bugf("input of struct should be structField"))
		}
	}
	if errs.HasError() {
		return ErrorValue(errs)
	}

	structVal, err := reflecting.InitStructWithReflectValues(v.structType, fields)
	if err != nil {
		return ErrorValue(err)
	}
	return SingleValue(structVal)
}

func (v *structValuer) String() string {
	return fmt.Sprintf("Struct: %+v", v.structType)
}

func (v *structValuer) Clone() Valuer {
	return &structValuer{v.structType}
}

func (v *structValuer) Equal(other interface{}) bool {
	o, ok := other.(*structValuer)
	if !ok {
		return false
	}

	if v == nil || o == nil {
		return v == nil && o == nil
	}

	return v.structType == o.structType
}

func Struct(structType reflect.Type) Valuer {
	return &structValuer{structType}
}
