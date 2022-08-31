package model

import (
	"fmt"
	"testing"

	"github.com/jison/uni/core/valuer"
	"github.com/jison/uni/internal/location"
	"github.com/stretchr/testify/assert"
)

func TestValueConsumer(t *testing.T) {
	t.Run("normal type", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		baseLoc := location.GetCallLocation(0)
		vc := ValueConsumer(
			TypeOf([]int{}),
			ByName("abc"),
			ByTags(tag1),
			Optional(true),
			AsCollector(true),
			nil,
		)
		con := vc.Consumer()

		deps := dependencyIteratorToArray(con.Dependencies())
		assert.Equal(t, 1, len(deps))

		dep := deps[0]
		assert.Equal(t, TypeOf(0), dep.Type())
		assert.Equal(t, "abc", dep.Name())
		assert.Equal(t, newSymbolSet(tag1), dep.Tags())
		assert.Equal(t, true, dep.Optional())
		assert.Equal(t, true, dep.IsCollector())
		assert.Equal(t, valuer.Identity(), dep.Valuer())
		assert.True(t, dep.Valuer() == dep.Valuer())

		assert.Same(t, con, dep.Consumer())

		assert.Equal(t, valuer.Identity(), con.Valuer())

		assert.Equal(t, baseLoc.FileName(), con.Location().FileName())
		assert.Equal(t, baseLoc.FileLine()+1, con.Location().FileLine())
	})

	t.Run("nil type", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		vc := ValueConsumer(
			nil,
			ByName("abc"),
			ByTags(tag1),
			Optional(true),
			AsCollector(true),
			nil,
		)
		con := vc.Consumer()
		err := con.Validate()
		assert.NotNil(t, err)
	})
}

func Test_valueConsumer_Consumer(t *testing.T) {
	t.Run("Dependencies", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		vc := ValueConsumer(TypeOf(1), ByName("abc"), ByTags(tag1), Optional(true), AsCollector(true))
		con := vc.Consumer()

		deps := dependencyIteratorToArray(con.Dependencies())
		assert.Equal(t, 1, len(deps))

		dep := deps[0]
		assert.Equal(t, TypeOf(1), dep.Type())
		assert.Equal(t, "abc", dep.Name())
		assert.Equal(t, newSymbolSet(tag1), dep.Tags())
		assert.Equal(t, true, dep.Optional())
		assert.Equal(t, true, dep.IsCollector())
		assert.Equal(t, valuer.Identity(), dep.Valuer())
		assert.True(t, dep.Valuer() == dep.Valuer())
		assert.Same(t, con, dep.Consumer())
	})

	t.Run("Valuer", func(t *testing.T) {
		vc := ValueConsumer(TypeOf(1))
		con := vc.Consumer()
		assert.Equal(t, valuer.Identity(), con.Valuer())
	})

	t.Run("Scope", func(t *testing.T) {
		vc := ValueConsumer(TypeOf(1))
		con := vc.Consumer()
		assert.Equal(t, GlobalScope, con.Scope())

		scope1 := NewScope("scope1")

		vc2 := ValueConsumer(TypeOf(1), InScope(scope1))
		con2 := vc2.Consumer()
		assert.Equal(t, scope1, con2.Scope())
	})

	t.Run("Location", func(t *testing.T) {
		baseLoc := location.GetCallLocation(0)
		vc := ValueConsumer(TypeOf(1))
		con := vc.Consumer()
		assert.Equal(t, baseLoc.FileName(), con.Location().FileName())
		assert.Equal(t, baseLoc.FileLine()+1, con.Location().FileLine())
	})

	t.Run("Validate", func(t *testing.T) {
		t.Run("not error", func(t *testing.T) {
			vc := ValueConsumer(TypeOf(1), AsCollector(false))
			con := vc.Consumer()
			deps := dependencyIteratorToArray(con.Dependencies())
			dep := deps[0]

			err := dep.Validate()
			assert.Nil(t, err)
		})

		t.Run("with nil type", func(t *testing.T) {
			vc := ValueConsumer(TypeOf(nil), AsCollector(false))
			con := vc.Consumer()
			deps := dependencyIteratorToArray(con.Dependencies())
			dep := deps[0]

			err := dep.Validate()
			assert.NotNil(t, err)
		})

		t.Run("with error type", func(t *testing.T) {
			vc := ValueConsumer(TypeOf((*error)(nil)), AsCollector(false))
			con := vc.Consumer()
			deps := dependencyIteratorToArray(con.Dependencies())
			dep := deps[0]

			err := dep.Validate()
			assert.NotNil(t, err)
		})

		t.Run("as collector while with no slice type", func(t *testing.T) {
			vc := ValueConsumer(TypeOf(1), AsCollector(true))
			con := vc.Consumer()
			deps := dependencyIteratorToArray(con.Dependencies())
			dep := deps[0]

			err := dep.Validate()
			assert.NotNil(t, err)
		})
	})

	t.Run("Format", func(t *testing.T) {
		vc := ValueConsumer(TypeOf(1))
		con := vc.Consumer()

		t.Run("not verbose", func(t *testing.T) {
			expected := fmt.Sprintf("ValueConsumer[%v]", TypeOf(1))
			assert.Equal(t, expected, fmt.Sprintf("%v", con))
		})

		t.Run("verbose", func(t *testing.T) {
			expected := fmt.Sprintf("ValueConsumer[%v] at %v", TypeOf(1), con.Location())
			assert.Equal(t, expected, fmt.Sprintf("%+v", con))
		})
	})
}

