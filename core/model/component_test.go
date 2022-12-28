package model

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jison/uni/core/valuer"

	"github.com/jison/uni/internal/errors"
	"github.com/stretchr/testify/assert"
)

func newProviderForComponentTest() Provider {
	return Value(1).Provider()
}

func Test_component_attributes(t *testing.T) {
	tag1 := NewSymbol("tag1")

	provider := newProviderForComponentTest()

	t.Run("attributes", func(t *testing.T) {
		com := &component{
			provider: provider,
			ignored:  true,
			hidden:   true,
			rType:    TypeOf(0),
			as:       newTypeSet(TypeOf(1)),
			name:     "abc",
			tags:     newSymbolSet(tag1),
			val:      valuer.Identity(),
		}

		assert.Equal(t, provider, com.Provider())
		assert.Equal(t, com.ignored, com.Ignored())
		assert.Equal(t, com.hidden, com.Hidden())
		assert.Equal(t, com.rType, com.Type())
		assert.Equal(t, com.as, com.As())
		assert.Equal(t, com.name, com.Name())
		assert.Equal(t, com.tags, com.Tags())
		assert.Equal(t, com.val, com.Valuer())
	})

	t.Run("equal", func(t *testing.T) {
		com := &component{
			provider: provider,
			ignored:  true,
			hidden:   true,
			rType:    TypeOf(0),
			as:       newTypeSet(TypeOf(1)),
			name:     "abc",
			tags:     newSymbolSet(tag1),
			val:      valuer.Identity(),
		}

		t.Run("equal", func(t *testing.T) {
			com2 := com.clone()
			assert.True(t, com.Equal(com2))
		})

		t.Run("non component", func(t *testing.T) {
			assert.False(t, com.Equal(123))
		})

		t.Run("ignored", func(t *testing.T) {
			com2 := com.clone()
			com2.ignored = false
			assert.False(t, com.Equal(com2))
		})

		t.Run("hidden", func(t *testing.T) {
			com2 := com.clone()
			com2.hidden = false
			assert.False(t, com.Equal(com2))
		})

		t.Run("type", func(t *testing.T) {
			com2 := com.clone()
			com2.rType = reflect.TypeOf("")
			assert.False(t, com.Equal(com2))
		})

		t.Run("as", func(t *testing.T) {
			com2 := com.clone()
			com2.as.Add(TypeOf(""))
			assert.False(t, com.Equal(com2))
		})

		t.Run("name", func(t *testing.T) {
			com2 := com.clone()
			com2.name = "def"
			assert.False(t, com.Equal(com2))
		})

		t.Run("tags", func(t *testing.T) {
			tag2 := NewSymbol("tag2")
			com2 := com.clone()
			com.tags.Add(tag2)
			assert.False(t, com.Equal(com2))
		})

		t.Run("valuer", func(t *testing.T) {
			t.Run("not nil", func(t *testing.T) {
				com2 := com.clone()
				com2.val = valuer.Index(2)
				assert.False(t, com.Equal(com2))
			})

			t.Run("nil", func(t *testing.T) {
				com2 := com.clone()
				com2.val = nil
				assert.False(t, com2.Equal(com))
				assert.False(t, com.Equal(com2))

				com3 := com.clone()
				com3.val = nil
				assert.True(t, com2.Equal(com3))
			})
		})
	})
}

func Test_component_builder(t *testing.T) {
	provider := newProviderForComponentTest()
	com := &component{
		provider: provider,
	}

	t.Run("SetIgnore", func(t *testing.T) {
		com.SetIgnore(true)
		assert.True(t, com.Ignored())
		com.SetIgnore(false)
		assert.False(t, com.Ignored())
	})

	t.Run("SetHidden", func(t *testing.T) {
		com.SetHidden(true)
		assert.True(t, com.Hidden())
		com.SetHidden(false)
		assert.False(t, com.Hidden())
	})

	t.Run("AddAs", func(t *testing.T) {
		type testInterface interface {
			MethodA()
		}
		i := TypeOf((*testInterface)(nil))
		com.AddAs(i)
		assert.True(t, com.As().Has(i))
	})

	t.Run("SetName", func(t *testing.T) {
		com.SetName("abc")
		assert.Equal(t, "abc", com.Name())
		com.SetName("def")
		assert.Equal(t, "def", com.Name())
	})

	t.Run("AddTags", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		tag2 := NewSymbol("tag2")

		com.AddTags(tag1)
		assert.True(t, com.Tags().Has(tag1))
		com.AddTags(tag2)
		assert.True(t, com.Tags().Has(tag2))
	})

	t.Run("Component", func(t *testing.T) {
		com2 := com.Component()
		assert.True(t, com2.Equal(com))
	})
}

