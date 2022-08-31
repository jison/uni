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
	err := UpdateReflectStructFields(structPtrVal, values)
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
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		// only pointer of struct can do this
		return errors.Newf("param should be a pointer of struct")
	}

	reflectValues := make(map[string]reflect.Value)
	for name, value := range values {
		reflectValues[name] = reflect.ValueOf(value)
	}
	return UpdateReflectStructFields(val, reflectValues)
}

func UpdateReflectStructFields(structOrPtrOfStruct reflect.Value, values map[string]reflect.Value) error {
	var structVal reflect.Value
	if structOrPtrOfStruct.Kind() == reflect.Ptr {
		structVal = structOrPtrOfStruct.Elem()
	} else {
		structVal = structOrPtrOfStruct
	}

	if structVal.Kind() != reflect.Struct {
		return errors.Newf("param should be a value of struct or a pointer of struct")
	}

	for name, value := range values {
		fieldVal := structVal.FieldByName(name)
		if !fieldVal.IsValid() {
			continue
		}
		if !value.Type().AssignableTo(fieldVal.Type()) {
			return errors.Newf("%v (%v) can not assignable to type %v of field `%v`",
				value, value.Type(), fieldVal.Type(), name)
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
