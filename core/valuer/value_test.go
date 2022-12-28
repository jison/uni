package valuer

import (
	"reflect"
	"testing"

	"github.com/jison/uni/internal/errors"
	"github.com/jison/uni/internal/reflecting"
	"github.com/stretchr/testify/assert"
)

func Test_singleValue(t *testing.T) {
	t.Run("SingleValue", func(t *testing.T) {
		t.Run("error", func(t *testing.T) {
			err := errors.Newf("this is an error")
			r := SingleValue(reflect.ValueOf(err))
			err2, ok := r.AsError()
			assert.True(t, ok)
			assert.Equal(t, err, err2)
		})

		t.Run("no error", func(t *testing.T) {
			val := 123
			res := SingleValue(reflect.ValueOf(val))
			val2, ok := res.AsSingle()
			assert.True(t, ok)
			assert.Equal(t, reflect.ValueOf(val), val2)
		})
	})

	t.Run("AsSingle", func(t *testing.T) {
		val := 123
		sv := &singleValue{val: reflect.ValueOf(val)}
		val2, ok := sv.AsSingle()
		assert.True(t, ok)
		assert.Equal(t, reflect.ValueOf(val), val2)
	})

	t.Run("AsError", func(t *testing.T) {
		val := 123
		sv := &singleValue{val: reflect.ValueOf(val)}
		err, ok := sv.AsError()
		assert.False(t, ok)
		assert.Nil(t, err)
	})

	t.Run("AsArray", func(t *testing.T) {
		val := 123
		sv := &singleValue{val: reflect.ValueOf(val)}
		arr, ok := sv.AsArray()
		assert.False(t, ok)
		assert.Nil(t, arr)
	})

	t.Run("Interface", func(t *testing.T) {
		t.Run("value is invalid", func(t *testing.T) {
			sv := &singleValue{val: reflect.ValueOf(nil)}
			_, err := sv.Interface()
			assert.NotNil(t, err)
		})

		t.Run("value can not Interface()", func(t *testing.T) {
			type testStruct struct {
				a int
			}
			ts := testStruct{123}
			tsVal := reflect.ValueOf(ts)
			sv := &singleValue{val: tsVal.FieldByName("a")}
			_, err := sv.Interface()
			assert.NotNil(t, err)
		})

		t.Run("can Interface()", func(t *testing.T) {
			val := 123
			sv := &singleValue{val: reflect.ValueOf(val)}
			val2, err := sv.Interface()
			assert.Nil(t, err)
			assert.Equal(t, val, val2)
		})
	})

	t.Run("initialized", func(t *testing.T) {
		val := 123
		sv := &singleValue{val: reflect.ValueOf(val)}
		assert.True(t, sv.Initialized())
	})
}

func Test_errorValue(t *testing.T) {
	t.Run("ErrorValue", func(t *testing.T) {
		t.Run("error is nil", func(t *testing.T) {
			ev := ErrorValue(nil)
			val, ok := ev.AsSingle()
			assert.True(t, ok)
			assert.False(t, val.IsValid())
		})

		t.Run("error is not nil", func(t *testing.T) {
			err := errors.Newf("this is an error")
			ev := ErrorValue(err)
			err2, ok := ev.AsError()
			assert.True(t, ok)
			assert.Equal(t, err, err2)
		})
	})

	t.Run("AsSingle", func(t *testing.T) {
		err := errors.Newf("this is an error")
		ev := &errorValue{err: err}
		val, ok := ev.AsSingle()
		assert.False(t, ok)
		assert.False(t, val.IsValid())
	})

	t.Run("AsError", func(t *testing.T) {
		err := errors.Newf("this is an error")
		ev := &errorValue{err: err}
		err2, ok := ev.AsError()
		assert.True(t, ok)
		assert.Equal(t, err, err2)
	})

	t.Run("AsArray", func(t *testing.T) {
		err := errors.Newf("this is an error")
		ev := &errorValue{err: err}
		arr, ok := ev.AsArray()
		assert.False(t, ok)
		assert.Nil(t, arr)
	})

	t.Run("Interface", func(t *testing.T) {
		err := errors.Newf("this is an error")
		ev := &errorValue{err: err}
		val, err2 := ev.Interface()
		assert.Nil(t, val)
		assert.Equal(t, err, err2)
	})

	t.Run("initialized", func(t *testing.T) {
		err := errors.Newf("this is an error")
		ev := &errorValue{err: err}
		assert.True(t, ev.Initialized())
	})
}

