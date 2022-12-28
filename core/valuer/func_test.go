package valuer

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jison/uni/internal/errors"
	"github.com/stretchr/testify/assert"
)

func testFunc1(a string, b int) string {
	return a + fmt.Sprintf("%v", b)
}

func testFunc2() {}

func Test_funcValuer_Value(t *testing.T) {
	t.Run("with normal function", func(t *testing.T) {
		valuer := Func(reflect.ValueOf(testFunc1))
		inputs := ValuesOf(
			funcParam{
				index: 1,
				val:   reflect.ValueOf(123),
			},
			funcParam{
				index: 0,
				val:   reflect.ValueOf("abc"),
			},
		)
		res := valuer.Value(inputs)
		arr, ok := res.AsArray()

		assert.True(t, ok)
		assert.Equal(t, 1, len(arr))
		assert.Equal(t, "abc123", arr[0].Interface())
	})

	t.Run("with wrong parameter number", func(t *testing.T) {
		valuer := Func(reflect.ValueOf(testFunc1))
		inputs := ValuesOf(
			funcParam{
				index: 1,
				val:   reflect.ValueOf(123),
			},
		)
		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("with wrong parameter number 2", func(t *testing.T) {
		valuer := Func(reflect.ValueOf(testFunc1))

		inputs := ValuesOf(
			funcParam{
				index: 1,
				val:   reflect.ValueOf(123),
			},
			funcParam{
				index: 0,
				val:   reflect.ValueOf("abc"),
			},
			funcParam{
				index: 2,
				val:   reflect.ValueOf(123),
			},
		)
		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("with wrong parameter number 3", func(t *testing.T) {
		valuer := Func(reflect.ValueOf(testFunc1))

		inputs := ValuesOf(
			funcParam{
				index: 1,
				val:   reflect.ValueOf(123),
			},
			funcParam{
				index: 2,
				val:   reflect.ValueOf(123),
			},
		)
		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("with wrong parameter type", func(t *testing.T) {
		valuer := Func(reflect.ValueOf(testFunc1))
		inputs := ValuesOf(
			funcParam{index: 1, val: reflect.ValueOf(123)},
			funcParam{index: 0, val: reflect.ValueOf(123)},
		)
		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("with duplicate parameter index", func(t *testing.T) {
		valuer := Func(reflect.ValueOf(testFunc1))
		inputs := ValuesOf(
			funcParam{index: 1, val: reflect.ValueOf(123)},
			funcParam{index: 1, val: reflect.ValueOf(123)},
		)
		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("with error parameter value", func(t *testing.T) {
		valuer := Func(reflect.ValueOf(testFunc1))
		err1 := errors.Newf("i am error")
		inputs := ValuesOf(
			funcParam{index: 1, val: reflect.ValueOf(123)},
			funcParam{index: 0, val: reflect.ValueOf(err1)},
		)
		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("with error input 2", func(t *testing.T) {
		valuer := Func(reflect.ValueOf(testFunc1))
		err1 := errors.Newf("i am error")
		inputs := ValuesOf(
			funcParam{index: 1, val: reflect.ValueOf(123)},
			err1,
		)
		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, err1)
	})

	t.Run("invalid input", func(t *testing.T) {
		type testStruct struct {
			a int
		}
		ts := testStruct{123}
		tsVal := reflect.ValueOf(ts)

		tests := []struct {
			name  string
			input Value
		}{
			{"input is not a single value", ArrayValue([]reflect.Value{})},
			{"CanInterface() is false", SingleValue(tsVal.FieldByName("a"))},
			{"input is not a funcParam", SingleValue(reflect.ValueOf(123))},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				valuer := Func(reflect.ValueOf(testFunc1))
				inputs := []Value{
					SingleValue(reflect.ValueOf(funcParam{index: 1, val: reflect.ValueOf(123)})),
					tt.input,
				}
				res := valuer.Value(inputs)
				err, ok := res.AsError()
				assert.True(t, ok)
				assert.NotNil(t, err)
			})
		}
	})

	t.Run("with no return value", func(t *testing.T) {
		func1 := func(a, b int) {

		}
		valuer := Func(reflect.ValueOf(func1))
		inputs := ValuesOf(
			funcParam{index: 1, val: reflect.ValueOf(1)},
			funcParam{index: 0, val: reflect.ValueOf(2)},
		)
		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.False(t, ok)
		assert.Nil(t, err)
	})

	t.Run("with only one return value", func(t *testing.T) {
		valuer := Func(reflect.ValueOf(testFunc1))
		inputs := ValuesOf(
			funcParam{index: 1, val: reflect.ValueOf(123)},
			funcParam{index: 0, val: reflect.ValueOf("abc")},
		)
		res := valuer.Value(inputs)
		arr, ok := res.AsArray()
		assert.True(t, ok)
		assert.Equal(t, 1, len(arr))
		assert.Equal(t, "abc123", arr[0].Interface())
	})

	t.Run("with multiple return value", func(t *testing.T) {
		func1 := func(a int, b string) (string, int) {
			return b, a
		}
		valuer := Func(reflect.ValueOf(func1))
		inputs := ValuesOf(
			funcParam{index: 1, val: reflect.ValueOf("abc")},
			funcParam{index: 0, val: reflect.ValueOf(123)},
		)
		res := valuer.Value(inputs)
		arr, ok := res.AsArray()
		assert.True(t, ok)
		assert.Equal(t, 2, len(arr))

		assert.Equal(t, "abc", arr[0].Interface())
		assert.Equal(t, 123, arr[1].Interface())
	})

	t.Run("function return error", func(t *testing.T) {
		err := errors.Newf("this is an error")
		func1 := func(a int, b string) (string, int, error) {
			return b, a, err
		}
		valuer := Func(reflect.ValueOf(func1))
		inputs := ValuesOf(
			funcParam{index: 1, val: reflect.ValueOf("abc")},
			funcParam{index: 0, val: reflect.ValueOf(123)},
		)
		res := valuer.Value(inputs)
		err2, ok := res.AsError()
		assert.True(t, ok)
		assert.ErrorIs(t, err2, err)
	})

	t.Run("with not func value", func(t *testing.T) {
		valuer := Func(reflect.ValueOf(123))
		res := valuer.Value(nil)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("variadic", func(t *testing.T) {
		func1 := func(a string, arr ...int) int {
			v := 0
			for _, i := range arr {
				v += i
			}

			return v
		}
		valuer := Func(reflect.ValueOf(func1))
		inputs := ValuesOf(
			funcParam{index: 0, val: reflect.ValueOf("abc")},
			funcParam{index: 1, val: reflect.ValueOf([]int{3, 5})},
		)
		res := valuer.Value(inputs)
		err, _ := res.AsError()
		fmt.Printf("%v\n", err)

		arr, ok := res.AsArray()
		assert.True(t, ok)
		assert.Equal(t, 1, len(arr))

		assert.Equal(t, 8, arr[0].Interface())
	})
}

func Test_funcValuer_String(t *testing.T) {
	valuer := Func(reflect.ValueOf(testFunc1))
	assert.Equal(t, "Func: func(string, int) string", valuer.String())
}

func Test_funcValuer_Clone(t *testing.T) {
	v1 := Func(reflect.ValueOf(testFunc1))
	v2 := v1.Clone()

	assert.False(t, v1 == v2)
	assert.Equal(t, v1, v2)
	assert.True(t, v1.Equal(v2))
}

func Test_funcValuer_Equal(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		v1 := Func(reflect.ValueOf(testFunc1))
		v2 := Func(reflect.ValueOf(testFunc1))
		assert.True(t, v1.Equal(v2))
	})

	t.Run("not equal", func(t *testing.T) {
		v1 := Func(reflect.ValueOf(testFunc1))
		v2 := Func(reflect.ValueOf(testFunc2))
		v3 := Identity()
		assert.False(t, v1.Equal(v2))
		assert.False(t, v1.Equal(v3))
	})

	t.Run("nil", func(t *testing.T) {
		var v1 *funcValuer
		var v2 *funcValuer
		var v3 = Func(reflect.ValueOf(testFunc1))
		assert.True(t, v1.Equal(v2))
		assert.False(t, v1.Equal(v3))
		assert.False(t, v3.Equal(v1))
	})
}
