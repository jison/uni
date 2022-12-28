package model

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jison/uni/core/valuer"
	"github.com/jison/uni/internal/location"
	"github.com/stretchr/testify/assert"
)

func TestStruct(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		//lint:ignore U1000 we need the field name to locate the field
		type testStruct struct {
			a int
			B string
			c []int
			D rune
		}

		type testInterface interface{}

		tag1 := NewSymbol("tag1")
		tag2 := NewSymbol("tag2")
		scope1 := NewScope("scope1")
		baseLoc := location.GetCallLocation(0)
		sp := Struct(TypeOf(&testStruct{}),
			Field("a", ByName("a1"), ByTags(tag1), Optional(true)),
			Field("c", ByName("c1"), AsCollector(true)),
			Field("D", ByName("D1")),
			IgnoreFields(func(field reflect.StructField) bool {
				return field.Name == "B"
			}),
			Name("testStruct"),
			Tags(tag2),
			As(TypeOf((*testInterface)(nil))),
			Hide(),
			Ignore(),
			InScope(scope1),
			nil,
		)

		pro := sp.Provider()
		err := pro.Validate()
		assert.Nil(t, err)

		deps := dependencyIteratorToArray(pro.Dependencies())
		assert.Equal(t, 3, len(deps))
		for _, dep := range deps {
			if dep.Name() == "a1" {
				assert.Equal(t, TypeOf(1), dep.Type())
				assert.Equal(t, newSymbolSet(tag1), dep.Tags())
				assert.True(t, dep.Optional())
				assert.Equal(t, valuer.Field("a"), dep.Valuer())
			} else if dep.Name() == "c1" {
				assert.Equal(t, TypeOf(0), dep.Type())
				assert.Equal(t, "c1", dep.Name())
				assert.Equal(t, valuer.Field("c"), dep.Valuer())
				assert.True(t, dep.IsCollector())
			} else if dep.Name() == "D1" {
				assert.Equal(t, TypeOf('d'), dep.Type())
				assert.Equal(t, "D1", dep.Name())
				assert.Equal(t, valuer.Field("D"), dep.Valuer())
			}

			assert.Same(t, pro, dep.Consumer())
		}

		coms := pro.Components().ToArray()
		assert.Equal(t, 1, len(coms))
		com := coms[0]
		assert.Equal(t, TypeOf(&testStruct{}), com.Type())
		assert.Equal(t, "testStruct", com.Name())
		assert.Equal(t, newSymbolSet(tag2), com.Tags())
		assert.Equal(t, true, com.Ignored())
		assert.Equal(t, true, com.Hidden())
		assert.Equal(t, newTypeSet(TypeOf((*testInterface)(nil))), com.As())
		assert.Same(t, pro, com.Provider())

		assert.Equal(t, valuer.Struct(TypeOf(&testStruct{})), pro.Valuer())
		assert.Equal(t, baseLoc.FileName(), pro.Location().FileName())
		assert.Equal(t, baseLoc.FileLine()+1, pro.Location().FileLine())
		assert.Equal(t, scope1, pro.Scope())
	})

	t.Run("nil type", func(t *testing.T) {
		type testInterface interface{}

		tag1 := NewSymbol("tag1")
		tag2 := NewSymbol("tag2")
		scope1 := NewScope("scope1")
		sp := Struct(TypeOf(nil),
			Field("a", ByName("a1"), ByTags(tag1), Optional(true)),
			Field("c", ByName("c1"), AsCollector(true)),
			Field("D", ByName("D1")),
			IgnoreFields(func(field reflect.StructField) bool {
				return field.Name == "B"
			}),
			Name("testStruct"),
			Tags(tag2),
			As(TypeOf((*testInterface)(nil))),
			Hide(),
			Ignore(),
			InScope(scope1),
			nil,
		)
		p := sp.Provider()
		err := p.Validate()
		assert.NotNil(t, err)
	})
}

