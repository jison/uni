package valuer

import (
	"testing"

	"github.com/jison/uni/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestParamValuer(t *testing.T) {
	t.Run("no error input", func(t *testing.T) {
		valuer := Param(2)
		inputs := ValuesOf(
			123,
		)
		res := valuer.Value(inputs)
		rVal, ok := res.AsSingle()
		assert.True(t, ok)
		param := rVal.Interface().(funcParam)
		assert.Equal(t, 2, param.index)
		assert.Equal(t, 123, param.val.Interface())

	})

	t.Run("error input", func(t *testing.T) {
		valuer := Param(2)
		err := errors.Newf("this is error")
		inputs := ValuesOf(
			err,
		)
		res := valuer.Value(inputs)
		err2, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err2)
		assert.Equal(t, err, err2)
	})

	t.Run("String", func(t *testing.T) {
		valuer := Param(2)
		assert.Equal(t, "Param: 2", valuer.String())
	})

	t.Run("Clone", func(t *testing.T) {
		v1 := Param(2)
		v2 := v1.Clone()

		assert.False(t, v1 == v2)
		assert.Equal(t, v1, v2)
		assert.True(t, v1.Equal(v2))
	})

	t.Run("Equal", func(t *testing.T) {
		t.Run("equal", func(t *testing.T) {
			v1 := Param(2)
			v2 := Param(2)
			assert.True(t, v1.Equal(v2))
		})

		t.Run("not equal", func(t *testing.T) {
			v1 := Param(2)
			v2 := Param(3)
			assert.False(t, v1.Equal(v2))
		})

		t.Run("nil", func(t *testing.T) {
			var v1 *paramValuer
			var v2 *paramValuer
			var v3 = &paramValuer{index: 2}
			assert.True(t, v1.Equal(v2))
			assert.False(t, v1.Equal(v3))
			assert.False(t, v3.Equal(v1))
		})
	})
}