func Test_component_options(t *testing.T) {
	provider := newProviderForComponentTest()
	com := &component{
		provider: provider,
	}

	t.Run("Ignore", func(t *testing.T) {
		Ignore().ApplyComponent(com)
		assert.True(t, com.Ignored())
	})

	t.Run("Hide", func(t *testing.T) {
		Hide().ApplyComponent(com)
		assert.True(t, com.Hidden())
	})

	t.Run("As", func(t *testing.T) {
		type testInterface interface {
			MethodA()
		}

		i := TypeOf((*testInterface)(nil))
		As(i).ApplyComponent(com)
		assert.True(t, com.As().Has(i))
	})

	t.Run("Name", func(t *testing.T) {
		Name("abc").ApplyComponent(com)
		assert.Equal(t, "abc", com.Name())
		Name("def").ApplyComponent(com)
		assert.Equal(t, "def", com.Name())
	})

	t.Run("AddTags", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		tag2 := NewSymbol("tag2")

		Tags(tag1).ApplyComponent(com)
		assert.True(t, com.Tags().Has(tag1))
		Tags(tag2).ApplyComponent(com)
		assert.True(t, com.Tags().Has(tag2))
		Tags().ApplyComponent(com)
	})
}

func Test_component_Format(t *testing.T) {
	provider := newProviderForComponentTest()

	t.Run("type", func(t *testing.T) {
		com := &component{
			provider: provider,
			rType:    TypeOf(1),
		}
		s := fmt.Sprintf("%v", com)
		assert.Equal(t, "Component[int]", s)
		vs := fmt.Sprintf("%+v", com)
		assert.Equal(t, "Component[int]", vs)
	})

	t.Run("name", func(t *testing.T) {
		com := &component{
			provider: provider,
			rType:    TypeOf(1),
		}
		com.SetName("abc")
		s := fmt.Sprintf("%v", com)
		assert.Equal(t, "Component[int]{name=\"abc\"}", s)
		vs := fmt.Sprintf("%+v", com)
		assert.Equal(t, "Component[int]{name=\"abc\"}", vs)
	})

	t.Run("tags", func(t *testing.T) {
		com := &component{
			provider: provider,
			rType:    TypeOf(1),
		}

		tag1 := NewSymbol("tag1")
		tag2 := NewSymbol("tag2")

		com.AddTags(tag1, tag2)
		s := fmt.Sprintf("%v", com)
		assert.Equal(t, "Component[int]{tags={tag1, tag2}}", s)
		vs := fmt.Sprintf("%+v", com)
		vsExpected := fmt.Sprintf("Component[int]{tags=%+v}", com.tags)
		assert.Equal(t, vsExpected, vs)
	})

	t.Run("ignored", func(t *testing.T) {
		com := &component{
			provider: provider,
			rType:    TypeOf(1),
		}

		com.SetIgnore(true)
		s := fmt.Sprintf("%v", com)
		assert.Equal(t, "Component[int]{ignored}", s)
		vs := fmt.Sprintf("%+v", com)
		assert.Equal(t, "Component[int]{ignored}", vs)
	})

	t.Run("hidden", func(t *testing.T) {
		com := &component{
			provider: provider,
			rType:    TypeOf(1),
		}

		com.SetHidden(true)
		s := fmt.Sprintf("%v", com)
		assert.Equal(t, "Component[int]{hidden}", s)
		vs := fmt.Sprintf("%+v", com)
		assert.Equal(t, "Component[int]{hidden}", vs)
	})

	t.Run("as", func(t *testing.T) {
		com := &component{
			provider: provider,
			rType:    TypeOf(1),
		}

		type testInterface1 interface{}
		type testInterface2 interface{}

		com.AddAs(TypeOf((*testInterface1)(nil)), TypeOf((*testInterface2)(nil)))
		s := fmt.Sprintf("%v", com)
		assert.Equal(t, "Component[int]{as={model.testInterface1, model.testInterface2}}", s)
		vs := fmt.Sprintf("%+v", com)
		assert.Equal(t, "Component[int]{as={model.testInterface1, model.testInterface2}}", vs)
	})

	t.Run("String", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		com := &component{
			provider: provider,
			ignored:  true,
			hidden:   true,
			rType:    TypeOf(0),
			as:       newTypeSet(TypeOf(1)),
			name:     "abc",
			tags:     newSymbolSet(tag1),
		}

		expected := fmt.Sprintf("Component[int]{name=\"abc\", tags=%v, as=%v, ignored, hidden}",
			com.tags, com.as)
		assert.Equal(t, expected, fmt.Sprintf("%v", com))
	})
}