func Test_structProvider_Provider(t *testing.T) {
	//lint:ignore U1000 we need the field name to locate the field
	type testStruct struct {
		a int
		B string
		c []int
		D rune
	}

	type testInterface interface{}

	tag1 := NewSymbol("tag1")
	tag2 := NewSymbol("tag2")
	scope1 := NewScope("scope1")
	loc1 := location.GetCallLocation(0)
	sp0 := structProviderOf(TypeOf(&testStruct{}),
		Field("a", ByName("a1"), ByTags(tag1), Optional(true)),
		Field("c", ByName("c1"), AsCollector(true)),
		Field("D", ByName("D1")),
		IgnoreFields(func(field reflect.StructField) bool {
			return field.Name == "B"
		}),
		Name("testStruct"),
		Tags(tag2),
		As(TypeOf((*testInterface)(nil))),
		Hide(),
		Ignore(),
		InScope(scope1),
		Location(loc1),
	)

	t.Run("Dependencies", func(t *testing.T) {
		sp := sp0.clone()
		pro := sp.Provider()

		deps := dependencyIteratorToArray(pro.Dependencies())
		assert.Equal(t, 3, len(deps))
		for _, dep := range deps {
			if dep.Name() == "a1" {
				assert.Equal(t, TypeOf(1), dep.Type())
				assert.Equal(t, newSymbolSet(tag1), dep.Tags())
				assert.True(t, dep.Optional())
				assert.Equal(t, valuer.Field("a"), dep.Valuer())
			} else if dep.Name() == "c1" {
				assert.Equal(t, TypeOf(0), dep.Type())
				assert.Equal(t, "c1", dep.Name())
				assert.Equal(t, valuer.Field("c"), dep.Valuer())
				assert.True(t, dep.IsCollector())
			} else if dep.Name() == "D1" {
				assert.Equal(t, TypeOf('d'), dep.Type())
				assert.Equal(t, "D1", dep.Name())
				assert.Equal(t, valuer.Field("D"), dep.Valuer())
			}
			assert.True(t, dep.Valuer() == dep.Valuer())
			assert.Same(t, pro, dep.Consumer())
		}
	})

	t.Run("Valuer", func(t *testing.T) {
		sp := sp0.clone()
		pro := sp.Provider()
		assert.Equal(t, valuer.Struct(TypeOf(&testStruct{})), pro.Valuer())
		assert.True(t, pro.Valuer() == pro.Valuer())
	})

	t.Run("Location", func(t *testing.T) {
		sp := sp0.clone()
		pro := sp.Provider()
		assert.Equal(t, loc1, pro.Location())
	})

	t.Run("UpdateCallLocation", func(t *testing.T) {
		baseLoc := location.GetCallLocation(0)
		var sp StructProviderBuilder
		func() {
			sp = structProviderOf(TypeOf(&testStruct{}), UpdateCallLocation())
		}()
		p := sp.Provider()

		assert.Equal(t, baseLoc.FileName(), p.Location().FileName())
		assert.Equal(t, baseLoc.FileLine()+4, p.Location().FileLine())
	})

	t.Run("Components", func(t *testing.T) {
		sp := sp0.clone()
		pro := sp.Provider()
		coms := pro.Components().ToArray()
		assert.Equal(t, 1, len(coms))
		com := coms[0]
		assert.Equal(t, TypeOf(&testStruct{}), com.Type())
		assert.Equal(t, "testStruct", com.Name())
		assert.Equal(t, newSymbolSet(tag2), com.Tags())
		assert.Equal(t, true, com.Ignored())
		assert.Equal(t, true, com.Hidden())
		assert.Equal(t, newTypeSet(TypeOf((*testInterface)(nil))), com.As())
		assert.Equal(t, valuer.Identity(), com.Valuer())
		assert.True(t, com.Valuer() == com.Valuer())
		assert.Same(t, pro, com.Provider())
	})

	t.Run("Scope", func(t *testing.T) {
		t.Run("scope", func(t *testing.T) {
			sp := sp0.clone()
			pro := sp.Provider()
			assert.Equal(t, scope1, pro.Scope())
		})

		t.Run("nil scope", func(t *testing.T) {
			sp := structProviderOf(TypeOf(&testStruct{}), InScope(nil))
			assert.Equal(t, GlobalScope, sp.Provider().Scope())
		})
	})

	t.Run("Validate", func(t *testing.T) {
		t.Run("no errors", func(t *testing.T) {
			sp := sp0.clone()
			pro := sp.Provider()
			err := pro.Validate()
			assert.Nil(t, err)
		})

		t.Run("no struct type", func(t *testing.T) {
			intVal := 1
			strVal := ""
			runeVal := 'a'

			tests := []struct {
				name string
				t    reflect.Type
			}{
				{"int", TypeOf(1)},
				{"string", TypeOf("")},
				{"rune", TypeOf('a')},
				{"[]int", TypeOf([]int{})},
				{"[]string", TypeOf([]string{})},
				{"[]rune", TypeOf([]rune{})},
				{"function", TypeOf(func() {})},
				{"*int", TypeOf(&intVal)},
				{"*string", TypeOf(&strVal)},
				{"*rune", TypeOf(&runeVal)},
				{"map[int]string", TypeOf(map[int]string{})},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					err := Struct(tt.t).Provider().Validate()
					assert.NotNil(t, err)
				})
			}
		})

		t.Run("error type", func(t *testing.T) {
			err := Struct(TypeOf(&errorTypeForStructConsumerTest{})).Provider().Validate()
			assert.NotNil(t, err)

			err2 := Struct(TypeOf(errorTypeForStructConsumerTest{})).Provider().Validate()
			assert.Nil(t, err2)
		})

		t.Run("field not exist", func(t *testing.T) {
			sp := sp0.clone()
			sp.Field("d", ByName("abc"))
			pro := sp.Provider()
			err := pro.Validate()

			assert.NotNil(t, err)
		})

		t.Run("field with error", func(t *testing.T) {
			sp := sp0.clone()
			sp.Field("a", AsCollector(true))
			pro := sp.Provider()
			err := pro.Validate()

			assert.NotNil(t, err)
		})

		t.Run("error from structConsumer", func(t *testing.T) {

		})

		t.Run("component with error", func(t *testing.T) {
			sp := sp0.clone()
			sp.AddAs(TypeOf(1))
			err := sp.Provider().Validate()
			assert.NotNil(t, err)
		})
	})

	t.Run("Format", func(t *testing.T) {
		sp := sp0.clone()
		pro := sp.Provider()

		t.Run("not verbose", func(t *testing.T) {
			expected := fmt.Sprintf("Struct[%v] in %v", TypeOf(&testStruct{}), sp.Scope())
			assert.Equal(t, expected, fmt.Sprintf("%v", pro))
		})

		t.Run("verbose", func(t *testing.T) {
			expected := fmt.Sprintf("Struct[%v] in %v at %v", TypeOf(&testStruct{}), sp.Scope(),
				pro.Location())
			assert.Equal(t, expected, fmt.Sprintf("%+v", pro))
		})
	})
}

