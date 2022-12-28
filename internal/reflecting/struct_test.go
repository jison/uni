package reflecting

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Field1 string
}

type testStructWithUnexportedField struct {
	Field1  string
	field2  string
	_field3 string
}

func TestInitStructWithValues(t *testing.T) {
	t.Run("no error input", func(t *testing.T) {
		structType := reflect.TypeOf(testStruct{})
		values := map[string]interface{}{
			"Field1": "a",
			"Field2": "b",
		}
		val, err := InitStructWithValues(structType, values)
		assert.Nil(t, err)
		assert.Equal(t, val.Interface(), testStruct{Field1: "a"})
	})

	t.Run("no struct type", func(t *testing.T) {
		aType := reflect.TypeOf("a")
		values := map[string]interface{}{
			"Field1": "a",
			"Field2": "b",
		}
		_, err := InitStructWithValues(aType, values)
		assert.NotNil(t, err)
	})
}

func TestInitStructWithReflectValues(t *testing.T) {
	t.Run("no error input", func(t *testing.T) {
		structType := reflect.TypeOf(testStruct{})
		values := map[string]reflect.Value{
			"Field1": reflect.ValueOf("a"),
			"Field2": reflect.ValueOf("b"),
		}
		val, err := InitStructWithReflectValues(structType, values)
		assert.Nil(t, err)
		assert.Equal(t, val.Interface(), testStruct{Field1: "a"})
	})

	t.Run("struct with unexported field", func(t *testing.T) {
		structType := reflect.TypeOf(testStructWithUnexportedField{})
		values := map[string]reflect.Value{
			"Field1":  reflect.ValueOf("a"),
			"field2":  reflect.ValueOf("b"),
			"_field3": reflect.ValueOf("c"),
		}
		val, err := InitStructWithReflectValues(structType, values)
		assert.Nil(t, err)
		assert.Equal(t, val.Interface(), testStructWithUnexportedField{Field1: "a", field2: "b", _field3: "c"})
	})

	t.Run("struct with can not Interface() value", func(t *testing.T) {
		ts := testStructWithUnexportedField{field2: "abc"}
		tsVal := reflect.ValueOf(ts)

		structType := reflect.TypeOf(testStruct{})
		values := map[string]reflect.Value{
			"Field1": tsVal.FieldByName("field2"),
		}
		_, err := InitStructWithReflectValues(structType, values)
		assert.NotNil(t, err)
	})

	t.Run("init pointer of struct", func(t *testing.T) {
		structPtrType := reflect.TypeOf(&testStruct{})
		values := map[string]reflect.Value{
			"Field1": reflect.ValueOf("a"),
			"Field2": reflect.ValueOf("b"),
		}
		val, err := InitStructWithReflectValues(structPtrType, values)
		assert.Nil(t, err)
		assert.Equal(t, val.Kind(), reflect.Ptr)
		assert.Equal(t, val.Interface(), &testStruct{Field1: "a"})
	})

	t.Run("init pointer of struct with unexported fields", func(t *testing.T) {
		structPtrType := reflect.TypeOf(&testStructWithUnexportedField{})
		values := map[string]reflect.Value{
			"Field1":  reflect.ValueOf("a"),
			"field2":  reflect.ValueOf("b"),
			"_field3": reflect.ValueOf("c"),
		}
		val, err := InitStructWithReflectValues(structPtrType, values)
		assert.Nil(t, err)
		assert.Equal(t, val.Kind(), reflect.Ptr)
		assert.Equal(t, val.Interface(), &testStructWithUnexportedField{Field1: "a", field2: "b", _field3: "c"})
	})
}

func TestUpdateStructValues(t *testing.T) {
	t.Run("no error input", func(t *testing.T) {
		s := testStruct{}
		values := map[string]interface{}{
			"Field1": "a",
			"Field2": "b",
		}
		err := UpdateStructFields(s, values)
		assert.NotNil(t, err)
	})

	t.Run("return pointer of struct", func(t *testing.T) {
		s := &testStruct{}
		values := map[string]interface{}{
			"Field1": "a",
			"Field2": "b",
		}
		err := UpdateStructFields(s, values)
		assert.Nil(t, err)
		assert.Equal(t, s.Field1, "a")
	})

	t.Run("wrong field type", func(t *testing.T) {
		s := &testStruct{}
		values := map[string]interface{}{
			"Field1": 123,
			"Field2": "b",
		}
		err := UpdateStructFields(s, values)
		assert.NotNil(t, err)
	})
}
