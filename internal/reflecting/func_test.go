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
		assert.Equal(t, int64(5+7), res[0].Int())
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
		assert.Equal(t, 1, len(res))
		assert.Equal(t, res[0].Int(), int64(8))
	})

	t.Run("variadic with no slice parameter", func(t *testing.T) {
		f := func(arr ...int) int {
			s := 0
			for _, i := range arr {
				s += i
			}
			return s
		}
		rf := ReflectFuncOfFunc(reflect.ValueOf(f))
		params := []reflect.Value{reflect.ValueOf([2]int{3, 5})}
		_, err := rf(params)
		t.Logf("%v\n", err)
		assert.NotNil(t, err)
	})

	t.Run("error input", func(t *testing.T) {
		params := []reflect.Value{reflect.ValueOf(5), reflect.ValueOf(7)}

		f1 := func(a, b int) (int, error) {
			return a + b, nil
		}
		rf1 := ReflectFuncOfFunc(reflect.ValueOf(f1))
		res1, err1 := rf1(params)
		assert.Nil(t, err1)
		assert.Equal(t, 2, len(res1))
		assert.Equal(t, int64(5+7), res1[0].Int())
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

	t.Run("invalid parameter value", func(t *testing.T) {
		params := []reflect.Value{reflect.ValueOf(5), reflect.ValueOf(nil)}

		f1 := func(a, b int) (int, error) {
			return a + b, nil
		}
		rf1 := ReflectFuncOfFunc(reflect.ValueOf(f1))
		_, err := rf1(params)
		assert.NotNil(t, err)
	})

	t.Run("return nil error", func(t *testing.T) {
		f := func(a, b int) (int, error) {
			return a + b, nil
		}
		rf := ReflectFuncOfFunc(reflect.ValueOf(f))
		params := []reflect.Value{reflect.ValueOf(5), reflect.ValueOf(7)}
		res, err := rf(params)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(res))
		assert.Equal(t, int64(5+7), res[0].Int())
		assert.Nil(t, res[1].Interface())
	})
}
