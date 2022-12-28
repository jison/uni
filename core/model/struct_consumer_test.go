package model

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jison/uni/core/valuer"
	"github.com/jison/uni/internal/location"
	"github.com/stretchr/testify/assert"
)

type errorTypeForStructConsumerTest struct {
	err string
}

func (s *errorTypeForStructConsumerTest) Error() string {
	return s.err
}

func Test_structField_Format(t *testing.T) {
	//lint:ignore U1000 we need the field name to locate the field
	type testStruct struct {
		a int
	}

	sc := structConsumerOf(TypeOf(testStruct{}))
	sf := sc.fields["a"]

	t.Run("no verbose", func(t *testing.T) {
		expected := fmt.Sprintf("%v at field `a` of %+v", sf.dependency, sf.dependency.Consumer())
		assert.Equal(t, expected, fmt.Sprintf("%+v", sf))
	})

	t.Run("verbose", func(t *testing.T) {
		expected := fmt.Sprintf("%v at field `a`", sf.dependency)
		assert.Equal(t, expected, fmt.Sprintf("%v", sf))
	})
}

func Test_structField_clone(t *testing.T) {
	//lint:ignore U1000 we need the field name to locate the field
	type testStruct struct {
		a int
	}

	sc := structConsumerOf(TypeOf(testStruct{}))
	sf := sc.fields["a"]

	t.Run("equality", func(t *testing.T) {
		sf2 := sf.clone()
		assert.True(t, sf2.Equal(sf))
	})

	t.Run("update isolation", func(t *testing.T) {
		sf2 := sf.clone()
		sf2.field.Name = "aaa"
		sf2.ignored = true
		sf2.dependency.optional = true

		assert.Equal(t, "a", sf.field.Name)
		assert.Equal(t, false, sf.ignored)
		assert.Equal(t, false, sf.optional)
	})

	t.Run("nil", func(t *testing.T) {
		var sf2 *structField
		assert.Nil(t, sf2.clone())
	})
}

func Test_structField_Equal(t *testing.T) {
	//lint:ignore U1000 we need the field name to locate the field
	type testStruct struct {
		a int
	}

	sc := structConsumerOf(TypeOf(testStruct{}))
	sf := sc.fields["a"]

	t.Run("equal", func(t *testing.T) {
		sf2 := sf.clone()
		assert.True(t, sf2.Equal(sf))
	})

	t.Run("not equal to non structField", func(t *testing.T) {
		assert.False(t, sf.Equal(123))
	})

	t.Run("nil equal nil", func(t *testing.T) {
		var sf1 *structField
		var sf2 *structField
		assert.True(t, sf1.Equal(sf2))
	})

	t.Run("ignored", func(t *testing.T) {
		sf2 := sf.clone()
		sf2.ignored = true
		assert.False(t, sf2.Equal(sf))
	})

	t.Run("field", func(t *testing.T) {
		sf2 := sf.clone()
		sf2.field.Name = "b"
		assert.False(t, sf2.Equal(sf))
	})

	t.Run("dependency", func(t *testing.T) {
		sf2 := sf.clone()
		sf2.dependency.optional = true
		assert.False(t, sf2.Equal(sf))

		sf3 := sf.clone()
		sf3.dependency = nil
		assert.False(t, sf3.Equal(sf))
	})
}

func TestStructConsumer(t *testing.T) {
	t.Run("normal type", func(t *testing.T) {
		//lint:ignore U1000 we need the field name to locate the field
		type testStruct struct {
			a int
			B string
			c []int
		}

		tag1 := NewSymbol("tag1")
		baseLoc := location.GetCallLocation(0)
		sc := StructConsumer(TypeOf(testStruct{}),
			Field("a", ByName("abc"), ByTags(tag1), Optional(true)),
			Field("c", ByName("cbd"), AsCollector(true)),
			IgnoreFields(func(field reflect.StructField) bool {
				return field.Name == "B"
			}),
			nil,
		)
		con := sc.Consumer()
		err := con.Validate()
		assert.Nil(t, err)

		deps := dependencyIteratorToArray(con.Dependencies())
		assert.Equal(t, 2, len(deps))
		for _, dep := range deps {
			if dep.Name() == "abc" {
				assert.Equal(t, TypeOf(1), dep.Type())
				assert.Equal(t, newSymbolSet(tag1), dep.Tags())
				assert.True(t, dep.Optional())
				assert.Equal(t, valuer.Field("a"), dep.Valuer())
			} else {
				assert.Equal(t, TypeOf(0), dep.Type())
				assert.Equal(t, "cbd", dep.Name())
				assert.Equal(t, valuer.Field("c"), dep.Valuer())
				assert.True(t, dep.IsCollector())
			}

			assert.Same(t, con, dep.Consumer())
		}

		assert.Equal(t, valuer.Struct(TypeOf(testStruct{})), con.Valuer())
		assert.Equal(t, baseLoc.FileName(), con.Location().FileName())
		assert.Equal(t, baseLoc.FileLine()+1, con.Location().FileLine())
	})

	t.Run("nil type", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		sc := StructConsumer(TypeOf(nil),
			Field("a", ByName("abc"), ByTags(tag1), Optional(true)),
			Field("c", ByName("cbd"), AsCollector(true)),
			IgnoreFields(func(field reflect.StructField) bool {
				return field.Name == "B"
			}),
			nil,
		)
		con := sc.Consumer()
		err := con.Validate()
		assert.NotNil(t, err)
	})
}