func Test_arrayValue(t *testing.T) {
	t.Run("ArrayValue", func(t *testing.T) {
		vals, _ := reflecting.ReflectValuesOf(123, "abc")
		av := ArrayValue(vals)
		vals2, ok := av.AsArray()
		assert.True(t, ok)
		assert.Equal(t, vals, vals2)
	})

	t.Run("AsSingle", func(t *testing.T) {
		vals, _ := reflecting.ReflectValuesOf(123, "abc")
		av := &arrayValue{arr: vals}
		vals2, ok := av.AsSingle()
		assert.False(t, ok)
		assert.False(t, vals2.IsValid())
	})

	t.Run("AsError", func(t *testing.T) {
		vals, _ := reflecting.ReflectValuesOf(123, "abc")
		av := &arrayValue{arr: vals}
		err, ok := av.AsError()
		assert.False(t, ok)
		assert.Nil(t, err)
	})

	t.Run("AsArray", func(t *testing.T) {
		vals, _ := reflecting.ReflectValuesOf(123, "abc")
		av := &arrayValue{arr: vals}
		vals2, ok := av.AsArray()
		assert.True(t, ok)
		assert.Equal(t, vals, vals2)
	})

	t.Run("Interface", func(t *testing.T) {
		vals, _ := reflecting.ReflectValuesOf(123, "abc")
		av := &arrayValue{arr: vals}
		val, err := av.Interface()
		assert.Nil(t, err)
		assert.Equal(t, []interface{}{123, "abc"}, val)
	})

	t.Run("initialized", func(t *testing.T) {
		vals, _ := reflecting.ReflectValuesOf(123, "abc")
		av := &arrayValue{arr: vals}
		assert.True(t, av.Initialized())
	})
}

func Test_lazyValue(t *testing.T) {
	t.Run("LazyValue", func(t *testing.T) {
		t.Run("function is nil", func(t *testing.T) {
			lv := LazyValue(nil)
			res, ok := lv.AsError()
			assert.True(t, ok)
			assert.NotNil(t, res)
		})

		t.Run("function is not nil", func(t *testing.T) {
			lv := LazyValue(func() Value {
				return SingleValue(reflect.ValueOf(1))
			})
			res, ok := lv.AsSingle()
			assert.True(t, ok)
			assert.Equal(t, reflect.ValueOf(1), res)
		})
	})

	t.Run("AsSingle", func(t *testing.T) {
		lv := lazyValue{f: func() Value {
			return SingleValue(reflect.ValueOf(1))
		}}
		res, ok := lv.AsSingle()
		assert.True(t, ok)
		assert.Equal(t, reflect.ValueOf(1), res)
	})

	t.Run("AsError", func(t *testing.T) {
		err := errors.Newf("this is an error")
		lv := lazyValue{f: func() Value {
			return ErrorValue(err)
		}}
		res, ok := lv.AsError()
		assert.True(t, ok)
		assert.Equal(t, err, res)
	})

	t.Run("AsArray", func(t *testing.T) {
		vals, _ := reflecting.ReflectValuesOf(123, "abc")
		lv := lazyValue{f: func() Value {
			return ArrayValue(vals)
		}}

		res, ok := lv.AsArray()
		assert.True(t, ok)
		assert.Equal(t, vals, res)
	})

	t.Run("Interface", func(t *testing.T) {
		lv := lazyValue{f: func() Value {
			return SingleValue(reflect.ValueOf(1))
		}}
		val, err := lv.Interface()
		assert.Nil(t, err)
		assert.Equal(t, 1, val)
	})

	t.Run("initialized", func(t *testing.T) {
		vals, _ := reflecting.ReflectValuesOf(123, "abc")
		lv := lazyValue{f: func() Value {
			return ArrayValue(vals)
		}}
		assert.False(t, lv.Initialized())

		_, _ = lv.AsArray()

		assert.True(t, lv.Initialized())
	})
}

func TestValuesOf(t *testing.T) {
	t.Run("value", func(t *testing.T) {
		vals := ValuesOf(123, "abc")
		for i, val := range vals {
			res, ok := val.AsSingle()
			assert.True(t, ok)
			if i == 0 {
				assert.Equal(t, reflect.ValueOf(123), res)
			} else {
				assert.Equal(t, reflect.ValueOf("abc"), res)
			}
		}
	})

	t.Run("error", func(t *testing.T) {
		err1 := errors.Newf("abc")
		err2 := errors.Newf("def")

		vals := ValuesOf(err1, err2)
		for i, val := range vals {
			res, ok := val.AsError()
			assert.True(t, ok)
			if i == 0 {
				assert.Equal(t, err1, res)
			} else {
				assert.Equal(t, err2, res)
			}
		}
	})

	t.Run("array", func(t *testing.T) {
		arr1 := []interface{}{123, "abc"}
		arr2 := []interface{}{456, "def"}

		vals := ValuesOf(arr1, arr2)
		for i, val := range vals {
			res, ok := val.AsArray()
			assert.True(t, ok)
			if i == 0 {
				vals1, _ := reflecting.ReflectValuesOf(arr1...)
				assert.Equal(t, vals1, res)
			} else {
				vals2, _ := reflecting.ReflectValuesOf(arr2...)
				assert.Equal(t, vals2, res)
			}
		}
	})
}