func Test_valueConsumer_ValueConsumerBuilder(t *testing.T) {
	t.Run("SetOptional", func(t *testing.T) {
		vc := ValueConsumer(TypeOf(1))
		vc.SetOptional(true)
		con := vc.Consumer()
		deps := dependencyIteratorToArray(con.Dependencies())
		dep := deps[0]

		assert.Equal(t, true, dep.Optional())
	})

	t.Run("SetAsCollector", func(t *testing.T) {
		vc := ValueConsumer(TypeOf(1))
		vc.SetAsCollector(true)
		con := vc.Consumer()
		deps := dependencyIteratorToArray(con.Dependencies())
		dep := deps[0]

		assert.Equal(t, true, dep.IsCollector())
	})

	t.Run("SetName", func(t *testing.T) {
		vc := ValueConsumer(TypeOf(1))
		vc.SetName("abc")
		con := vc.Consumer()
		deps := dependencyIteratorToArray(con.Dependencies())
		dep := deps[0]

		assert.Equal(t, "abc", dep.Name())
	})

	t.Run("AddTags", func(t *testing.T) {
		tag1 := NewSymbol("tag1")

		vc := ValueConsumer(TypeOf(1))
		vc.AddTags(tag1)
		con := vc.Consumer()
		deps := dependencyIteratorToArray(con.Dependencies())
		dep := deps[0]

		assert.Equal(t, newSymbolSet(tag1), dep.Tags())
	})

	t.Run("SetScope", func(t *testing.T) {
		t.Run("set scope", func(t *testing.T) {
			scope1 := NewScope("scope1")
			vc := ValueConsumer(TypeOf(1))
			vc.SetScope(scope1)
			con := vc.Consumer()
			assert.Equal(t, scope1, con.Scope())
		})

		t.Run("set nil scope", func(t *testing.T) {
			vc := ValueConsumer(TypeOf(1))
			vc.SetScope(nil)
			con := vc.Consumer()
			assert.Equal(t, GlobalScope, con.Scope())
		})
	})

	t.Run("SetLocation", func(t *testing.T) {
		loc := location.GetCallLocation(0)
		vc := ValueConsumer(TypeOf(1))
		vc.SetLocation(loc)
		con := vc.Consumer()
		assert.Equal(t, loc, con.Location())
	})

	t.Run("UpdateCallLocation", func(t *testing.T) {
		t.Run("location have been set", func(t *testing.T) {
			loc1 := location.GetCallLocation(0)
			vc := valueConsumerOf(TypeOf(1))
			vc.SetLocation(loc1)
			vc.UpdateCallLocation(nil)
			assert.Equal(t, loc1, vc.Location())
		})

		t.Run("location have not been set", func(t *testing.T) {
			vc := valueConsumerOf(TypeOf(1))
			baseLoc := location.GetCallLocation(0)
			func() {
				vc.UpdateCallLocation(nil)
			}()
			assert.Equal(t, baseLoc.FileName(), vc.Location().FileName())
			assert.Equal(t, baseLoc.FileLine()+3, vc.Location().FileLine())
		})

		t.Run("location is not nil", func(t *testing.T) {
			loc1 := location.GetCallLocation(0)
			vc := valueConsumerOf(TypeOf(1))
			vc.UpdateCallLocation(loc1)
			assert.Equal(t, loc1, vc.Location())
		})
	})

	t.Run("Consumer", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		baseLoc := location.GetCallLocation(0)
		vc := ValueConsumer(TypeOf(1)).
			SetName("abc").
			AddTags(tag1).
			SetOptional(true).
			SetAsCollector(true)
		con := vc.Consumer()

		deps := dependencyIteratorToArray(con.Dependencies())
		assert.Equal(t, 1, len(deps))

		dep := deps[0]
		assert.Equal(t, TypeOf(1), dep.Type())
		assert.Equal(t, "abc", dep.Name())
		assert.Equal(t, newSymbolSet(tag1), dep.Tags())
		assert.Equal(t, true, dep.Optional())
		assert.Equal(t, true, dep.IsCollector())
		assert.Equal(t, valuer.Identity(), dep.Valuer())
		assert.Same(t, con, dep.Consumer())

		assert.Equal(t, valuer.Identity(), con.Valuer())

		assert.Equal(t, baseLoc.FileName(), con.Location().FileName())
		assert.Equal(t, baseLoc.FileLine()+1, con.Location().FileLine())
	})
}

