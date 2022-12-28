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
				// if function is variadic then the last parameter should be slice type
				//if param.Type().Kind() != reflect.Slice {
				//	errs = errs.AddErrors(errors.Bugf("variadic param should pass slice value"))
				//	continue
				//}
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

				// can not construct a invalid return value
				//if !errVal.IsValid() {
				//	errs = errs.AddErrors(errors.Newf("provider function return an invalid value"))
				//	continue
				//}

				if errVal.IsNil() {
					continue
				}

				// can not construct a return value that CanInterface() is false
				//if !errVal.CanInterface() {
				//	errs = errs.AddErrors(errors.Newf("error can not convert to interface{}"))
				//	continue
				//}

				errs = errs.AddErrors(errVal.Interface().(error))
			}
		}

		if errs.HasError() {
			return nil, errs
		}

		return rValues, nil
	}
}