func Test_structConsumer_Consumer(t *testing.T) {
	//lint:ignore U1000 we need the field name to locate the field
	type testStruct struct {
		a int
		B string
		c []int
	}

	tag1 := NewSymbol("tag1")
	loc1 := location.GetCallLocation(0)
	sc := structConsumerOf(TypeOf(testStruct{}),
		Field("a", ByName("abc"), ByTags(tag1), Optional(true)),
		Field("c", ByName("cbd"), AsCollector(true)),
		IgnoreFields(func(field reflect.StructField) bool {
			return field.Name == "B"
		}),
		Location(loc1),
	)

	t.Run("Dependencies", func(t *testing.T) {
		con := sc.Consumer()

		deps := dependencyIteratorToArray(con.Dependencies())
		for _, dep := range deps {
			if dep.Name() == "abc" {
				assert.Equal(t, TypeOf(1), dep.Type())
				assert.Equal(t, newSymbolSet(tag1), dep.Tags())
				assert.True(t, dep.Optional())
				assert.Equal(t, valuer.Field("a"), dep.Valuer())
			} else {
				assert.Equal(t, TypeOf(0), dep.Type())
				assert.Equal(t, "cbd", dep.Name())
				assert.Equal(t, valuer.Field("c"), dep.Valuer())
				assert.True(t, dep.IsCollector())
			}
			assert.True(t, dep.Valuer() == dep.Valuer())
			assert.Same(t, con, dep.Consumer())
		}
	})

	t.Run("Valuer", func(t *testing.T) {
		con := sc.Consumer()
		assert.Equal(t, valuer.Struct(TypeOf(testStruct{})), con.Valuer())
		assert.True(t, con.Valuer() == con.Valuer())
	})

	t.Run("Scope", func(t *testing.T) {
		t.Run("scope", func(t *testing.T) {
			scope1 := NewScope("scope1")

			sc1 := structConsumerOf(TypeOf(testStruct{}), InScope(scope1))
			con := sc1.Consumer()
			assert.Equal(t, scope1, con.Scope())
		})

		t.Run("nil scope", func(t *testing.T) {
			con := sc.Consumer()
			assert.Equal(t, GlobalScope, con.Scope())
		})
	})

	t.Run("Location", func(t *testing.T) {
		con := sc.Consumer()
		assert.Equal(t, loc1, con.Location())
	})

	t.Run("UpdateCallLocation", func(t *testing.T) {
		baseLoc := location.GetCallLocation(0)
		var sc2 *structConsumer
		func() {
			sc2 = structConsumerOf(TypeOf(testStruct{}), UpdateCallLocation())
		}()
		c := sc2.Consumer()
		assert.Equal(t, baseLoc.FileName(), c.Location().FileName())
		assert.Equal(t, baseLoc.FileLine()+4, c.Location().FileLine())
	})

	t.Run("Validate", func(t *testing.T) {
		t.Run("no errors", func(t *testing.T) {
			con := sc.Consumer()
			err := con.Validate()
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
					err := StructConsumer(tt.t).Consumer().Validate()
					assert.NotNil(t, err)
				})
			}
		})

		t.Run("error type", func(t *testing.T) {
			err := StructConsumer(TypeOf(&errorTypeForStructConsumerTest{})).Consumer().Validate()
			assert.NotNil(t, err)

			err2 := StructConsumer(TypeOf(errorTypeForStructConsumerTest{})).Consumer().Validate()
			assert.Nil(t, err2)
		})

		t.Run("field not exist", func(t *testing.T) {
			sc.Field("d", ByName("abc"))
			con := sc.Consumer()
			err := con.Validate()

			assert.NotNil(t, err)
		})

		t.Run("field with error", func(t *testing.T) {
			sc.Field("a", AsCollector(true))
			con := sc.Consumer()
			err := con.Validate()

			assert.NotNil(t, err)
		})
	})

	t.Run("Format", func(t *testing.T) {
		con := sc.Consumer()

		t.Run("not verbose", func(t *testing.T) {
			expected := fmt.Sprintf("StructConsumer[%v]", TypeOf(testStruct{}))
			assert.Equal(t, expected, fmt.Sprintf("%v", con))
		})

		t.Run("verbose", func(t *testing.T) {
			expected := fmt.Sprintf("StructConsumer[%v] at %v", TypeOf(testStruct{}), con.Location())
			assert.Equal(t, expected, fmt.Sprintf("%+v", con))
		})
	})
}