func Test_structProvider_StructProviderBuilder(t *testing.T) {
	//lint:ignore U1000 we need the field name to locate the field
	type testStruct struct {
		a int
		B string
		c []int
		D rune
	}
	type testInterface interface{}

	tag1 := NewSymbol("tag1")
	tag2 := NewSymbol("tag2")
	scope1 := NewScope("scope1")
	loc1 := location.GetCallLocation(0)
	sp := structProviderOf(TypeOf(&testStruct{}))

	t.Run("ApplyModule", func(t *testing.T) {
		sp1 := sp.clone()
		sp1.SetName("testStruct")
		sp1.AddTags(tag2)
		sp1.AddAs(TypeOf((*testInterface)(nil)))
		sp1.SetHidden(true)
		sp1.SetIgnore(true)
		sp1.SetScope(scope1)
		sp1.SetLocation(loc1)

		mb := NewModuleBuilder()
		sp1.ApplyModule(mb)

		coms := mb.Module().AllComponents().ToArray()
		assert.Equal(t, 1, len(coms))
		com := coms[0]
		assert.Equal(t, TypeOf(&testStruct{}), com.Type())
		assert.Equal(t, "testStruct", com.Name())
		assert.Equal(t, newSymbolSet(tag2), com.Tags())
		assert.Equal(t, true, com.Ignored())
		assert.Equal(t, true, com.Hidden())
		assert.Equal(t, newTypeSet(TypeOf((*testInterface)(nil))), com.As())
	})

	t.Run("Provide", func(t *testing.T) {
		sp1 := sp.clone()
		sp1.Field("a", ByName("a1"), ByTags(tag1), Optional(true))
		sp1.Field("c", ByName("c1"), AsCollector(true))
		sp1.Field("D", ByName("D1"))
		sp1.IgnoreFields(func(field reflect.StructField) bool {
			return field.Name == "B"
		})
		sp1.SetName("testStruct")
		sp1.AddTags(tag2)
		sp1.AddAs(TypeOf((*testInterface)(nil)))
		sp1.SetHidden(true)
		sp1.SetIgnore(true)
		sp1.SetScope(scope1)
		sp1.SetLocation(loc1)

		pro := sp1.Provider()
		err := pro.Validate()
		assert.Nil(t, err)

		deps := dependencyIteratorToArray(pro.Dependencies())
		assert.Equal(t, 3, len(deps))
		for _, dep := range deps {
			if dep.Name() == "a1" {
				assert.Equal(t, TypeOf(1), dep.Type())
				assert.Equal(t, newSymbolSet(tag1), dep.Tags())
				assert.True(t, dep.Optional())
				assert.Equal(t, valuer.Field("a"), dep.Valuer())
			} else if dep.Name() == "c1" {
				assert.Equal(t, TypeOf(0), dep.Type())
				assert.Equal(t, "c1", dep.Name())
				assert.Equal(t, valuer.Field("c"), dep.Valuer())
				assert.True(t, dep.IsCollector())
			} else if dep.Name() == "D1" {
				assert.Equal(t, TypeOf('d'), dep.Type())
				assert.Equal(t, "D1", dep.Name())
				assert.Equal(t, valuer.Field("D"), dep.Valuer())
			}

			assert.Same(t, pro, dep.Consumer())
		}

		coms := pro.Components().ToArray()
		assert.Equal(t, 1, len(coms))
		com := coms[0]
		assert.Equal(t, TypeOf(&testStruct{}), com.Type())
		assert.Equal(t, "testStruct", com.Name())
		assert.Equal(t, newSymbolSet(tag2), com.Tags())
		assert.Equal(t, true, com.Ignored())
		assert.Equal(t, true, com.Hidden())
		assert.Equal(t, newTypeSet(TypeOf((*testInterface)(nil))), com.As())
		assert.Same(t, pro, com.Provider())

		assert.Equal(t, valuer.Struct(TypeOf(&testStruct{})), pro.Valuer())
		assert.Equal(t, loc1, pro.Location())
		assert.Equal(t, scope1, pro.Scope())
	})

	t.Run("Field", func(t *testing.T) {
		sp1 := sp.clone()
		sp1.Field("a", ByName("a1"), ByTags(tag1), Optional(true))
		sp1.Field("c", ByName("c1"), AsCollector(true))
		sp1.Field("D", ByName("D1"))
		sp1.IgnoreFields(func(field reflect.StructField) bool {
			return field.Name == "B"
		})

		pro := sp1.Provider()
		err := pro.Validate()
		assert.Nil(t, err)

		deps := dependencyIteratorToArray(pro.Dependencies())
		assert.Equal(t, 3, len(deps))
		for _, dep := range deps {
			if dep.Name() == "a1" {
				assert.Equal(t, TypeOf(1), dep.Type())
				assert.Equal(t, newSymbolSet(tag1), dep.Tags())
				assert.True(t, dep.Optional())
				assert.Equal(t, valuer.Field("a"), dep.Valuer())
			} else if dep.Name() == "c1" {
				assert.Equal(t, TypeOf(0), dep.Type())
				assert.Equal(t, "c1", dep.Name())
				assert.Equal(t, valuer.Field("c"), dep.Valuer())
				assert.True(t, dep.IsCollector())
			} else if dep.Name() == "D1" {
				assert.Equal(t, TypeOf('d'), dep.Type())
				assert.Equal(t, "D1", dep.Name())
				assert.Equal(t, valuer.Field("D"), dep.Valuer())
			}

			assert.Same(t, pro, dep.Consumer())
		}
	})

	t.Run("IgnoreFields", func(t *testing.T) {
		sp1 := sp.clone()
		sp1.Field("B", ByName("B1"))
		sp1.IgnoreFields(func(field reflect.StructField) bool {
			return field.Name == "B"
		})
		pro := sp1.Provider()
		meet := false
		pro.Dependencies().Iterate(func(dep Dependency) bool {
			if dep.Name() == "B1" {
				meet = true
			}
			return true
		})

		assert.False(t, meet)
	})

	t.Run("SetIgnore", func(t *testing.T) {
		sp1 := sp.clone()
		sp1.SetIgnore(true)
		coms := sp1.Provider().Components().ToArray()
		assert.Equal(t, 1, len(coms))
		com := coms[0]
		assert.Equal(t, true, com.Ignored())
	})

	t.Run("SetHidden", func(t *testing.T) {
		sp1 := sp.clone()
		sp1.SetHidden(true)
		coms := sp1.Provider().Components().ToArray()
		assert.Equal(t, 1, len(coms))
		com := coms[0]
		assert.Equal(t, true, com.Hidden())
	})

	t.Run("AddAs", func(t *testing.T) {
		sp1 := sp.clone()
		sp1.AddAs(TypeOf(1))
		coms := sp1.Provider().Components().ToArray()
		assert.Equal(t, 1, len(coms))
		com := coms[0]
		fmt.Printf("%v\n", com)
		assert.Equal(t, newTypeSet(TypeOf(1)), com.As())
	})

	t.Run("SetName", func(t *testing.T) {
		sp1 := sp.clone()
		sp1.SetName("abc")
		coms := sp1.Provider().Components().ToArray()
		assert.Equal(t, 1, len(coms))
		com := coms[0]
		assert.Equal(t, "abc", com.Name())
	})

	t.Run("AddTags", func(t *testing.T) {
		sp1 := sp.clone()
		sp1.AddTags(tag2)
		coms := sp1.Provider().Components().ToArray()
		assert.Equal(t, 1, len(coms))
		com := coms[0]
		assert.Equal(t, newSymbolSet(tag2), com.Tags())
	})

	t.Run("SetScope", func(t *testing.T) {
		t.Run("set scope", func(t *testing.T) {
			sp1 := sp.clone()
			sp1.SetScope(scope1)
			pro := sp1.Provider()
			assert.Equal(t, scope1, pro.Scope())
		})

		t.Run("set nil", func(t *testing.T) {
			sp1 := sp.clone()
			sp1.SetScope(nil)
			pro := sp1.Provider()
			assert.Equal(t, GlobalScope, pro.Scope())
		})
	})

	t.Run("SetLocation", func(t *testing.T) {
		loc2 := location.GetCallLocation(0)

		sp1 := sp.clone()
		sp1.SetLocation(loc2)
		pro := sp1.Provider()
		assert.Equal(t, loc2, pro.Location())
	})

	t.Run("UpdateCallLocation", func(t *testing.T) {
		t.Run("location have been set", func(t *testing.T) {
			sp1 := sp.clone()
			sp1.SetLocation(loc1)
			sp1.UpdateCallLocation(nil)
			assert.Equal(t, loc1, sp1.Location())
		})

		t.Run("location have not been set", func(t *testing.T) {
			sp1 := sp.clone()
			baseLoc := location.GetCallLocation(0)
			func() {
				sp1.UpdateCallLocation(nil)
			}()
			assert.Equal(t, baseLoc.FileName(), sp1.Location().FileName())
			assert.Equal(t, baseLoc.FileLine()+3, sp1.Location().FileLine())
		})

		t.Run("location is not nil", func(t *testing.T) {
			sp1 := sp.clone()
			sp1.UpdateCallLocation(loc1)
			assert.Equal(t, loc1, sp1.Location())
		})
	})
}

