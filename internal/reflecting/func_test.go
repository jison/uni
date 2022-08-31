package reflecting

import (
	"reflect"
	"testing"

	"github.com/jison/uni/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestReflectFuncOfFunc(t *testing.T) {

	t.Run("no error input", func(t *testing.T) {
		f := func(a, b int) int {
			return a + b
		}
		rf := ReflectFuncOfFunc(reflect.ValueOf(f))
		params := []reflect.Value{reflect.ValueOf(5), reflect.ValueOf(7)}
		res, err := rf(params)
		assert.Nil(t, err)
		assert.Equal(t, len(res), 1)
		assert.Equal(t, res[0].Int(), int64(5+7))
	})

	t.Run("variadic", func(t *testing.T) {
		f := func(arr ...int) int {
			s := 0
			for _, i := range arr {
				s += i
			}
			return s
		}
		rf := ReflectFuncOfFunc(reflect.ValueOf(f))
		params := []reflect.Value{reflect.ValueOf([]int{3, 5})}
		res, err := rf(params)
		assert.Nil(t, err)
		assert.Equal(t, len(res), 1)
		assert.Equal(t, int64(8), res[0].Int())
	})

	t.Run("error input", func(t *testing.T) {
		params := []reflect.Value{reflect.ValueOf(5), reflect.ValueOf(7)}

		f1 := func(a, b int) (int, error) {
			return a + b, nil
		}
		rf1 := ReflectFuncOfFunc(reflect.ValueOf(f1))
		res1, err1 := rf1(params)
		assert.Nil(t, err1)
		assert.Equal(t, len(res1), 2)
		assert.Equal(t, res1[0].Int(), int64(5+7))
		assert.True(t, res1[1].IsNil())

		f2 := func(a, b int) (int, error) {
			return 0, errors.Newf("this is an error")
		}
		rf2 := ReflectFuncOfFunc(reflect.ValueOf(f2))
		res2, err2 := rf2(params)
		assert.NotNil(t, err2)
		assert.Nil(t, res2)
	})

	t.Run("error input 2", func(t *testing.T) {
		params := []reflect.Value{reflect.ValueOf(5), reflect.ValueOf(7)}

		f := func(a, b int) (int, error) {
			return 0, errors.Newf("this is an error")
		}
		rf := ReflectFuncOfFunc(reflect.ValueOf(f))
		res, err := rf(params)
		assert.NotNil(t, err)
		assert.Nil(t, res)
	})

	t.Run("wrong parameter number", func(t *testing.T) {
		params := []reflect.Value{reflect.ValueOf(5), reflect.ValueOf(7), reflect.ValueOf(9)}

		f1 := func(a, b int) (int, error) {
			return a + b, nil
		}
		rf1 := ReflectFuncOfFunc(reflect.ValueOf(f1))
		_, err := rf1(params)
		assert.NotNil(t, err)
	})

	t.Run("wrong parameter type", func(t *testing.T) {
		params := []reflect.Value{reflect.ValueOf(5), reflect.ValueOf("abc")}

		f1 := func(a, b int) (int, error) {
			return a + b, nil
		}
		rf1 := ReflectFuncOfFunc(reflect.ValueOf(f1))
		_, err := rf1(params)
		assert.NotNil(t, err)
	})

	t.Run("no function value", func(t *testing.T) {
		rf := ReflectFuncOfFunc(reflect.ValueOf(1))
		params := []reflect.Value{reflect.ValueOf(5), reflect.ValueOf(7)}
		_, err := rf(params)
		assert.NotNil(t, err)
	})
}

type testStruct1 struct {
	A int
	B string
}

type testStruct2 struct {
	a int
	B string
	_ interface{}
}

func TestReflectFuncOfStruct(t *testing.T) {
	t.Run("no error input", func(t *testing.T) {
		rf := ReflectFuncOfStruct(reflect.TypeOf(testStruct1{}), false)
		params := []reflect.Value{reflect.ValueOf(123), reflect.ValueOf("abc")}
		res, err := rf(params)
		assert.Nil(t, err)
		assert.Equal(t, len(res), 1)
		val := res[0].Interface().(testStruct1)
		assert.Equal(t, val.A, 123)
		assert.Equal(t, val.B, "abc")
	})

	t.Run("no struct type", func(t *testing.T) {
		rf := ReflectFuncOfStruct(reflect.TypeOf(1), false)
		params := []reflect.Value{reflect.ValueOf(123), reflect.ValueOf("abc")}
		_, err := rf(params)
		assert.NotNil(t, err)
	})

	t.Run("return pointer of struct", func(t *testing.T) {
		rf := ReflectFuncOfStruct(reflect.TypeOf(testStruct1{}), true)
		params := []reflect.Value{reflect.ValueOf(123), reflect.ValueOf("abc")}
		res, err := rf(params)
		assert.Nil(t, err)
		assert.Equal(t, len(res), 1)
		assert.True(t, res[0].Kind() == reflect.Ptr)
		val := res[0].Interface().(*testStruct1)
		assert.Equal(t, val.A, 123)
		assert.Equal(t, val.B, "abc")
	})

	t.Run("struct with unexported field", func(t *testing.T) {
		rf := ReflectFuncOfStruct(reflect.TypeOf(testStruct2{}), false)
		params := []reflect.Value{reflect.ValueOf("abc")}
		res, err := rf(params)
		assert.Nil(t, err)
		assert.Equal(t, len(res), 1)
		val := res[0].Interface().(testStruct2)
		assert.Equal(t, val.a, 0)
		assert.Equal(t, val.B, "abc")
	})
}
