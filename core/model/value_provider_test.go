package model

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jison/uni/core/valuer"
	"github.com/jison/uni/internal/errors"
	"github.com/jison/uni/internal/location"
	"github.com/stretchr/testify/assert"
)

func TestValue(t *testing.T) {
	t.Run("normal value", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		scope1 := NewScope("scope1")
		baseLoc := location.GetCallLocation(0)
		vp := Value(123,
			Name("abc"),
			Tags(tag1),
			As(TypeOf(1)),
			Ignore(),
			Hide(),
			InScope(scope1),
			nil,
		)
		p := vp.Provider()

		deps := dependencyIteratorToArray(p.Dependencies())
		assert.Equal(t, 0, len(deps))

		assert.Equal(t, valuer.Const(reflect.ValueOf(123)), p.Valuer())

		assert.Equal(t, baseLoc.FileName(), p.Location().FileName())
		assert.Equal(t, baseLoc.FileLine()+1, p.Location().FileLine())
		assert.Equal(t, scope1, p.Scope())

		coms := p.Components().ToArray()
		assert.Equal(t, 1, len(coms))
		com := coms[0]

		assert.Equal(t, TypeOf(123), com.Type())
		assert.Equal(t, "abc", com.Name())
		assert.Equal(t, newSymbolSet(tag1), com.Tags())
		assert.Equal(t, true, com.Ignored())
		assert.Equal(t, true, com.Hidden())
		assert.Equal(t, newTypeSet(TypeOf(1)), com.As())
		assert.Equal(t, valuer.Identity(), com.Valuer())
		assert.Same(t, p, com.Provider())
	})

	t.Run("nil value", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		scope1 := NewScope("scope1")
		vp := Value(nil,
			Name("abc"),
			Tags(tag1),
			As(TypeOf(1)),
			Ignore(),
			Hide(),
			InScope(scope1),
			nil,
		)
		p := vp.Provider()
		err := p.Validate()
		assert.NotNil(t, err)
	})
}

func Test_valueProvider_Provider(t *testing.T) {
	t.Run("Dependencies", func(t *testing.T) {
		vp := Value(123, Name("abc"))
		p := vp.Provider()

		deps := dependencyIteratorToArray(p.Dependencies())
		assert.Equal(t, 0, len(deps))

		for _, dep := range deps {
			assert.True(t, dep.Valuer() == dep.Valuer())
		}
	})

	t.Run("Valuer", func(t *testing.T) {
		vp := Value(123)
		p := vp.Provider()

		assert.Equal(t, valuer.Const(reflect.ValueOf(123)), p.Valuer())
		assert.True(t, p.Valuer() == p.Valuer())
	})

	t.Run("Location", func(t *testing.T) {
		baseLoc := location.GetCallLocation(0)
		vp := Value(123, Name("abc"))
		p := vp.Provider()

		assert.Equal(t, baseLoc.FileName(), p.Location().FileName())
		assert.Equal(t, baseLoc.FileLine()+1, p.Location().FileLine())
	})

	t.Run("Components", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		scope1 := NewScope("scope1")
		vp := Value(123, Name("abc"), Tags(tag1), As(TypeOf(1)), Ignore(), Hide(), InScope(scope1))
		p := vp.Provider()

		coms := p.Components().ToArray()
		assert.Equal(t, 1, len(coms))
		com := coms[0]

		assert.Equal(t, TypeOf(123), com.Type())
		assert.Equal(t, "abc", com.Name())
		assert.Equal(t, newSymbolSet(tag1), com.Tags())
		assert.Equal(t, true, com.Ignored())
		assert.Equal(t, true, com.Hidden())
		assert.Equal(t, newTypeSet(TypeOf(1)), com.As())
		assert.Equal(t, valuer.Identity(), com.Valuer())
		assert.True(t, com.Valuer() == com.Valuer())
		assert.Same(t, p, com.Provider())
	})

	t.Run("Scope", func(t *testing.T) {
		t.Run("set scope", func(t *testing.T) {
			scope1 := NewScope("scope1")
			vp := Value(123, InScope(scope1))
			p := vp.Provider()

			assert.Same(t, scope1, p.Scope())
		})

		t.Run("set nil scope", func(t *testing.T) {
			vp := Value(123, InScope(nil))
			p := vp.Provider()

			assert.Same(t, GlobalScope, p.Scope())
		})

		t.Run("have no set scope", func(t *testing.T) {
			vp := Value(123)
			p := vp.Provider()

			assert.Same(t, GlobalScope, p.Scope())
		})
	})

	t.Run("Validate", func(t *testing.T) {
		t.Run("no error", func(t *testing.T) {
			tag1 := NewSymbol("tag1")
			scope1 := NewScope("scope1")
			vp := Value(123, Name("abc"), Tags(tag1), Ignore(), Hide(), InScope(scope1))
			p := vp.Provider()
			err := p.Validate()
			assert.Nil(t, err)
		})

		t.Run("value CanInterface() is false", func(t *testing.T) {
			// can not construct a value CanInterface() is false, and pass to Value
		})

		t.Run("value is error", func(t *testing.T) {
			tag1 := NewSymbol("tag1")
			scope1 := NewScope("scope1")
			vp := Value(errors.Newf("this is an error"), Name("abc"), Tags(tag1), InScope(scope1))
			p := vp.Provider()
			err := p.Validate()
			assert.NotNil(t, err)
		})

		t.Run("error in component", func(t *testing.T) {
			t.Run("as a non interface type", func(t *testing.T) {
				tag1 := NewSymbol("tag1")
				scope1 := NewScope("scope1")
				vp := Value(123, Name("abc"), Tags(tag1), As(TypeOf(1)), Ignore(), Hide(), InScope(scope1))
				p := vp.Provider()
				err := p.Validate()
				assert.NotNil(t, err)
			})
		})
	})

	t.Run("Format", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		scope1 := NewScope("scope1")
		vp := Value(123, Name("abc"), Tags(tag1), As(TypeOf(1)), Ignore(), Hide(), InScope(scope1))
		p := vp.Provider()

		t.Run("not verbose", func(t *testing.T) {
			expected := fmt.Sprintf("Value[%v](%v) in %v", TypeOf(123), 123, p.Scope())
			assert.Equal(t, expected, fmt.Sprintf("%v", p))
		})

		t.Run("verbose", func(t *testing.T) {
			expected := fmt.Sprintf("Value[%v](%v) in %v at %v", TypeOf(123), 123, p.Scope(),
				p.Location())
			assert.Equal(t, expected, fmt.Sprintf("%+v", p))
		})
	})
}

