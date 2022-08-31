package valuer

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstValuer(t *testing.T) {
	t.Run("value", func(t *testing.T) {
		valuer := Const(reflect.ValueOf("abc"))
		res := valuer.Value(nil)

		rVal, ok := res.AsSingle()
		assert.True(t, ok)
		assert.Equal(t, "abc", rVal.Interface())
	})

	t.Run("String", func(t *testing.T) {
		valuer := Const(reflect.ValueOf("abc"))
		assert.Equal(t, "Const: string(abc)", valuer.String())
	})

	t.Run("Clone", func(t *testing.T) {
		v1 := Const(reflect.ValueOf("abc"))
		v2 := v1.Clone()

		assert.False(t, v1 == v2)
		assert.Equal(t, v1, v2)
		assert.True(t, v1.Equal(v2))
	})

	t.Run("Equal", func(t *testing.T) {
		t.Run("equal", func(t *testing.T) {
			v1 := Const(reflect.ValueOf("abc"))
			v2 := Const(reflect.ValueOf("abc"))
			assert.True(t, v1.Equal(v2))
		})

		t.Run("not equal", func(t *testing.T) {
			v1 := Const(reflect.ValueOf("abc"))
			v2 := Const(reflect.ValueOf("def"))
			assert.False(t, v1.Equal(v2))
		})

		t.Run("nil", func(t *testing.T) {
			var v1 *constValuer
			var v2 *constValuer
			var v3 = Const(reflect.ValueOf("abc"))
			assert.True(t, v1.Equal(v2))
			assert.False(t, v1.Equal(v3))
			assert.False(t, v3.Equal(v1))
		})
	})
}