func Test_structConsumer_StructConsumerBuilder(t *testing.T) {
	//lint:ignore U1000 we need the field name to locate the field
	type testStruct struct {
		a int
		B string
		c []int
	}

	tag1 := NewSymbol("tag1")
	sc0 := structConsumerOf(TypeOf(testStruct{}))

	t.Run("Field", func(t *testing.T) {
		t.Run("Optional", func(t *testing.T) {
			sc := sc0.clone()
			sc.Field("a", ByName("abc"), Optional(true))
			meet := false
			sc.Consumer().Dependencies().Iterate(func(dep Dependency) bool {
				if dep.Name() == "abc" {
					meet = true
					assert.True(t, dep.Optional())
				}
				return true
			})
			assert.True(t, meet)
		})

		t.Run("AsCollector", func(t *testing.T) {
			sc := sc0.clone()
			sc.Field("c", ByName("cde"), AsCollector(true))
			meet := false
			sc.Consumer().Dependencies().Iterate(func(dep Dependency) bool {
				if dep.Name() == "cde" {
					meet = true
					assert.True(t, dep.IsCollector())
				}
				return true
			})
			assert.True(t, meet)
		})

		t.Run("Name", func(t *testing.T) {
			sc := sc0.clone()
			sc.Field("a", ByName("abc"))
			meet := false
			sc.Consumer().Dependencies().Iterate(func(dep Dependency) bool {
				if dep.Name() == "abc" {
					meet = true
				}
				return true
			})
			assert.True(t, meet)
		})

		t.Run("Tags", func(t *testing.T) {
			sc := sc0.clone()
			sc.Field("a", ByName("abc"), ByTags(tag1))
			meet := false
			sc.Consumer().Dependencies().Iterate(func(dep Dependency) bool {
				if dep.Name() == "abc" {
					meet = true
					assert.Equal(t, newSymbolSet(tag1), dep.Tags())
				}
				return true
			})
			assert.True(t, meet)
		})

		t.Run("field not exist", func(t *testing.T) {
			sc := sc0.clone()
			sc.Field("d", ByName("def"), ByTags(tag1))
			err := sc.Consumer().Validate()
			assert.NotNil(t, err)
		})

		t.Run("nil option", func(t *testing.T) {
			sc := sc0.clone()
			sc.Field("a", nil)
			err := sc.Consumer().Validate()
			assert.Nil(t, err)
		})
	})

	t.Run("IgnoreFields", func(t *testing.T) {
		t.Run("ignore fields", func(t *testing.T) {
			sc := sc0.clone()
			sc.IgnoreFields(func(field reflect.StructField) bool {
				return field.Name == "B"
			})
			con := sc.Consumer()
			con.Dependencies().Iterate(func(dep Dependency) bool {
				if dep.Type() == TypeOf("") {
					assert.Fail(t, "should not reach here")
				}
				return true
			})
		})

		t.Run("predicate is nil", func(t *testing.T) {
			sc := sc0.clone()
			sc.IgnoreFields(nil)
			assert.True(t, sc.Equal(sc0))
		})
	})

	t.Run("SetScope", func(t *testing.T) {
		t.Run("set scope", func(t *testing.T) {
			scope1 := NewScope("scope1")

			sc := sc0.clone()
			sc.SetScope(scope1)
			con := sc.Consumer()
			assert.Equal(t, scope1, con.Scope())
		})

		t.Run("set nil", func(t *testing.T) {
			sc := sc0.clone()
			sc.SetScope(nil)
			con := sc.Consumer()
			assert.Equal(t, GlobalScope, con.Scope())
		})
	})

	t.Run("SetLocation", func(t *testing.T) {
		sc := sc0.clone()
		loc := location.GetCallLocation(0)
		sc.SetLocation(loc)
		con := sc.Consumer()
		assert.Equal(t, loc, con.Location())
	})

	t.Run("UpdateCallLocation", func(t *testing.T) {
		t.Run("location have been set", func(t *testing.T) {
			loc1 := location.GetCallLocation(0)
			sc := sc0.clone()
			sc.SetLocation(loc1)
			sc.UpdateCallLocation(nil)
			assert.Equal(t, loc1, sc.Location())
		})

		t.Run("location have not been set", func(t *testing.T) {
			sc := sc0.clone()
			baseLoc := location.GetCallLocation(0)
			func() {
				sc.UpdateCallLocation(nil)
			}()
			assert.Equal(t, baseLoc.FileName(), sc.Location().FileName())
			assert.Equal(t, baseLoc.FileLine()+3, sc.Location().FileLine())
		})

		t.Run("location is not nil", func(t *testing.T) {
			loc1 := location.GetCallLocation(0)
			sc := sc0.clone()
			sc.UpdateCallLocation(loc1)
			assert.Equal(t, loc1, sc.Location())
		})
	})

	t.Run("Consumer", func(t *testing.T) {
		sc := sc0.clone()
		loc := location.GetCallLocation(0)
		sc.Field("a", ByName("abc"), ByTags(tag1), Optional(true))
		sc.Field("c", ByName("cbd"), AsCollector(true))
		sc.IgnoreFields(func(field reflect.StructField) bool {
			return field.Name == "B"
		})
		sc.SetLocation(loc)
		con := sc.Consumer()

		deps := dependencyIteratorToArray(con.Dependencies())
		assert.Equal(t, 2, len(deps))
		for _, dep := range deps {
			if dep.Name() == "abc" {
				assert.Equal(t, TypeOf(1), dep.Type())
				assert.Equal(t, newSymbolSet(tag1), dep.Tags())
				assert.True(t, dep.Optional())
				assert.Equal(t, valuer.Field("a"), dep.Valuer())
			} else {
				assert.Equal(t, TypeOf(0), dep.Type())
				assert.Equal(t, "cbd", dep.Name())
				assert.Equal(t, valuer.Field("c"), dep.Valuer())
				assert.True(t, dep.IsCollector())
			}

			assert.Same(t, con, dep.Consumer())
		}

		assert.Equal(t, valuer.Struct(TypeOf(testStruct{})), con.Valuer())
		assert.Equal(t, loc, con.Location())
	})
}

