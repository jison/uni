package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewScope(t *testing.T) {
	t.Run("NewScope", func(t *testing.T) {
		s1 := NewScope("scope1")
		s2 := NewScope("scope2", s1)
		assert.NotNil(t, s1)
		assert.NotNil(t, s2)

		assert.True(t, s2.CanEnterFrom(s1))
		assert.False(t, s1.CanEnterFrom(s2))
	})
}

func TestScope(t *testing.T) {
	t.Run("attributes", func(t *testing.T) {
		s1 := NewScope("scope1")
		s2 := NewScope("scope2")
		assert.NotEqual(t, s1, s2)
		assert.NotEqual(t, s1.ID(), s2.ID())

		assert.Equal(t, "scope1", s1.Name())
		assert.Equal(t, "scope2", s2.Name())

		assert.False(t, s1.CanEnterFrom(s2))
		assert.False(t, s2.CanEnterFrom(s1))
	})
}

func TestScope_canEnterFrom(t *testing.T) {
	t.Run("CanEnterFrom", func(t *testing.T) {
		s1 := NewScope("scope1")
		s2 := NewScope("scope2", s1)
		assert.NotNil(t, s1)
		assert.NotNil(t, s2)

		assert.False(t, s1.CanEnterFrom(s1))
		assert.False(t, s2.CanEnterFrom(s2))

		assert.True(t, s2.CanEnterFrom(s1))
		assert.False(t, s1.CanEnterFrom(s2))

		assert.True(t, s1.CanEnterFrom(GlobalScope))
		assert.True(t, s2.CanEnterFrom(GlobalScope))
		assert.False(t, GlobalScope.CanEnterFrom(s1))
		assert.False(t, GlobalScope.CanEnterFrom(s2))

		s3 := NewScope("scope3", s1, s2)
		assert.True(t, s3.CanEnterFrom(s1))
		assert.True(t, s3.CanEnterFrom(s2))
		assert.True(t, s3.CanEnterFrom(GlobalScope))

		s4 := scope{
			name:    "scope4",
			id:      nil,
			loc:     nil,
			parents: nil,
		}
		assert.False(t, s4.CanEnterFrom(s3))
	})
}

func TestScope_canEnterDirectlyFrom(t *testing.T) {
	t.Run("CanEnterDirectlyFrom", func(t *testing.T) {
		s1 := NewScope("scope1")
		s2 := NewScope("scope2", s1)
		s3 := NewScope("scope3", s2)

		assert.False(t, s1.CanEnterDirectlyFrom(s1))
		assert.False(t, s2.CanEnterDirectlyFrom(s2))
		assert.False(t, s3.CanEnterDirectlyFrom(s3))

		assert.False(t, s3.CanEnterDirectlyFrom(s1))

		assert.True(t, s2.CanEnterDirectlyFrom(s1))
		assert.True(t, s3.CanEnterDirectlyFrom(s2))
	})
}

func TestScope_Format(t *testing.T) {
	t.Run("formatting", func(t *testing.T) {
		s1 := NewScope("scope1")
		assert.Equal(t, "scope1", fmt.Sprintf("%v", s1))
		assert.Equal(t, "scope1", fmt.Sprintf("%s", s1))
		assert.Equal(t, "github.com/jison/uni/core/model.TestScope_Format.func1.scope1",
			fmt.Sprintf("%+v", s1))
	})
}

func TestGlobalScope(t *testing.T) {
	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, "Global", GlobalScope.Name())
	})

	t.Run("ID", func(t *testing.T) {
		assert.Equal(t, GlobalScope, GlobalScope.ID())
	})

	t.Run("CanEnterFrom", func(t *testing.T) {
		s1 := NewScope("scope1")
		s2 := NewScope("scope2", s1)
		s3 := NewScope("scope3", s2)

		assert.False(t, GlobalScope.CanEnterFrom(s1))
		assert.False(t, GlobalScope.CanEnterFrom(s2))
		assert.False(t, GlobalScope.CanEnterFrom(s3))
	})

	t.Run("CanEnterDirectlyFrom", func(t *testing.T) {
		s1 := NewScope("scope1")
		s2 := NewScope("scope2", s1)
		s3 := NewScope("scope3", s2)

		assert.False(t, GlobalScope.CanEnterDirectlyFrom(s1))
		assert.False(t, GlobalScope.CanEnterDirectlyFrom(s2))
		assert.False(t, GlobalScope.CanEnterDirectlyFrom(s3))
	})

	t.Run("Format", func(t *testing.T) {
		assert.Equal(t, "Global", fmt.Sprintf("%v", GlobalScope))
	})
}
