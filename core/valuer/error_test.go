package valuer

import (
	"testing"

	"github.com/jison/uni/internal/errors"
	"github.com/stretchr/testify/assert"
)

func Test_errorValuer_Value(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		err := errors.Newf("this is error")
		valuer := Error(err)
		res := valuer.Value(nil)
		err2, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err2)
		assert.ErrorIs(t, err2, err)
	})

	t.Run("nil value", func(t *testing.T) {
		valuer := Error(nil)
		res := valuer.Value(nil)
		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})
}

func Test_errorValuer_String(t *testing.T) {
	err := errors.Newf("this is error")
	valuer := Error(err)
	assert.Equal(t, "Error: this is error", valuer.String())
}

func Test_errorValuer_Clone(t *testing.T) {
	err := errors.Newf("this is error")
	v1 := Error(err)
	v2 := v1.Clone()

	assert.False(t, v1 == v2)
	assert.Equal(t, v1, v2)
	assert.True(t, v1.Equal(v2))
}

func Test_errorValuer_Equal(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		err := errors.Newf("this is error")
		v1 := Error(err)
		v2 := Error(err)
		assert.True(t, v1.Equal(v2))
	})

	t.Run("not equal", func(t *testing.T) {
		err1 := errors.Newf("this is error1")
		err2 := errors.Newf("this is error2")
		v1 := Error(err1)
		v2 := Error(err2)
		assert.False(t, v1.Equal(v2))
	})

	t.Run("nil", func(t *testing.T) {
		var v1 *errorValuer
		var v2 *errorValuer
		var v3 = Error(errors.Newf("this is error"))
		assert.True(t, v1.Equal(v2))
		assert.False(t, v1.Equal(nil))
		assert.False(t, v1.Equal(v3))
		assert.False(t, v3.Equal(v1))
	})
}