func Test_structConsumer_clone(t *testing.T) {
	//lint:ignore U1000 we need the field name to locate the field
	type testStruct struct {
		a int
		B string
		c []int
		D rune
	}

	scope1 := NewScope("scope1")
	tag1 := NewSymbol("tag1")
	loc := location.GetCallLocation(0)

	sc := structConsumerOf(TypeOf(testStruct{}))
	sc.Field("a", ByName("a1"), ByTags(tag1), Optional(true))
	sc.Field("c", ByName("c1"), AsCollector(true))
	sc.Field("D", ByName("D1"))
	sc.IgnoreFields(func(field reflect.StructField) bool {
		return field.Name == "B"
	})
	sc.SetLocation(loc)
	sc.SetScope(scope1)

	verifyConsumer := func(t *testing.T, con Consumer) {
		deps := dependencyIteratorToArray(con.Dependencies())
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
			} else {
				assert.Equal(t, TypeOf('a'), dep.Type())
				assert.Equal(t, "D1", dep.Name())
				assert.Equal(t, valuer.Field("D"), dep.Valuer())
			}

			assert.Same(t, con, dep.Consumer())
		}

		assert.Equal(t, valuer.Struct(TypeOf(testStruct{})), con.Valuer())
		assert.Equal(t, loc, con.Location())
		assert.Equal(t, scope1, con.Scope())
	}

	t.Run("equality", func(t *testing.T) {
		sc2 := sc.clone()
		verifyConsumer(t, sc2.Consumer())

		assert.False(t, sc2.Valuer() == sc.Valuer())
	})

	t.Run("update isolation", func(t *testing.T) {
		scope2 := NewScope("scope2")
		tag2 := NewSymbol("tag2")
		loc2 := location.GetCallLocation(0)

		sc2 := sc.clone()
		sc2.Field("a", ByName("a2"), ByTags(tag2), Optional(false))
		sc2.Field("c", ByName("c2"), AsCollector(false))
		sc2.IgnoreFields(func(field reflect.StructField) bool {
			return field.Name == "D"
		})
		sc2.SetLocation(loc2)
		sc2.SetScope(scope2)

		verifyConsumer(t, sc.Consumer())
	})

	t.Run("update isolation 2", func(t *testing.T) {
		scope2 := NewScope("scope2")
		tag2 := NewSymbol("tag2")
		loc2 := location.GetCallLocation(0)

		sc2 := sc.clone()
		sc3 := sc2.clone()

		sc2.Field("a", ByName("a2"), ByTags(tag2), Optional(false))
		sc2.Field("c", ByName("c2"), AsCollector(false))
		sc2.IgnoreFields(func(field reflect.StructField) bool {
			return field.Name == "D"
		})
		sc2.SetLocation(loc2)
		sc2.SetScope(scope2)

		verifyConsumer(t, sc3.Consumer())
	})

	t.Run("nil", func(t *testing.T) {
		var sc2 *structConsumer
		assert.Nil(t, sc2.clone())
	})
}

