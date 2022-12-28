package valuer

import (
	"testing"

	"github.com/jison/uni/internal/errors"
	"github.com/stretchr/testify/assert"
)

func Test_identityValuer_Value(t *testing.T) {

	t.Run("no error input", func(t *testing.T) {
		valuer := Identity()
		inputs := ValuesOf(
			123,
		)
		res := valuer.Value(inputs)
		rVal, ok := res.AsSingle()

		assert.True(t, ok)
		assert.Equal(t, 123, rVal.Interface())
	})

	t.Run("error input", func(t *testing.T) {
		valuer := Identity()
		inputs := ValuesOf(
			errors.Newf("abc"),
		)
		res := valuer.Value(inputs)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("identity != identity", func(t *testing.T) {
		v1 := Identity()
		v2 := Identity()
		assert.True(t, v1 != v2)
	})
}

func Test_identityValuer_String(t *testing.T) {
	valuer := Identity()
	assert.Equal(t, "identity", valuer.String())
}

func Test_identityValuer_Clone(t *testing.T) {
	v1 := Identity()
	v2 := v1.Clone()

	assert.False(t, v1 == v2)
	assert.Equal(t, v1, v2)
	assert.True(t, v1.Equal(v2))
}

func Test_identityValuer_Equal(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		v1 := Identity()
		v2 := Identity()
		assert.True(t, v1.Equal(v2))
	})

	t.Run("not equal", func(t *testing.T) {
		v1 := Identity()
		v2 := OneOf()
		assert.False(t, v1.Equal(v2))
	})

	t.Run("nil", func(t *testing.T) {
		var v1 *identityValuer
		var v2 *identityValuer
		var v3 = &identityValuer{}
		assert.True(t, v1.Equal(v2))
		assert.True(t, v1.Equal(v3))
		assert.True(t, v3.Equal(v1))
		assert.True(t, v1.Equal(nil))
	})
}
