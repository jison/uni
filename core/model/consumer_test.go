package model

import (
	"reflect"
	"testing"

	"github.com/jison/uni/core/valuer"

	"github.com/jison/uni/internal/location"
	"github.com/stretchr/testify/assert"
)

func Test_baseConsumer(t *testing.T) {
	t.Run("Scope", func(t *testing.T) {
		t.Run("nil scope", func(t *testing.T) {
			bc := &baseConsumer{scope: nil}
			assert.Equal(t, GlobalScope, bc.Scope())
		})

		t.Run("scope", func(t *testing.T) {
			s := NewScope("scope")
			bc := &baseConsumer{scope: s}
			assert.Equal(t, s, bc.Scope())
		})
	})

	t.Run("SetScope", func(t *testing.T) {
		t.Run("nil scope", func(t *testing.T) {
			bc := &baseConsumer{scope: nil}
			bc.SetScope(nil)
			assert.Equal(t, GlobalScope, bc.Scope())
		})

		t.Run("scope", func(t *testing.T) {
			s := NewScope("scope")
			bc := &baseConsumer{scope: s}
			bc.SetScope(s)
			assert.Equal(t, s, bc.Scope())
		})
	})

	t.Run("Location", func(t *testing.T) {
		loc := location.GetCallLocation(0)
		bc := &baseConsumer{loc: loc}
		assert.Equal(t, loc, bc.Location())
	})

	t.Run("SetLocation", func(t *testing.T) {
		loc := location.GetCallLocation(0)
		bc := &baseConsumer{loc: loc}
		bc.SetLocation(loc)
		assert.Equal(t, loc, bc.Location())
	})

	t.Run("Valuer", func(t *testing.T) {
		bc := &baseConsumer{val: valuer.Const(reflect.ValueOf(123))}
		assert.Equal(t, valuer.Const(reflect.ValueOf(123)), bc.Valuer())
		assert.True(t, bc.Valuer() == bc.Valuer())
	})

	t.Run("clone", func(t *testing.T) {
		scope1 := NewScope("scope1")
		loc1 := location.GetCallLocation(0)
		val1 := valuer.Const(reflect.ValueOf(123))
		bc := &baseConsumer{
			scope: scope1,
			loc:   loc1,
			val:   val1,
		}

		verifyConsumer := func(t *testing.T, con *baseConsumer) {
			assert.Equal(t, scope1, bc.Scope())
			assert.Equal(t, loc1, bc.Location())
			assert.Equal(t, val1, bc.Valuer())
		}

		t.Run("equality", func(t *testing.T) {
			bc2 := bc.clone()
			verifyConsumer(t, bc2)
			assert.False(t, bc2.Valuer() == bc.Valuer())
		})

		t.Run("update isolation", func(t *testing.T) {
			scope2 := NewScope("scope2")
			loc2 := location.GetCallLocation(0)
			val2 := valuer.Const(reflect.ValueOf(456))

			bc2 := bc.clone()
			bc2.scope = scope2
			bc2.loc = loc2
			bc2.val = val2
			verifyConsumer(t, bc)
		})
	})

	t.Run("equal", func(t *testing.T) {
		scope1 := NewScope("scope1")
		loc1 := location.GetCallLocation(0)
		val1 := valuer.Const(reflect.ValueOf(123))
		c := &baseConsumer{
			scope: scope1,
			loc:   loc1,
			val:   val1,
		}

		t.Run("equal", func(t *testing.T) {
			c2 := c.clone()
			assert.True(t, c.Equal(c2))
		})

		t.Run("scope", func(t *testing.T) {
			scope2 := NewScope("scope2")

			c2 := c.clone()
			c2.SetScope(scope2)
			assert.False(t, c.Equal(c2))
		})

		t.Run("location", func(t *testing.T) {
			loc2 := location.GetCallLocation(0)

			c2 := c.clone()
			c2.SetLocation(loc2)
			assert.False(t, c.Equal(c2))
		})

		t.Run("valuer", func(t *testing.T) {
			c2 := c.clone()
			c2.val = nil
			assert.False(t, c2.Equal(c))
		})
	})
}