func Test_valueProvider_ValueProviderBuilder(t *testing.T) {
	t.Run("ApplyModule", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		scope1 := NewScope("scope1")
		mb := NewModuleBuilder()
		vp := valueProviderOf(123).
			SetName("abc").
			AddTags(tag1).
			AddAs(TypeOf(1)).
			SetIgnore(true).
			SetHidden(true).
			SetScope(scope1)
		vp.ApplyModule(mb)

		coms := mb.Module().AllComponents().ToArray()
		assert.Equal(t, 1, len(coms))
		com := coms[0]

		assert.Equal(t, TypeOf(123), com.Type())
		assert.Equal(t, "abc", com.Name())
		assert.Equal(t, newSymbolSet(tag1), com.Tags())
		assert.Equal(t, newTypeSet(TypeOf(1)), com.As())
		assert.Equal(t, true, com.Ignored())
		assert.Equal(t, true, com.Hidden())
	})

	t.Run("Provide", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		scope1 := NewScope("scope1")
		baseLoc := location.GetCallLocation(0)
		vp := Value(123).
			SetName("abc").
			AddTags(tag1).
			AddAs(TypeOf(1)).
			SetIgnore(true).
			SetHidden(true).
			SetScope(scope1)
		p := vp.Provider()

		deps := dependencyIteratorToArray(p.Dependencies())
		assert.Equal(t, 0, len(deps))

		assert.Equal(t, valuer.Const(reflect.ValueOf(123)), p.Valuer())

		assert.Equal(t, baseLoc.FileName(), p.Location().FileName())
		assert.Equal(t, baseLoc.FileLine()+1, p.Location().FileLine())
		assert.Equal(t, scope1, p.Scope())

		coms := p.Components().ToArray()
		assert.Equal(t, 1, len(coms))
		com := coms[0]

		assert.Equal(t, TypeOf(123), com.Type())
		assert.Equal(t, "abc", com.Name())
		assert.Equal(t, newSymbolSet(tag1), com.Tags())
		assert.Equal(t, true, com.Ignored())
		assert.Equal(t, true, com.Hidden())
	})

	t.Run("SetIgnore", func(t *testing.T) {
		vp := Value(123)
		vp.SetIgnore(true)
		com := vp.Provider().Components().ToArray()[0]
		assert.True(t, com.Ignored())
	})

	t.Run("SetHidden", func(t *testing.T) {
		vp := Value(123)
		vp.SetHidden(true)
		com := vp.Provider().Components().ToArray()[0]
		assert.True(t, com.Hidden())
	})

	t.Run("AddAs", func(t *testing.T) {
		vp := Value(123)
		vp.AddAs(TypeOf(1))
		com := vp.Provider().Components().ToArray()[0]
		assert.True(t, com.As().Has(TypeOf(1)))
	})

	t.Run("SetName", func(t *testing.T) {
		vp := Value(123)
		vp.SetName("abc")
		com := vp.Provider().Components().ToArray()[0]
		assert.Equal(t, "abc", com.Name())
	})

	t.Run("AddTags", func(t *testing.T) {
		tag1 := NewSymbol("tag1")

		vp := Value(123)
		vp.AddTags(tag1)
		com := vp.Provider().Components().ToArray()[0]
		assert.True(t, com.Tags().Has(tag1))
	})

	t.Run("SetScope", func(t *testing.T) {
		t.Run("set scope", func(t *testing.T) {
			scope1 := NewScope("scope1")
			vp := Value(123)
			vp.SetScope(scope1)
			p := vp.Provider()
			assert.Equal(t, scope1, p.Scope())
		})

		t.Run("set nil scope", func(t *testing.T) {
			vp := Value(123)
			vp.SetScope(nil)
			p := vp.Provider()
			assert.Equal(t, GlobalScope, p.Scope())
		})
	})

	t.Run("SetLocation", func(t *testing.T) {
		loc := location.GetCallLocation(0)
		vp := Value(123)
		vp.SetLocation(loc)
		p := vp.Provider()
		assert.Equal(t, loc, p.Location())
	})

	t.Run("UpdateCallLocation", func(t *testing.T) {
		t.Run("location have been set", func(t *testing.T) {
			loc1 := location.GetCallLocation(0)
			vp := valueProviderOf(123)
			vp.SetLocation(loc1)
			vp.UpdateCallLocation(nil)
			assert.Equal(t, loc1, vp.Location())
		})

		t.Run("location have not been set", func(t *testing.T) {
			vp := valueProviderOf(123)
			baseLoc := location.GetCallLocation(0)
			func() {
				vp.UpdateCallLocation(nil)
			}()
			assert.Equal(t, baseLoc.FileName(), vp.Location().FileName())
			assert.Equal(t, baseLoc.FileLine()+3, vp.Location().FileLine())
		})

		t.Run("location is not nil", func(t *testing.T) {
			loc1 := location.GetCallLocation(0)
			vp := valueProviderOf(123)
			vp.UpdateCallLocation(loc1)
			assert.Equal(t, loc1, vp.Location())
		})
	})
}

