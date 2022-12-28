package reflecting

import (
	"reflect"
	"unsafe"

	"github.com/jison/uni/internal/errors"
)

func InitStructWithValues(structOrPtrOfStructType reflect.Type, values map[string]interface{}) (reflect.Value, error) {
	reflectValues := make(map[string]reflect.Value)
	for name, value := range values {
		reflectValues[name] = reflect.ValueOf(value)
	}

	return InitStructWithReflectValues(structOrPtrOfStructType, reflectValues)
}

func InitStructWithReflectValues(structOrPtrOfStructType reflect.Type, values map[string]reflect.Value,
) (reflect.Value, error) {
	var structType reflect.Type
	if structOrPtrOfStructType.Kind() == reflect.Ptr {
		structType = structOrPtrOfStructType.Elem()
	} else {
		structType = structOrPtrOfStructType
	}

	structPtrVal := reflect.New(structType)
	err := updateReflectStructFields(structPtrVal, values)
	if err != nil {
		return reflect.Value{}, err
	}

	if structOrPtrOfStructType.Kind() == reflect.Ptr {
		return structPtrVal, nil
	} else {
		return structPtrVal.Elem(), nil
	}
}

func UpdateStructFields(structPtr interface{}, values map[string]interface{}) error {
	val := reflect.ValueOf(structPtr)
	reflectValues := make(map[string]reflect.Value)
	for name, value := range values {
		reflectValues[name] = reflect.ValueOf(value)
	}
	return updateReflectStructFields(val, reflectValues)
}

func updateReflectStructFields(structPtr reflect.Value, values map[string]reflect.Value) error {
	if structPtr.Kind() != reflect.Ptr || structPtr.Elem().Kind() != reflect.Struct {
		// only pointer of struct can do this
		return errors.Newf("param should be a pointer of struct")
	}
	structVal := structPtr.Elem()

	for name, value := range values {
		fieldVal := structVal.FieldByName(name)
		if !fieldVal.IsValid() {
			continue
		}
		if !value.Type().AssignableTo(fieldVal.Type()) {
			return errors.Newf("%v (%v) can not assignable to type %v of field `%v`",
				value, value.Type(), fieldVal.Type(), name)
		}
		if !value.CanInterface() {
			return errors.Newf("%v CanInterface() is false", value)
		}
	}

	for name, value := range values {
		fieldVal := structVal.FieldByName(name)
		if !fieldVal.IsValid() {
			continue
		}

		if fieldVal.CanSet() {
			fieldVal.Set(value)
		} else {
			// if field is unexported
			reflect.NewAt(fieldVal.Type(), unsafe.Pointer(fieldVal.UnsafeAddr())).
				Elem().Set(value)
		}
	}

	return nil
}