func Test_structProvider_clone(t *testing.T) {
	//lint:ignore U1000 we need the field name to locate the field
	type testStruct struct {
		a int
		B string
		c []int
		D rune
	}
	type testInterface interface{}

	tag1 := NewSymbol("tag1")
	tag2 := NewSymbol("tag2")
	scope1 := NewScope("scope1")
	baseLoc := location.GetCallLocation(0)
	sp := structProviderOf(TypeOf(&testStruct{}))
	sp.Field("a", ByName("a1"), ByTags(tag1), Optional(true))
	sp.Field("c", ByName("c1"), AsCollector(true))
	sp.Field("D", ByName("D1"))
	sp.IgnoreFields(func(field reflect.StructField) bool {
		return field.Name == "B"
	})
	sp.SetName("testStruct")
	sp.AddTags(tag2)
	sp.AddAs(TypeOf((*testInterface)(nil)))
	sp.SetHidden(true)
	sp.SetIgnore(true)
	sp.SetScope(scope1)
	sp.SetLocation(baseLoc)

	verifyProvider := func(t *testing.T, pro Provider) {
		deps := dependencyIteratorToArray(pro.Dependencies())
		assert.Equal(t, 3, len(deps))
		for _, dep := range deps {
			if dep.Name() == "a1" {
				assert.Equal(t, TypeOf(1), dep.Type())
				assert.Equal(t, newSymbolSet(tag1), dep.Tags())
				assert.True(t, dep.Optional())
				assert.Equal(t, valuer.Field("a"), dep.Valuer())
			} else if dep.Name() == "c1" {
				assert.Equal(t, TypeOf(0), dep.Type())
				assert.Equal(t, "c1", dep.Name())
				assert.Equal(t, valuer.Field("c"), dep.Valuer())
				assert.True(t, dep.IsCollector())
			} else if dep.Name() == "D1" {
				assert.Equal(t, TypeOf('d'), dep.Type())
				assert.Equal(t, "D1", dep.Name())
				assert.Equal(t, valuer.Field("D"), dep.Valuer())
			}

			assert.Same(t, pro, dep.Consumer())
		}

		coms := pro.Components().ToArray()
		assert.Equal(t, 1, len(coms))
		com := coms[0]
		assert.Equal(t, TypeOf(&testStruct{}), com.Type())
		assert.Equal(t, "testStruct", com.Name())
		assert.Equal(t, newSymbolSet(tag2), com.Tags())
		assert.Equal(t, true, com.Ignored())
		assert.Equal(t, true, com.Hidden())
		assert.Equal(t, newTypeSet(TypeOf((*testInterface)(nil))), com.As())
		assert.Same(t, pro, com.Provider())

		assert.Equal(t, valuer.Struct(TypeOf(&testStruct{})), pro.Valuer())
		assert.Equal(t, baseLoc, pro.Location())
		assert.Equal(t, scope1, pro.Scope())
	}

	t.Run("equality", func(t *testing.T) {
		sp2 := sp.clone()
		verifyProvider(t, sp2.Provider())

		assert.False(t, sp2.Valuer() == sp.Valuer())
	})

	t.Run("update isolation", func(t *testing.T) {
		scope2 := NewScope("scope2")
		loc2 := location.GetCallLocation(0)

		sp2 := sp.clone()
		sp2.Field("a", ByName("a2"), ByTags(tag2), Optional(false))
		sp2.Field("c", ByName("c2"), AsCollector(false))
		sp2.IgnoreFields(func(field reflect.StructField) bool {
			return field.Name == "D"
		})
		sp2.SetName("testStruct2")
		sp2.AddTags(tag1)
		sp2.AddAs(TypeOf(1))
		sp2.SetHidden(false)
		sp2.SetIgnore(false)
		sp2.SetScope(scope2)
		sp2.SetLocation(loc2)

		verifyProvider(t, sp.Provider())
	})

	t.Run("update isolation 2", func(t *testing.T) {
		scope2 := NewScope("scope2")
		loc2 := location.GetCallLocation(0)

		sp2 := sp.clone()
		sp3 := sp2.clone()

		sp2.Field("a", ByName("a2"), ByTags(tag2), Optional(false))
		sp2.Field("c", ByName("c2"), AsCollector(false))
		sp2.IgnoreFields(func(field reflect.StructField) bool {
			return field.Name == "D"
		})
		sp2.SetName("testStruct2")
		sp2.AddTags(tag1)
		sp2.AddAs(TypeOf(1))
		sp2.SetHidden(false)
		sp2.SetIgnore(false)
		sp2.SetScope(scope2)
		sp2.SetLocation(loc2)

		verifyProvider(t, sp3.Provider())
	})

	t.Run("nil", func(t *testing.T) {
		var sp *structProvider
		assert.Nil(t, sp.clone())
	})
}

