package reflecting

import (
	"reflect"

	"github.com/jison/uni/internal/errors"
)

type ReflectFunc func(params []reflect.Value) ([]reflect.Value, error)

func reflectFuncOfError(err error) ReflectFunc {
	return func(params []reflect.Value) ([]reflect.Value, error) {
		return nil, err
	}
}

func ReflectFuncOfFunc(f reflect.Value) ReflectFunc {
	funcType := f.Type()
	if funcType.Kind() != reflect.Func {
		err := errors.Bugf("type %v is not a function", funcType)
		return reflectFuncOfError(err)
	}

	return func(inputs []reflect.Value) ([]reflect.Value, error) {
		if funcType.NumIn() != len(inputs) {
			return nil, errors.Bugf("function need %d parameters, but provided %d",
				funcType.NumIn(), len(inputs))
		}

		errs := errors.Empty()
		var params []reflect.Value
		for i := 0; i < funcType.NumIn(); i++ {
			param := inputs[i]
			if !param.IsValid() {
				errs = errs.AddErrors(errors.Bugf("invalid param"))
				continue
			}
			if !param.Type().AssignableTo(funcType.In(i)) {
				errs = errs.AddErrors(errors.Bugf("%v (%v) is not assignable to param %d of %v",
					param, param.Type(), i, funcType))
				continue
			}
			if i == funcType.NumIn()-1 && funcType.IsVariadic() {
				if param.Type().Kind() != reflect.Slice {
					errs = errs.AddErrors(errors.Bugf("variadic param should pass slice value"))
					continue
				}
				for ii := 0; ii < param.Len(); ii++ {
					params = append(params, param.Index(ii))
				}
				continue
			}

			params = append(params, param)
		}
		if errs.HasError() {
			return nil, errs
		}

		rValues := f.Call(params)

		for i := 0; i < funcType.NumOut(); i++ {
			outType := funcType.Out(i)
			if IsErrorType(outType) {
				errVal := rValues[i]

				if !errVal.IsValid() {
					errs = errs.AddErrors(errors.Newf("provider function return an invalid value"))
					continue
				}
				if errVal.IsNil() {
					continue
				}
				var err error
				if !errVal.CanInterface() {
					err = errors.Newf("error can not convert to interface{}")
				} else {
					err = errVal.Interface().(error)
				}

				errs = errs.AddErrors(err)
			}
		}

		if errs.HasError() {
			return nil, errs
		}

		return rValues, nil
	}
}

func ReflectFuncOfStruct(rType reflect.Type, isPtr bool) ReflectFunc {
	if rType.Kind() != reflect.Struct {
		err := errors.Bugf("parameter of type %v is not a struct", rType)
		return reflectFuncOfError(err)
	}

	return func(params []reflect.Value) ([]reflect.Value, error) {
		val := reflect.New(rType).Elem()
		paramIndex := 0
		for i := 0; i < rType.NumField(); i++ {
			if !val.Field(i).CanSet() {
				continue
			}

			if paramIndex >= len(params) {
				if len(params) != rType.NumField() {
					return nil, errors.Bugf("len of param does not match the struct needs")
				}
			}

			val.Field(i).Set(params[paramIndex])
			paramIndex += 1
		}

		if isPtr {
			if !val.CanAddr() {
				return nil, errors.Newf("value of %v is not addressable", rType)
			}

			val = val.Addr()
		}

		return []reflect.Value{val}, nil
	}
}
