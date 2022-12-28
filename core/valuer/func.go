package valuer

import (
	"fmt"
	"reflect"

	"github.com/jison/uni/internal/errors"
	"github.com/jison/uni/internal/reflecting"
)

type funcValuer struct {
	funcVal reflect.Value
}

func (v *funcValuer) Value(inputs []Value) Value {
	if !v.funcVal.IsValid() || v.funcVal.Kind() != reflect.Func {
		return ErrorValue(errors.Bugf("%v is not a function", v.funcVal))
	}

	params, err := v.params(inputs)
	if err != nil {
		return ErrorValue(err)
	}

	callValues, callErr := reflecting.ReflectFuncOfFunc(v.funcVal)(params)
	if callErr != nil {
		return ErrorValue(callErr)
	}

	//if len(callValues) == 0 {
	//	return ErrorValue(errors.Bugf("function %v has not return value", v.funcVal.Type()))
	//}

	return ArrayValue(callValues)
}

func (v *funcValuer) params(inputs []Value) ([]reflect.Value, error) {
	funcType := v.funcVal.Type()
	params := make([]reflect.Value, funcType.NumIn())

	errs := errors.Empty()

	for _, inVal := range inputs {
		if err, ok := inVal.AsError(); ok {
			errs = errs.AddErrors(err)
			continue
		}

		rVal, isSingle := inVal.AsSingle()
		if !isSingle {
			errs = errs.AddErrors(errors.Bugf("input of func should be funcParam"))
			continue
		}

		if !rVal.CanInterface() {
			errs = errs.AddErrorf("%+v .CanInterface() is false", inVal)
			continue
		}

		param, ok := rVal.Interface().(funcParam)
		if !ok {
			errs = errs.AddErrors(errors.Bugf("input of func should be funcParam"))
			continue
		}

		if param.index < 0 || param.index >= len(params) {
			err := errors.Bugf("index %v is out of range. [0, %v]", param.index, len(params)-1)
			errs = errs.AddErrors(err)
			continue
		}
		if params[param.index].IsValid() {
			errs = errs.AddErrors(errors.Bugf("duplicate index %v of param element", param.index))
			continue
		}

		if !param.val.Type().AssignableTo(funcType.In(param.index)) {
			err := errors.Bugf("%v (%v) is not assignable to param %d of %v",
				param.val, param.val.Type(), param.index, funcType)
			errs = errs.AddErrors(err)
			continue
		}
		params[param.index] = param.val
	}

	if errs.HasError() {
		return nil, errs
	}

	for i, param := range params {
		if !param.IsValid() {
			errs = errs.AddErrors(errors.Bugf("missing parameter at index %v", i))
		}
	}

	if errs.HasError() {
		return nil, errs
	}

	return params, nil
}

func (v *funcValuer) String() string {
	return fmt.Sprintf("Func: %+v", v.funcVal.Type())
}

func (v *funcValuer) Clone() Valuer {
	return &funcValuer{v.funcVal}
}

func (v *funcValuer) Equal(other interface{}) bool {
	o, ok := other.(*funcValuer)
	if !ok {
		return false
	}

	if v == nil || o == nil {
		return v == nil && o == nil
	}

	return v.funcVal == o.funcVal
}

func Func(funcVal reflect.Value) Valuer {
	return &funcValuer{funcVal}
}
