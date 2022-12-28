package valuer

import (
	"reflect"
	"testing"

	"github.com/jison/uni/internal/errors"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	a int
	B string
}

func Test_structValuer_Valuer(t *testing.T) {
	t.Run("structType is nil", func(t *testing.T) {
		valuer := Struct(nil)
		inputs := ValuesOf(
			structField{
				name: "a",
				val:  reflect.ValueOf(123),
			},
			structField{
				name: "B",
				val:  reflect.ValueOf("abc"),
			},
			structField{
				name: "C",
				val:  reflect.ValueOf([]int{1, 2, 3}),
			},
		)

		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("no error input", func(t *testing.T) {
		valuer := Struct(reflect.TypeOf(testStruct{}))
		inputs := ValuesOf(
			structField{
				name: "a",
				val:  reflect.ValueOf(123),
			},
			structField{
				name: "B",
				val:  reflect.ValueOf("abc"),
			},
			structField{
				name: "C",
				val:  reflect.ValueOf([]int{1, 2, 3}),
			},
		)

		res := valuer.Value(inputs)
		rVal, ok := res.AsSingle()
		assert.True(t, ok)
		assert.Equal(t, testStruct{a: 123, B: "abc"}, rVal.Interface())
	})

	t.Run("return pointer of struct", func(t *testing.T) {
		valuer := Struct(reflect.TypeOf(&testStruct{}))
		inputs := ValuesOf(
			structField{
				name: "a",
				val:  reflect.ValueOf(123),
			},
			structField{
				name: "B",
				val:  reflect.ValueOf("abc"),
			},
			structField{
				name: "C",
				val:  reflect.ValueOf([]int{1, 2, 3}),
			},
		)

		res := valuer.Value(inputs)
		rVal, ok := res.AsSingle()
		assert.True(t, ok)
		assert.Equal(t, reflect.Ptr, rVal.Kind())
		assert.Equal(t, testStruct{a: 123, B: "abc"}, rVal.Elem().Interface())
	})

	t.Run("input is not structField", func(t *testing.T) {
		valuer := Struct(reflect.TypeOf(testStruct{}))
		inputs := ValuesOf(
			[]interface{}{1, 2},
			structField{
				name: "B",
				val:  reflect.ValueOf("abc"),
			},
			funcParam{
				index: 0,
				val:   reflect.ValueOf(123),
			},
		)

		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("CanInterface() of input equal false", func(t *testing.T) {
		ts := testStruct{a: 123}
		tsVal := reflect.ValueOf(ts)

		valuer := Struct(reflect.TypeOf(testStruct{}))
		inputs := []Value{
			SingleValue(tsVal.FieldByName("a")),
			SingleValue(reflect.ValueOf(structField{
				name: "B",
				val:  reflect.ValueOf("abc"),
			})),
			SingleValue(reflect.ValueOf(structField{
				name: "C",
				val:  reflect.ValueOf([]int{1, 2, 3}),
			})),
		}

		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("wrong field name", func(t *testing.T) {
		valuer := Struct(reflect.TypeOf(testStruct{}))
		inputs := ValuesOf(
			structField{
				name: "A",
				val:  reflect.ValueOf(123),
			},
			structField{
				name: "B",
				val:  reflect.ValueOf("abc"),
			},
			structField{
				name: "C",
				val:  reflect.ValueOf([]int{1, 2, 3}),
			},
		)

		res := valuer.Value(inputs)
		rVal, ok := res.AsSingle()
		assert.True(t, ok)
		assert.Equal(t, testStruct{a: 0, B: "abc"}, rVal.Interface())
	})

	t.Run("duplicate field name", func(t *testing.T) {
		valuer := Struct(reflect.TypeOf(testStruct{}))
		inputs := ValuesOf(
			structField{
				name: "a",
				val:  reflect.ValueOf(123),
			},
			structField{
				name: "B",
				val:  reflect.ValueOf("abc"),
			},
			structField{
				name: "a",
				val:  reflect.ValueOf(456),
			},
		)

		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("wrong field type", func(t *testing.T) {
		valuer := Struct(reflect.TypeOf(testStruct{}))
		inputs := ValuesOf(
			structField{
				name: "a",
				val:  reflect.ValueOf("abc"),
			},
			structField{
				name: "B",
				val:  reflect.ValueOf(123),
			},
			structField{
				name: "C",
				val:  reflect.ValueOf([]int{1, 2, 3}),
			},
		)

		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("field with error value", func(t *testing.T) {
		valuer := Struct(reflect.TypeOf(testStruct{}))
		err1 := errors.Newf("this is error1")
		inputs := ValuesOf(
			structField{
				name: "a",
				val:  reflect.ValueOf(123),
			},
			err1,
		)
		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, err1))
	})

	t.Run("field with error value 2", func(t *testing.T) {
		valuer := Struct(reflect.TypeOf(testStruct{}))
		err1 := errors.Newf("this is error1")
		inputs := ValuesOf(
			structField{
				name: "a",
				val:  reflect.ValueOf(123),
			},
			structField{
				name: "B",
				val:  reflect.ValueOf(err1),
			},
		)
		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("no struct type", func(t *testing.T) {
		valuer := Struct(reflect.TypeOf(123))
		inputs := ValuesOf(
			structField{
				name: "a",
				val:  reflect.ValueOf(123),
			},
			structField{
				name: "B",
				val:  reflect.ValueOf("abc"),
			},
			structField{
				name: "C",
				val:  reflect.ValueOf([]int{1, 2, 3}),
			},
		)

		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})
}

func Test_structValuer_String(t *testing.T) {
	valuer := Struct(reflect.TypeOf(testStruct{}))
	assert.Equal(t, "Struct: valuer.testStruct", valuer.String())
}

func Test_structValuer_Clone(t *testing.T) {
	v1 := Struct(reflect.TypeOf(123))
	v2 := v1.Clone()

	assert.False(t, v1 == v2)
	assert.Equal(t, v1, v2)
	assert.True(t, v1.Equal(v2))
}

func Test_structValuer_Equal(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		v1 := Struct(reflect.TypeOf(testStruct{}))
		v2 := Struct(reflect.TypeOf(testStruct{}))
		assert.True(t, v1.Equal(v2))
	})

	t.Run("not equal", func(t *testing.T) {
		v1 := Struct(reflect.TypeOf(testStruct{}))
		v2 := Struct(reflect.TypeOf(&testStruct{}))
		assert.False(t, v1.Equal(v2))
	})

	t.Run("not equal to other valuer", func(t *testing.T) {
		v1 := Struct(reflect.TypeOf(testStruct{}))
		v2 := OneOf()
		assert.False(t, v1.Equal(v2))
	})

	t.Run("nil", func(t *testing.T) {
		var v1 *structValuer
		var v2 *structValuer
		var v3 = Struct(reflect.TypeOf(testStruct{}))
		assert.True(t, v1.Equal(v2))
		assert.False(t, v1.Equal(v3))
		assert.False(t, v3.Equal(v1))
	})
}
