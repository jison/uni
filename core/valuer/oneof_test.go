package valuer

import (
	"testing"

	"github.com/jison/uni/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestOneOfValuer(t *testing.T) {
	t.Run("one error input", func(t *testing.T) {
		valuer := OneOf()
		inputs := ValuesOf(
			errors.Newf("i am error"),
			123,
			"abc",
		)

		res := valuer.Value(inputs)
		rVal, ok := res.AsSingle()
		assert.True(t, ok)
		assert.Contains(t, []interface{}{123, "abc"}, rVal.Interface())

	})

	t.Run("all error inputs", func(t *testing.T) {
		valuer := OneOf()

		err1 := errors.Newf("i am error")
		err2 := errors.Newf("i am error2")
		inputs := ValuesOf(
			err1,
			err2,
		)

		res := valuer.Value(inputs)
		err3, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err3)
		assert.True(t, errors.Is(err3, err1) || errors.Is(err3, err2))
	})

	t.Run("empty inputs", func(t *testing.T) {
		valuer := OneOf()
		var inputs []Value

		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("oneOf != oneOf", func(t *testing.T) {
		v1 := OneOf()
		v2 := OneOf()
		assert.True(t, v1 != v2)
	})

	t.Run("String", func(t *testing.T) {
		valuer := OneOf()
		assert.Equal(t, "OneOf", valuer.String())
	})

	t.Run("Clone", func(t *testing.T) {
		v1 := OneOf()
		v2 := v1.Clone()

		assert.False(t, v1 == v2)
		assert.Equal(t, v1, v2)
		assert.True(t, v1.Equal(v2))
	})

	t.Run("Equal", func(t *testing.T) {
		t.Run("equal", func(t *testing.T) {
			v1 := OneOf()
			v2 := OneOf()
			assert.True(t, v1.Equal(v2))
		})

		t.Run("not equal", func(t *testing.T) {
			v1 := Identity()
			v2 := OneOf()
			assert.False(t, v1.Equal(v2))
		})

		t.Run("nil", func(t *testing.T) {
			var v1 *oneOfValuer
			var v2 *oneOfValuer
			var v3 = &oneOfValuer{}
			assert.True(t, v1.Equal(v2))
			assert.True(t, v1.Equal(v3))
			assert.True(t, v3.Equal(v1))
		})
	})
}