func Test_valueProvider_clone(t *testing.T) {
	tag1 := NewSymbol("tag1")
	scope1 := NewScope("scope1")
	loc := location.GetCallLocation(0)
	vp := valueProviderOf(123)

	vp.SetName("abc").
		AddTags(tag1).
		AddAs(TypeOf(1)).
		SetIgnore(true).
		SetHidden(true).
		SetScope(scope1).
		SetLocation(loc)

	verifyProvider := func(t *testing.T, pro Provider) {
		assert.Equal(t, scope1, pro.Scope())
		assert.Equal(t, loc, pro.Location())
		com := pro.Components().ToArray()[0]

		assert.Equal(t, true, com.Ignored())
		assert.Equal(t, true, com.Hidden())
		assert.Equal(t, "abc", com.Name())
		assert.Equal(t, newSymbolSet(tag1), com.Tags())
		assert.Equal(t, newTypeSet(TypeOf(1)), com.As())
		assert.Same(t, pro, com.Provider())
	}

	t.Run("equality", func(t *testing.T) {
		vp2 := vp.clone()
		verifyProvider(t, vp2.Provider())

		assert.False(t, vp2.Valuer() == vp.Valuer())
	})

	t.Run("update isolation", func(t *testing.T) {
		tag2 := NewSymbol("tag2")
		scope2 := NewScope("scope2")
		loc2 := location.GetCallLocation(0)

		vp2 := vp.clone()
		vp2.SetName("def").
			AddTags(tag2).
			AddAs(TypeOf("a")).
			SetIgnore(false).
			SetHidden(false).
			SetScope(scope2).
			SetLocation(loc2)

		verifyProvider(t, vp.Provider())
	})

	t.Run("update isolation 2", func(t *testing.T) {
		tag2 := NewSymbol("tag2")
		scope2 := NewScope("scope2")
		loc2 := location.GetCallLocation(0)

		vp2 := vp.clone()
		vp3 := vp2.clone()
		vp2.SetName("def").
			AddTags(tag2).
			AddAs(TypeOf("a")).
			SetIgnore(false).
			SetHidden(false).
			SetScope(scope2).
			SetLocation(loc2)

		verifyProvider(t, vp3.Provider())
	})
}

func Test_valuerProvider_Equal(t *testing.T) {
	tag1 := NewSymbol("tag1")
	scope1 := NewScope("scope1")
	loc := location.GetCallLocation(0)
	vp := valueProviderOf(123)

	vp.SetName("abc").
		AddTags(tag1).
		AddAs(TypeOf(1)).
		SetIgnore(true).
		SetHidden(true).
		SetScope(scope1).
		SetLocation(loc)

	t.Run("equal", func(t *testing.T) {
		vp2 := vp.clone()
		assert.True(t, vp2.Equal(vp))
	})

	t.Run("baseConsumer", func(t *testing.T) {
		vp2 := vp.clone()
		vp2.SetLocation(location.GetCallLocation(0))
		assert.False(t, vp2.Equal(vp))
	})

	t.Run("value", func(t *testing.T) {
		vp2 := vp.clone()
		vp2.value = reflect.ValueOf(456)
		assert.False(t, vp2.Equal(vp))
	})

	t.Run("component", func(t *testing.T) {
		vp2 := vp.clone()
		vp2.SetName("def")
		assert.False(t, vp2.Equal(vp))
	})
}
