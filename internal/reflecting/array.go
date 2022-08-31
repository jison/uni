package reflecting

import (
	"reflect"

	"github.com/jison/uni/internal/errors"
)

func ArrayOfReflectValues(rVals []reflect.Value) ([]interface{}, error) {
	vals := make([]interface{}, 0, len(rVals))
	for _, v := range rVals {
		if !v.CanInterface() {
			return nil, errors.Newf("%v .CanInterface() is false", v)
		}
		vals = append(vals, v.Interface())
	}

	return vals, nil
}

func ReflectValuesOf(vals ...interface{}) ([]reflect.Value, error) {
	rVals := make([]reflect.Value, 0, len(vals))
	for _, v := range vals {
		rVal := reflect.ValueOf(v)
		if !rVal.IsValid() {
			return nil, errors.Newf("reflect.ValueOf(%v) is invalid", v)
		}
		rVals = append(rVals, rVal)
	}

	return rVals, nil
}
