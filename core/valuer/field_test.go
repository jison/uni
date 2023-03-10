package valuer

import (
	"testing"

	"github.com/jison/uni/internal/errors"
	"github.com/stretchr/testify/assert"
)

func Test_fieldValuer_Value(t *testing.T) {
	t.Run("no error input", func(t *testing.T) {
		valuer := Field("abc")
		inputs := ValuesOf(
			123,
		)
		res := valuer.Value(inputs)
		rVal, ok := res.AsSingle()
		assert.True(t, ok)
		field := rVal.Interface().(structField)
		assert.Equal(t, "abc", field.name)
		assert.Equal(t, 123, field.val.Interface())
	})

	t.Run("error input", func(t *testing.T) {
		valuer := Field("abc")
		err := errors.Newf("this is error")
		inputs := ValuesOf(
			err,
		)
		res := valuer.Value(inputs)
		err2, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err2)
		assert.Equal(t, err2, err)
	})

	t.Run("invalid input", func(t *testing.T) {
		valuer := Field("abc")
		inputs := ValuesOf(
			[]int{1, 2, 3},
		)
		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})
}

func Test_fieldValuer_String(t *testing.T) {
	valuer := Field("abc")
	assert.Equal(t, "Field: abc", valuer.String())
}

func Test_fieldValuer_Clone(t *testing.T) {
	v1 := Field("abc")
	v2 := v1.Clone()

	assert.False(t, v1 == v2)
	assert.Equal(t, v1, v2)
	assert.True(t, v1.Equal(v2))
}

func Test_fieldValuer_Equal(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		v1 := Field("abc")
		v2 := Field("abc")
		assert.True(t, v1.Equal(v2))
	})

	t.Run("not equal", func(t *testing.T) {
		v1 := Field("abc")
		v2 := Field("def")
		v3 := Identity()
		assert.False(t, v1.Equal(v2))
		assert.False(t, v1.Equal(v3))
	})

	t.Run("nil", func(t *testing.T) {
		var v1 *structFieldValuer
		var v2 *structFieldValuer
		var v3 = &structFieldValuer{fieldName: "abc"}
		assert.True(t, v1.Equal(v2))
		assert.False(t, v1.Equal(v3))
		assert.False(t, v3.Equal(v1))
	})
}
