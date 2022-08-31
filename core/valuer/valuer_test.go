package valuer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testValuer struct {
	_ int
}

func (v testValuer) ValueOne(input Value) Value {
	return input
}

func (v testValuer) String() string {
	return "test"
}

func (v testValuer) Clone() OneInputValuer {
	return testValuer{}
}

func (v testValuer) Equal(other interface{}) bool {
	_, ok := other.(testValuer)
	return ok
}

func TestOneInputValuer(t *testing.T) {
	t.Run("only one input", func(t *testing.T) {
		oneInput := &oneInputValuer{testValuer{}}

		inputs := ValuesOf("abc")
		res := oneInput.Value(inputs)

		_, isErr := res.AsError()
		assert.False(t, isErr)

		rVal, isSingle := res.AsSingle()
		assert.True(t, isSingle)
		assert.Equal(t, "abc", rVal.Interface())
		assert.Equal(t, "test", oneInput.String())
	})

	t.Run("multiple inputs", func(t *testing.T) {
		oneInput := &oneInputValuer{testValuer{}}

		inputs := ValuesOf("abc", 123)
		res := oneInput.Value(inputs)

		err, ok := res.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("clone", func(t *testing.T) {
		v1 := &oneInputValuer{testValuer{}}
		v2 := v1.Clone()

		assert.Equal(t, v1, v2)
		assert.False(t, v1 == v2)
		assert.True(t, v1.Equal(v2))
	})

	t.Run("equal", func(t *testing.T) {
		t.Run("equal", func(t *testing.T) {
			v1 := &oneInputValuer{testValuer{}}
			v2 := &oneInputValuer{testValuer{}}

			assert.True(t, v1.Equal(v2))
		})
		t.Run("not equal", func(t *testing.T) {
			v1 := &oneInputValuer{testValuer{}}
			v2 := &oneInputValuer{&identityValuer{}}

			assert.False(t, v1.Equal(v2))
		})
		t.Run("nil", func(t *testing.T) {
			var v1 *oneInputValuer
			var v2 *oneInputValuer
			var v3 = &oneInputValuer{testValuer{}}

			assert.True(t, v1.Equal(v2))
			assert.False(t, v1.Equal(v3))
			assert.False(t, v3.Equal(v1))
		})
	})
}