func Test_valueConsumer_clone(t *testing.T) {
	scope1 := NewScope("scope1")
	tag1 := NewSymbol("tag1")
	loc1 := location.GetCallLocation(0)
	vc := valueConsumerOf(TypeOf(1))
	vc.SetName("abc")
	vc.AddTags(tag1)
	vc.SetOptional(true)
	vc.SetAsCollector(true)
	vc.SetLocation(loc1)
	vc.SetScope(scope1)

	verifyConsumer := func(t *testing.T, con Consumer) {
		deps := dependencyIteratorToArray(con.Dependencies())
		assert.Equal(t, 1, len(deps))

		dep := deps[0]
		assert.Equal(t, TypeOf(1), dep.Type())
		assert.Equal(t, "abc", dep.Name())
		assert.Equal(t, newSymbolSet(tag1), dep.Tags())
		assert.Equal(t, true, dep.Optional())
		assert.Equal(t, true, dep.IsCollector())
		assert.Equal(t, valuer.Identity(), dep.Valuer())
		assert.Same(t, con, dep.Consumer())

		assert.Equal(t, valuer.Identity(), con.Valuer())
		assert.Equal(t, loc1, con.Location())
		assert.Equal(t, scope1, con.Scope())
	}

	t.Run("equality", func(t *testing.T) {
		vc2 := vc.clone()
		verifyConsumer(t, vc2.Consumer())

		assert.False(t, vc2.Valuer() == vc.Valuer())
	})

	t.Run("update isolation", func(t *testing.T) {
		scope2 := NewScope("scope2")
		loc2 := location.GetCallLocation(0)
		tag2 := NewSymbol("tag2")
		vc2 := vc.clone()
		vc2.SetName("bcd")
		vc2.AddTags(tag2)
		vc2.SetOptional(false)
		vc2.SetAsCollector(false)
		vc2.SetLocation(loc2)
		vc2.SetScope(scope2)

		verifyConsumer(t, vc.Consumer())
	})

	t.Run("update isolation 2", func(t *testing.T) {
		scope2 := NewScope("scope2")
		loc2 := location.GetCallLocation(0)
		tag2 := NewSymbol("tag2")
		vc2 := vc.clone()
		vc3 := vc2.clone()

		vc2.SetName("bcd")
		vc2.AddTags(tag2)
		vc2.SetOptional(false)
		vc2.SetAsCollector(false)
		vc2.SetLocation(loc2)
		vc2.SetScope(scope2)

		verifyConsumer(t, vc3.Consumer())
	})
}

func Test_valueConsumer_equal(t *testing.T) {
	scope1 := NewScope("scope1")
	tag1 := NewSymbol("tag1")
	loc1 := location.GetCallLocation(0)
	vc := valueConsumerOf(TypeOf(1))
	vc.SetName("abc")
	vc.AddTags(tag1)
	vc.SetOptional(true)
	vc.SetAsCollector(true)
	vc.SetLocation(loc1)
	vc.SetScope(scope1)

	t.Run("equal", func(t *testing.T) {
		vc2 := vc.clone()
		assert.True(t, vc2.Equal(vc))
	})

	t.Run("dependency", func(t *testing.T) {
		vc2 := vc.clone()
		vc2.SetName("def")
		assert.False(t, vc2.Equal(vc))
	})

	t.Run("baseConsumer", func(t *testing.T) {
		loc2 := location.GetCallLocation(0)

		vc2 := vc.clone()
		vc2.SetLocation(loc2)
		assert.False(t, vc2.Equal(vc))
	})
}
