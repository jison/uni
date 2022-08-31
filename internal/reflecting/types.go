package reflecting

import "reflect"

var ErrorType = reflect.TypeOf((*error)(nil)).Elem()
var AnyType = reflect.TypeOf((*interface{})(nil)).Elem()

func IsErrorType(rType reflect.Type) bool {
	if rType == nil {
		return false
	}

	return rType.Implements(ErrorType)
}

func IsErrorValue(val reflect.Value) bool {
	if !val.IsValid() {
		return false
	}

	return IsErrorType(val.Type())
}

func AsError(val reflect.Value) (error, bool) {
	if !val.IsValid() {
		return nil, false
	}

	if !IsErrorType(val.Type()) {
		return nil, false
	}
	if !val.CanInterface() {
		return nil, false
	}
	err, ok := val.Interface().(error)
	return err, ok
}

func IsNilValue(rVal reflect.Value) bool {
	if !rVal.IsValid() {
		return true
	}

	if !rVal.IsZero() {
		return false
	}

	vKind := rVal.Kind()
	if vKind != reflect.Chan && vKind != reflect.Func && vKind != reflect.Interface &&
		vKind != reflect.Map && vKind != reflect.Ptr && vKind != reflect.Slice {
		return false
	}

	return rVal.IsNil()
}

func IsKindOrPtrOfKind(t reflect.Type, kind reflect.Kind) bool {
	if t == nil {
		return false
	}
	if t.Kind() == kind {
		return true
	}
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == kind {
		return true
	}
	return false
}

func CanBeMapKey(val interface{}) bool {
	if val == nil {
		return true
	}

	rType := reflect.TypeOf(val)
	k := rType.Kind()
	return k != reflect.Slice && k != reflect.Map && k != reflect.Func
}