func Test_structConsumer_Equal(t *testing.T) {
	//lint:ignore U1000 we need the field name to locate the field
	type testStruct struct {
		a int
		B string
		c []int
	}

	tag1 := NewSymbol("tag1")
	loc1 := location.GetCallLocation(0)
	sc := structConsumerOf(TypeOf(testStruct{}),
		Field("a", ByName("abc"), ByTags(tag1), Optional(true)),
		Field("c", ByName("cbd"), AsCollector(true)),
		IgnoreFields(func(field reflect.StructField) bool {
			return field.Name == "B"
		}),
		Location(loc1),
	)

	t.Run("equal", func(t *testing.T) {
		sc2 := sc.clone()
		assert.True(t, sc.Equal(sc2))
	})

	t.Run("nil equal nil", func(t *testing.T) {
		var sc1 *structConsumer
		var sc2 *structConsumer
		assert.True(t, sc1.Equal(sc2))
	})

	t.Run("not equal to non structConsumer", func(t *testing.T) {
		assert.False(t, sc.Equal(123))
	})

	t.Run("type", func(t *testing.T) {
		sc2 := sc.clone()
		sc2.sType = TypeOf(0)
		assert.False(t, sc.Equal(sc2))
	})

	t.Run("fields", func(t *testing.T) {
		t.Run("len of fields", func(t *testing.T) {
			sc2 := sc.clone()
			sc2.fields["d"] = &structField{}
			assert.False(t, sc.Equal(sc2))
		})

		t.Run("field with same name", func(t *testing.T) {
			sc2 := sc.clone()
			sc2.Field("a", ByName("def"))
			assert.False(t, sc.Equal(sc2))
		})

		t.Run("field name not equal", func(t *testing.T) {
			sc2 := sc.clone()
			sc2.fields["d"] = &structField{}
			sc3 := sc.clone()
			sc3.fields["e"] = &structField{}
			assert.False(t, sc3.Equal(sc2))
		})

	})

	t.Run("fakeFields", func(t *testing.T) {
		t.Run("len of fakeFields", func(t *testing.T) {
			sc2 := sc.clone()
			sc2.Field("d", ByName("def"))
			assert.False(t, sc.Equal(sc2))
		})

		t.Run("fakeFields with same name", func(t *testing.T) {
			sc2 := sc.clone()
			sc2.Field("d", ByName("abc"))
			sc3 := sc.clone()
			sc3.Field("d", ByName("def"))
			assert.False(t, sc3.Equal(sc2))
		})

		t.Run("field name not equal", func(t *testing.T) {
			sc2 := sc.clone()
			sc2.Field("d", ByName("abc"))
			sc3 := sc.clone()
			sc3.Field("e", ByName("abc"))
			assert.False(t, sc3.Equal(sc2))
		})
	})

	t.Run("baseConsumer", func(t *testing.T) {
		t.Run("not nil", func(t *testing.T) {
			sc2 := sc.clone()
			sc2.SetLocation(location.GetCallLocation(0))
			assert.False(t, sc.Equal(sc2))
		})

		t.Run("nil", func(t *testing.T) {
			sc2 := sc.clone()
			sc2.baseConsumer = nil
			assert.False(t, sc2.Equal(sc))
			assert.False(t, sc.Equal(sc2))

			sc3 := sc.clone()
			sc3.baseConsumer = nil
			assert.True(t, sc3.Equal(sc2))
		})

	})
}