func Test_component_Validate(t *testing.T) {
	t.Run("not error", func(t *testing.T) {
		type testInterface interface{}
		tag1 := NewSymbol("tag1")

		provider := newProviderForComponentTest()
		com := &component{
			provider: provider,
			ignored:  true,
			hidden:   true,
			rType:    TypeOf(0),
			as:       newTypeSet(TypeOf((*testInterface)(nil))),
			name:     "abc",
			tags:     newSymbolSet(tag1),
		}
		err := com.Validate()
		assert.Nil(t, err)
	})

	t.Run("with nil type", func(t *testing.T) {
		type testInterface interface{}
		tag1 := NewSymbol("tag1")

		provider := newProviderForComponentTest()
		com := &component{
			provider: provider,
			ignored:  true,
			hidden:   true,
			rType:    TypeOf(nil),
			as:       newTypeSet(TypeOf((*testInterface)(nil))),
			name:     "abc",
			tags:     newSymbolSet(tag1),
		}
		err := com.Validate()

		assert.NotNil(t, err)
	})

	t.Run("with error type", func(t *testing.T) {
		tag1 := NewSymbol("tag1")

		provider := newProviderForComponentTest()
		com := &component{
			provider: provider,
			ignored:  true,
			hidden:   true,
			rType:    TypeOf(errors.Newf("this is an error")),
			as:       newTypeSet(),
			name:     "abc",
			tags:     newSymbolSet(tag1),
		}
		err := com.Validate()

		assert.NotNil(t, err)
	})

	t.Run("as an error type", func(t *testing.T) {
		tag1 := NewSymbol("tag1")

		provider := newProviderForComponentTest()
		com := &component{
			provider: provider,
			ignored:  true,
			hidden:   true,
			rType:    TypeOf(1),
			as:       newTypeSet(TypeOf((*error)(nil))),
			name:     "abc",
			tags:     newSymbolSet(tag1),
		}
		err := com.Validate()
		assert.NotNil(t, err)
	})

	t.Run("as a not interface type", func(t *testing.T) {
		tag1 := NewSymbol("tag1")

		provider := newProviderForComponentTest()
		com := &component{
			provider: provider,
			ignored:  true,
			hidden:   true,
			rType:    TypeOf(1),
			as:       newTypeSet(TypeOf(1)),
			name:     "abc",
			tags:     newSymbolSet(tag1),
		}
		err := com.Validate()
		assert.NotNil(t, err)
	})

	t.Run("as a interface that not be implemented", func(t *testing.T) {
		type testInterface interface {
			MethodA()
		}
		tag1 := NewSymbol("tag1")

		provider := newProviderForComponentTest()
		com := &component{
			provider: provider,
			ignored:  true,
			hidden:   true,
			rType:    TypeOf(1),
			as:       newTypeSet(TypeOf((*testInterface)(nil))),
			name:     "abc",
			tags:     newSymbolSet(tag1),
		}
		err := com.Validate()
		assert.NotNil(t, err)
	})
}

func Test_component_clone(t *testing.T) {
	tag1 := NewSymbol("tag1")
	val1 := valuer.Const(reflect.ValueOf(123))

	provider := newProviderForComponentTest()
	com := &component{
		provider: provider,
		ignored:  true,
		hidden:   true,
		rType:    TypeOf(0),
		as:       newTypeSet(TypeOf(1)),
		name:     "abc",
		tags:     newSymbolSet(tag1),
		val:      val1,
	}

	verifyComponent := func(t *testing.T, com Component) {
		assert.Equal(t, provider, com.Provider())
		assert.Equal(t, true, com.Ignored())
		assert.Equal(t, true, com.Hidden())
		assert.Equal(t, TypeOf(0), com.Type())
		assert.Equal(t, newTypeSet(TypeOf(1)), com.As())
		assert.Equal(t, "abc", com.Name())
		assert.Equal(t, newSymbolSet(tag1), com.Tags())
		assert.Equal(t, val1, com.Valuer())
	}

	t.Run("equality", func(t *testing.T) {
		com2 := com.clone()
		verifyComponent(t, com2)
		assert.False(t, com2.Valuer() == com.Valuer())
	})

	t.Run("update isolation", func(t *testing.T) {
		tag2 := NewSymbol("tag2")

		com2 := com.clone()
		com2.SetIgnore(false)
		com2.SetHidden(false)
		com2.AddAs(TypeOf("abc"))
		com2.SetName("def")
		com2.AddTags(tag2)

		verifyComponent(t, com)
	})

	t.Run("update isolation 2", func(t *testing.T) {
		tag2 := NewSymbol("tag2")
		com3 := com.clone()
		com.SetIgnore(false)
		com.SetHidden(false)
		com.AddAs(TypeOf("abc"))
		com.SetName("def")
		com.AddTags(tag2)

		verifyComponent(t, com3)
	})

	t.Run("nil", func(t *testing.T) {
		var com2 *component
		assert.Nil(t, com2.clone())
	})
}

func Test_component_Iterate(t *testing.T) {
	t.Run("continue", func(t *testing.T) {
		provider := newProviderForComponentTest()
		com := &component{
			provider: provider,
		}

		isContinue := com.Iterate(func(c Component) bool {
			assert.Equal(t, com, c)
			return true
		})

		assert.True(t, isContinue)
	})

	t.Run("not continue", func(t *testing.T) {
		provider := newProviderForComponentTest()
		com := &component{
			provider: provider,
		}

		isContinue := com.Iterate(func(c Component) bool {
			assert.Equal(t, com, c)
			return false
		})

		assert.False(t, isContinue)
	})
}
