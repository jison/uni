package valuer

import (
	"reflect"
	"testing"

	"github.com/jison/uni/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestCollectorValuer(t *testing.T) {
	t.Run("no error input", func(t *testing.T) {
		valuer := Collector(reflect.TypeOf(0))
		inputs := ValuesOf(
			1,
			2,
		)
		res := valuer.Value(inputs)
		rVal, ok := res.AsSingle()
		assert.True(t, ok)
		assert.Equal(t, rVal.Interface(), []int{1, 2})
	})

	t.Run("error input", func(t *testing.T) {
		valuer := Collector(reflect.TypeOf(0))
		err := errors.Newf("this is error")
		inputs := ValuesOf(
			1,
			err,
		)
		res := valuer.Value(inputs)
		err2, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err2)
		assert.ErrorIs(t, err2, err)
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
			{"input type is wrong", SingleValue(reflect.ValueOf("abc"))},
			{"input is not a single value", ArrayValue([]reflect.Value{})},
			{"CanInterface() is false", SingleValue(tsVal.FieldByName("a"))},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				valuer := Collector(reflect.TypeOf(0))
				inputs := []Value{
					SingleValue(reflect.ValueOf(1)),
					tt.input,
				}

				res := valuer.Value(inputs)
				err, ok := res.AsError()
				assert.True(t, ok)
				assert.NotNil(t, err)
			})
		}
	})
}

func Test_collectorValuer_String(t *testing.T) {
	valuer := Collector(reflect.TypeOf(0))
	assert.Equal(t, "Collect: int", valuer.String())
}

func Test_collectorValuer_Clone(t *testing.T) {
	v1 := Collector(reflect.TypeOf(0))
	v2 := v1.Clone()

	assert.False(t, v1 == v2)
	assert.Equal(t, v1, v2)
	assert.True(t, v1.Equal(v2))
}

func Test_collectorValuer_Equal(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		v1 := Collector(reflect.TypeOf(0))
		v2 := Collector(reflect.TypeOf(0))
		assert.True(t, v1.Equal(v2))
	})

	t.Run("not equal", func(t *testing.T) {
		v1 := Collector(reflect.TypeOf(0))
		v2 := Collector(reflect.TypeOf(""))
		v3 := Identity()
		assert.False(t, v1.Equal(v2))
		assert.False(t, v1.Equal(v3))
	})

	t.Run("nil", func(t *testing.T) {
		var v1 *collectorValuer
		var v2 *collectorValuer
		var v3 = Collector(reflect.TypeOf(0))
		assert.True(t, v1.Equal(v2))
		assert.False(t, v1.Equal(v3))
		assert.False(t, v3.Equal(v1))
	})
}
