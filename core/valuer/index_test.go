package valuer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexValuer(t *testing.T) {
	t.Run("valid index", func(t *testing.T) {
		valuer := Index(2)
		inputs := ValuesOf(
			[]string{"abc", "def", "ghi"},
		)
		res := valuer.Value(inputs)
		rVal, ok := res.AsSingle()

		assert.True(t, ok)
		assert.Equal(t, "ghi", rVal.Interface())
	})

	t.Run("invalid index", func(t *testing.T) {
		valuer := Index(3)
		inputs := ValuesOf(
			[]string{"abc", "def", "ghi"},
		)
		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("invalid index 2", func(t *testing.T) {
		valuer := Index(-2)
		inputs := ValuesOf(
			[]string{"abc", "def", "ghi"},
		)
		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("String", func(t *testing.T) {
		valuer := Index(2)
		assert.Equal(t, "Index: 2", valuer.String())
	})

	t.Run("Clone", func(t *testing.T) {
		v1 := Index(2)
		v2 := v1.Clone()

		assert.False(t, v1 == v2)
		assert.Equal(t, v1, v2)
		assert.True(t, v1.Equal(v2))
	})

	t.Run("Equal", func(t *testing.T) {
		t.Run("equal", func(t *testing.T) {
			v1 := Index(2)
			v2 := Index(2)
			assert.True(t, v1.Equal(v2))
		})

		t.Run("not equal", func(t *testing.T) {
			v1 := Index(2)
			v2 := Index(3)
			assert.False(t, v1.Equal(v2))
		})

		t.Run("nil", func(t *testing.T) {
			var v1 *indexValuer
			var v2 *indexValuer
			var v3 = &indexValuer{index: 1}
			assert.True(t, v1.Equal(v2))
			assert.False(t, v1.Equal(v3))
			assert.False(t, v3.Equal(v1))
		})
	})
}