func Test_structProvider_Equal(t *testing.T) {
	//lint:ignore U1000 we need the field name to locate the field
	type testStruct struct {
		a int
		B string
		c []int
		D rune
	}
	type testInterface interface{}

	tag1 := NewSymbol("tag1")
	tag2 := NewSymbol("tag2")
	scope1 := NewScope("scope1")
	baseLoc := location.GetCallLocation(0)
	sp := structProviderOf(TypeOf(&testStruct{}))
	sp.Field("a", ByName("a1"), ByTags(tag1), Optional(true))
	sp.Field("c", ByName("c1"), AsCollector(true))
	sp.Field("D", ByName("D1"))
	sp.IgnoreFields(func(field reflect.StructField) bool {
		return field.Name == "B"
	})
	sp.SetName("testStruct")
	sp.AddTags(tag2)
	sp.AddAs(TypeOf((*testInterface)(nil)))
	sp.SetHidden(true)
	sp.SetIgnore(true)
	sp.SetScope(scope1)
	sp.SetLocation(baseLoc)

	t.Run("equal", func(t *testing.T) {
		sp2 := sp.clone()
		assert.True(t, sp2.Equal(sp))
	})

	t.Run("not equal with other kind of provider", func(t *testing.T) {
		sp2 := sp.clone()
		vp := valueProviderOf(123)
		assert.False(t, sp2.Equal(vp))
	})

	t.Run("nil equal nil", func(t *testing.T) {
		var sp2 *structProvider
		var sp3 *structProvider
		assert.True(t, sp2.Equal(sp3))
	})

	t.Run("structConsumer", func(t *testing.T) {
		t.Run("not nil", func(t *testing.T) {
			sp2 := sp.clone()
			sp2.Field("a", ByName("a2"))
			assert.False(t, sp2.Equal(sp))
		})

		t.Run("nil", func(t *testing.T) {
			sp2 := sp.clone()
			sp2.structConsumer = nil
			assert.False(t, sp2.Equal(sp))
			assert.False(t, sp.Equal(sp2))

			sp3 := sp.clone()
			sp3.structConsumer = nil
			assert.True(t, sp3.Equal(sp2))
		})

	})

	t.Run("component", func(t *testing.T) {
		t.Run("not nil", func(t *testing.T) {
			sp2 := sp.clone()
			sp2.SetName("s2")
			assert.False(t, sp2.Equal(sp))
		})

		t.Run("nil", func(t *testing.T) {
			sp2 := sp.clone()
			sp2.com = nil
			assert.False(t, sp2.Equal(sp))
			assert.False(t, sp.Equal(sp2))

			sp3 := sp.clone()
			sp3.com = nil

			assert.True(t, sp3.Equal(sp2))
		})
	})

}
