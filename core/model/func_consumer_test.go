package model

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jison/uni/core/valuer"
	"github.com/jison/uni/internal/location"
	"github.com/stretchr/testify/assert"
)

type testStructForFuncConsumerTest struct {
	a int
	b string
	c []int
	d rune
}

func funcForFuncConsumerTest(a int, b string, c []int, d rune) (*testStructForFuncConsumerTest, error) {
	return &testStructForFuncConsumerTest{a, b, c, d}, nil
}

func Test_funcParam_Format(t *testing.T) {
	fc := funcConsumerOf(funcForFuncConsumerTest)
	fp := fc.params[0]

	t.Run("no verbose", func(t *testing.T) {
		expected := fmt.Sprintf("%+v at parameter `%d`", fp.dependency, fp.index)
		assert.Equal(t, expected, fmt.Sprintf("%v", fp))
	})

	t.Run("verbose", func(t *testing.T) {
		expected := fmt.Sprintf("%+v at parameter `%d` of %+v", fp.dependency, fp.index, fp.Consumer())
		assert.Equal(t, expected, fmt.Sprintf("%+v", fp))
	})
}

func Test_funcParam_clone(t *testing.T) {
	fc := funcConsumerOf(funcForFuncConsumerTest)
	fp := fc.params[0]

	t.Run("equality", func(t *testing.T) {
		fp2 := fp.clone()
		assert.True(t, fp2.Equal(fp))
		assert.NotSame(t, fp2, fp)
	})

	t.Run("update isolation", func(t *testing.T) {
		fp2 := fp.clone()
		fp2.index = 2
		fp2.dependency.optional = true

		assert.Equal(t, 0, fp.index)
		assert.Equal(t, false, fp.optional)
	})

	t.Run("nil", func(t *testing.T) {
		var fp2 *funcParam
		assert.Nil(t, fp2.clone())
	})
}

func Test_funcParam_Equal(t *testing.T) {
	fc := funcConsumerOf(funcForFuncConsumerTest)
	fp := fc.params[0]

	t.Run("equal", func(t *testing.T) {
		fp2 := fp.clone()
		assert.True(t, fp2.Equal(fp))
	})

	t.Run("not equal to non funcParam", func(t *testing.T) {
		assert.False(t, fp.Equal(123))
	})

	t.Run("nil equal nil", func(t *testing.T) {
		var fp2 *funcParam
		var fp3 *funcParam
		assert.True(t, fp2.Equal(fp3))
	})

	t.Run("index", func(t *testing.T) {
		fp2 := fc.params[1]
		assert.False(t, fp.Equal(fp2))
	})

	t.Run("dependency", func(t *testing.T) {
		t.Run("not nil", func(t *testing.T) {
			fp2 := fp.clone()
			fp2.optional = true
			assert.False(t, fp2.Equal(fp))
		})

		t.Run("nil", func(t *testing.T) {
			fp2 := fp.clone()
			fp2.dependency = nil
			assert.False(t, fp2.Equal(fp))
			assert.False(t, fp.Equal(fp2))

			fp3 := fp.clone()
			fp3.dependency = nil
			assert.True(t, fp3.Equal(fp2))
		})
	})
}

func Test_paramByIndex(t *testing.T) {
	fc := funcConsumerOf(funcForFuncConsumerTest)
	fp1 := fc.params[0]
	fp2 := fc.params[1]
	fp3 := fc.params[2]

	tests := []struct {
		name string
		it   paramByIndex
		want []Dependency
	}{
		{"nil", nil, []Dependency{}},
		{"0", paramByIndex{}, []Dependency{}},
		{"1", paramByIndex{0: fp1}, []Dependency{fp1}},
		{"2", paramByIndex{0: fp1, 1: fp2}, []Dependency{fp1, fp2}},
		{"n", paramByIndex{0: fp1, 1: fp2, 2: fp3}, []Dependency{fp1, fp2, fp3}},
	}

	for _, tt := range tests {
		testDependencyIterator(t, tt.it, tt.want)
	}
}

func TestFuncConsumer(t *testing.T) {
	t.Run("normal function", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		baseLoc := location.GetCallLocation(0)
		fc := FuncConsumer(funcForFuncConsumerTest,
			Param(0, ByName("a1"), ByTags(tag1), Optional(true)),
			Param(1, ByName("b1")),
			Param(2, ByName("c1"), AsCollector(true)),
			Param(3),
			nil,
		)

		con := fc.Consumer()
		err := con.Validate()
		assert.Nil(t, err)

		deps := dependencyIteratorToArray(con.Dependencies())
		assert.Equal(t, 4, len(deps))
		for _, dep := range deps {
			if dep.Name() == "a1" {
				assert.Equal(t, TypeOf(1), dep.Type())
				assert.Equal(t, "a1", dep.Name())
				assert.Equal(t, newSymbolSet(tag1), dep.Tags())
				assert.Equal(t, valuer.Param(0), dep.Valuer())
				assert.True(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			} else if dep.Name() == "b1" {
				assert.Equal(t, TypeOf(""), dep.Type())
				assert.Equal(t, "b1", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Param(1), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			} else if dep.Name() == "c1" {
				assert.Equal(t, TypeOf(0), dep.Type())
				assert.Equal(t, "c1", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Param(2), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.True(t, dep.IsCollector())
			} else {
				assert.Equal(t, TypeOf('d'), dep.Type())
				assert.Equal(t, "", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Param(3), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			}

			assert.Same(t, con, dep.Consumer())
		}

		assert.Equal(t, valuer.Func(reflect.ValueOf(funcForFuncConsumerTest)), con.Valuer())
		assert.Equal(t, baseLoc.FileName(), con.Location().FileName())
		assert.Equal(t, baseLoc.FileLine()+1, con.Location().FileLine())
	})

	t.Run("nil function", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		fc := FuncConsumer(nil,
			Param(0, ByName("a1"), ByTags(tag1), Optional(true)),
			Param(1, ByName("b1")),
			Param(2, ByName("c1"), AsCollector(true)),
			Param(3),
			nil,
		)

		con := fc.Consumer()
		err := con.Validate()
		assert.NotNil(t, err)
	})

	t.Run("variadic function", func(t *testing.T) {
		fc := FuncConsumer(func(a ...int) {

		})
		con := fc.Consumer()
		deps := dependencyIteratorToArray(con.Dependencies())
		assert.Equal(t, 1, len(deps))
		dep := deps[0]
		assert.Equal(t, TypeOf(0), dep.Type())
		assert.True(t, dep.IsCollector())
	})
}

func Test_funcConsumer_Consumer(t *testing.T) {
	tag1 := NewSymbol("tag1")
	loc1 := location.GetCallLocation(0)
	fc := funcConsumerOf(funcForFuncConsumerTest,
		Param(0, ByName("a1"), ByTags(tag1), Optional(true)),
		Param(1, ByName("b1")),
		Param(2, ByName("c1"), AsCollector(true)),
		Param(3),
		Location(loc1),
	)

	t.Run("Dependencies", func(t *testing.T) {
		con := fc.Consumer()
		deps := dependencyIteratorToArray(con.Dependencies())
		assert.Equal(t, 4, len(deps))
		for _, dep := range deps {
			if dep.Name() == "a1" {
				assert.Equal(t, TypeOf(1), dep.Type())
				assert.Equal(t, "a1", dep.Name())
				assert.Equal(t, newSymbolSet(tag1), dep.Tags())
				assert.Equal(t, valuer.Param(0), dep.Valuer())
				assert.True(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			} else if dep.Name() == "b1" {
				assert.Equal(t, TypeOf(""), dep.Type())
				assert.Equal(t, "b1", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Param(1), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			} else if dep.Name() == "c1" {
				assert.Equal(t, TypeOf(0), dep.Type())
				assert.Equal(t, "c1", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Param(2), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.True(t, dep.IsCollector())
			} else {
				assert.Equal(t, TypeOf('d'), dep.Type())
				assert.Equal(t, "", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Param(3), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			}
			assert.True(t, dep.Valuer() == dep.Valuer())
			assert.Same(t, con, dep.Consumer())
		}
	})

	t.Run("Valuer", func(t *testing.T) {
		con := fc.Consumer()
		assert.Equal(t, valuer.Func(reflect.ValueOf(funcForFuncConsumerTest)), con.Valuer())
		assert.True(t, con.Valuer() == con.Valuer())
	})

	t.Run("Scope", func(t *testing.T) {
		t.Run("scope", func(t *testing.T) {
			scope1 := NewScope("scope1")
			fc2 := funcConsumerOf(funcForFuncConsumerTest, InScope(scope1))
			con := fc2.Consumer()
			assert.Equal(t, scope1, con.Scope())
		})

		t.Run("scope", func(t *testing.T) {
			con := fc.Consumer()
			assert.Equal(t, GlobalScope, con.Scope())
		})
	})

	t.Run("Location", func(t *testing.T) {
		loc2 := location.GetCallLocation(0)
		fc2 := funcConsumerOf(funcForFuncConsumerTest, Location(loc2))
		assert.Equal(t, loc2, fc2.Consumer().Location())
	})

	t.Run("UpdateCallLocation", func(t *testing.T) {
		baseLoc := location.GetCallLocation(0)
		var fc2 *funcConsumer
		func() {
			fc2 = funcConsumerOf(funcForFuncConsumerTest, UpdateCallLocation())
		}()

		c := fc2.Consumer()
		assert.Equal(t, baseLoc.FileName(), c.Location().FileName())
		assert.Equal(t, baseLoc.FileLine()+4, c.Location().FileLine())
	})

	t.Run("Validate", func(t *testing.T) {
		t.Run("val is not a function", func(t *testing.T) {
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
				{"*int", TypeOf(&intVal)},
				{"*string", TypeOf(&strVal)},
				{"*rune", TypeOf(&runeVal)},
				{"map[int]string", TypeOf(map[int]string{})},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					err := FuncConsumer(tt.t).Consumer().Validate()
					assert.NotNil(t, err)
				})
			}
		})

		t.Run("parameter not exist", func(t *testing.T) {
			con := FuncConsumer(funcForFuncConsumerTest, Param(4, ByName("p4"))).Consumer()
			err := con.Validate()
			assert.NotNil(t, err)
		})

		t.Run("error at parameters", func(t *testing.T) {
			con := FuncConsumer(funcForFuncConsumerTest, Param(0, AsCollector(true))).Consumer()
			err := con.Validate()
			assert.NotNil(t, err)
		})
	})

	t.Run("Format", func(t *testing.T) {
		con := fc.Consumer()

		t.Run("not verbose", func(t *testing.T) {
			expected := fmt.Sprintf("FunctionConsumer[%v]", TypeOf(funcForFuncConsumerTest))
			assert.Equal(t, expected, fmt.Sprintf("%v", con))
		})

		t.Run("verbose", func(t *testing.T) {
			expected := fmt.Sprintf("FunctionConsumer[%v] at %v", TypeOf(funcForFuncConsumerTest), con.Location())
			assert.Equal(t, expected, fmt.Sprintf("%+v", con))
		})
	})
}

func Test_funcConsumer_StructConsumerBuilder(t *testing.T) {
	fc0 := funcConsumerOf(funcForFuncConsumerTest)

	t.Run("Param", func(t *testing.T) {
		t.Run("Optional", func(t *testing.T) {
			fc := fc0.clone()
			fc.Param(0, ByName("p1"), Optional(true))
			con := fc.Consumer()

			meet := false
			con.Dependencies().Iterate(func(dep Dependency) bool {
				if dep.Name() == "p1" {
					meet = true
					assert.True(t, dep.Optional())
				}

				return true
			})

			assert.True(t, meet)
		})

		t.Run("AsCollector", func(t *testing.T) {
			fc := fc0.clone()
			fc.Param(2, ByName("p2"), AsCollector(true))
			con := fc.Consumer()

			meet := false
			con.Dependencies().Iterate(func(dep Dependency) bool {
				if dep.Name() == "p2" {
					meet = true
					assert.True(t, dep.IsCollector())
				}

				return true
			})

			assert.True(t, meet)
		})

		t.Run("Name", func(t *testing.T) {
			fc := fc0.clone()
			fc.Param(1, ByName("p1"))
			con := fc.Consumer()

			meet := false
			con.Dependencies().Iterate(func(dep Dependency) bool {
				if dep.Name() == "p1" {
					meet = true
					assert.Equal(t, "p1", dep.Name())
				}

				return true
			})

			assert.True(t, meet)
		})

		t.Run("Tags", func(t *testing.T) {
			tag1 := NewSymbol("tag1")
			fc := fc0.clone()
			fc.Param(1, ByName("p1"), ByTags(tag1))
			con := fc.Consumer()

			meet := false
			con.Dependencies().Iterate(func(dep Dependency) bool {
				if dep.Name() == "p1" {
					meet = true
					assert.Equal(t, newSymbolSet(tag1), dep.Tags())
				}

				return true
			})

			assert.True(t, meet)
		})

		t.Run("param not exist", func(t *testing.T) {
			fc := fc0.clone()
			fc.Param(4, ByName("p4"))
			con := fc.Consumer()

			meet := false
			con.Dependencies().Iterate(func(dep Dependency) bool {
				if dep.Name() == "p4" {
					meet = true
				}

				return true
			})

			assert.False(t, meet)
		})

		t.Run("option is nil", func(t *testing.T) {
			fc := fc0.clone()
			fc.Param(1, ByName("p1"), nil)
			con := fc.Consumer()

			meet := false
			con.Dependencies().Iterate(func(dep Dependency) bool {
				if dep.Name() == "p1" {
					meet = true
				}

				return true
			})

			assert.True(t, meet)
		})
	})

	t.Run("SetLocation", func(t *testing.T) {
		loc1 := location.GetCallLocation(0)
		fc := fc0.clone()
		fc.SetLocation(loc1)
		con := fc.Consumer()
		assert.Equal(t, loc1, con.Location())
	})

	t.Run("UpdateCallLocation", func(t *testing.T) {
		t.Run("location have been set", func(t *testing.T) {
			loc1 := location.GetCallLocation(0)
			fc := fc0.clone()
			fc.SetLocation(loc1)
			fc.UpdateCallLocation(nil)
			assert.Equal(t, loc1, fc.Location())
		})

		t.Run("location have not been set", func(t *testing.T) {
			fc := fc0.clone()
			baseLoc := location.GetCallLocation(0)
			func() {
				fc.UpdateCallLocation(nil)
			}()
			assert.Equal(t, baseLoc.FileName(), fc.Location().FileName())
			assert.Equal(t, baseLoc.FileLine()+3, fc.Location().FileLine())
		})

		t.Run("location is not nil", func(t *testing.T) {
			loc1 := location.GetCallLocation(0)
			fc := fc0.clone()
			fc.UpdateCallLocation(loc1)
			assert.Equal(t, loc1, fc.Location())
		})
	})

	t.Run("Consumer", func(t *testing.T) {
		loc1 := location.GetCallLocation(0)
		tag1 := NewSymbol("tag1")
		fc := fc0.clone()
		fc.Param(0, ByName("a1"), ByTags(tag1), Optional(true))
		fc.Param(1, ByName("b1"))
		fc.Param(2, ByName("c1"), AsCollector(true))
		fc.Param(3)
		fc.SetLocation(loc1)

		con := fc.Consumer()
		err := con.Validate()
		assert.Nil(t, err)

		deps := dependencyIteratorToArray(con.Dependencies())
		assert.Equal(t, 4, len(deps))
		for _, dep := range deps {
			if dep.Name() == "a1" {
				assert.Equal(t, TypeOf(1), dep.Type())
				assert.Equal(t, "a1", dep.Name())
				assert.Equal(t, newSymbolSet(tag1), dep.Tags())
				assert.Equal(t, valuer.Param(0), dep.Valuer())
				assert.True(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			} else if dep.Name() == "b1" {
				assert.Equal(t, TypeOf(""), dep.Type())
				assert.Equal(t, "b1", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Param(1), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			} else if dep.Name() == "c1" {
				assert.Equal(t, TypeOf(0), dep.Type())
				assert.Equal(t, "c1", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Param(2), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.True(t, dep.IsCollector())
			} else {
				assert.Equal(t, TypeOf('d'), dep.Type())
				assert.Equal(t, "", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Param(3), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			}

			assert.Same(t, con, dep.Consumer())
		}

		assert.Equal(t, valuer.Func(reflect.ValueOf(funcForFuncConsumerTest)), con.Valuer())
		assert.Equal(t, loc1, con.Location())
	})
}

func Test_funcConsumer_clone(t *testing.T) {
	scope1 := NewScope("scope1")
	loc1 := location.GetCallLocation(0)
	tag1 := NewSymbol("tag1")
	fc := funcConsumerOf(funcForFuncConsumerTest)
	fc.Param(0, ByName("a1"), ByTags(tag1), Optional(true))
	fc.Param(1, ByName("b1"))
	fc.Param(2, ByName("c1"), AsCollector(true))
	fc.Param(3)
	fc.SetLocation(loc1)
	fc.SetScope(scope1)

	verifyConsumer := func(t *testing.T, con Consumer) {
		deps := dependencyIteratorToArray(con.Dependencies())
		assert.Equal(t, 4, len(deps))
		for _, dep := range deps {
			if dep.Name() == "a1" {
				assert.Equal(t, TypeOf(1), dep.Type())
				assert.Equal(t, "a1", dep.Name())
				assert.Equal(t, newSymbolSet(tag1), dep.Tags())
				assert.Equal(t, valuer.Param(0), dep.Valuer())
				assert.True(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			} else if dep.Name() == "b1" {
				assert.Equal(t, TypeOf(""), dep.Type())
				assert.Equal(t, "b1", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Param(1), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			} else if dep.Name() == "c1" {
				assert.Equal(t, TypeOf(0), dep.Type())
				assert.Equal(t, "c1", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Param(2), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.True(t, dep.IsCollector())
			} else {
				assert.Equal(t, TypeOf('d'), dep.Type())
				assert.Equal(t, "", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Param(3), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			}

			assert.Same(t, con, dep.Consumer())
		}

		assert.Equal(t, valuer.Func(reflect.ValueOf(funcForFuncConsumerTest)), con.Valuer())
		assert.Equal(t, loc1, con.Location())
		assert.Equal(t, scope1, con.Scope())
	}

	t.Run("equality", func(t *testing.T) {
		fc2 := fc.clone()
		con2 := fc2.Consumer()
		verifyConsumer(t, con2)
		assert.False(t, fc2.Valuer() == fc.Valuer())
	})

	t.Run("update isolation", func(t *testing.T) {
		scope2 := NewScope("scope2")
		loc2 := location.GetCallLocation(0)
		tag2 := NewSymbol("tag2")

		fc2 := fc.clone()

		fc2.Param(0, ByName("a2"), ByTags(tag2), Optional(false))
		fc2.Param(1, ByName("b2"))
		fc2.Param(2, ByName("c2"), AsCollector(false))
		fc2.Param(3)
		fc2.SetScope(scope2)
		fc2.SetLocation(loc2)

		verifyConsumer(t, fc.Consumer())
	})

	t.Run("update isolation 2", func(t *testing.T) {
		scope2 := NewScope("scope2")
		loc2 := location.GetCallLocation(0)
		tag2 := NewSymbol("tag2")

		fc2 := fc.clone()
		fc3 := fc2.clone()

		fc2.Param(0, ByName("a2"), ByTags(tag2), Optional(false))
		fc2.Param(1, ByName("b2"))
		fc2.Param(2, ByName("c2"), AsCollector(false))
		fc2.Param(3)
		fc2.SetScope(scope2)
		fc2.SetLocation(loc2)

		verifyConsumer(t, fc3.Consumer())
	})

	t.Run("nil", func(t *testing.T) {
		var fc2 *funcConsumer
		assert.Nil(t, fc2.clone())
	})
}

func Test_funcConsumer_Equal(t *testing.T) {
	scope1 := NewScope("scope1")
	loc1 := location.GetCallLocation(0)
	tag1 := NewSymbol("tag1")
	fc := funcConsumerOf(funcForFuncConsumerTest)
	fc.Param(0, ByName("a1"), ByTags(tag1), Optional(true))
	fc.Param(1, ByName("b1"))
	fc.Param(2, ByName("c1"), AsCollector(true))
	fc.Param(3)
	fc.SetLocation(loc1)
	fc.SetScope(scope1)

	t.Run("equal", func(t *testing.T) {
		fc2 := fc.clone()
		assert.True(t, fc2.Equal(fc))
	})

	t.Run("not equal to non funcConsumer", func(t *testing.T) {
		assert.False(t, fc.Equal(123))
	})

	t.Run("nil equal nil", func(t *testing.T) {
		var fc2 *funcConsumer
		var fc3 *funcConsumer
		assert.True(t, fc2.Equal(fc3))
	})

	t.Run("function", func(t *testing.T) {
		fc2 := fc.clone()
		fc2.funcVal = reflect.ValueOf(func() {})
		assert.False(t, fc2.Equal(fc))
	})

	t.Run("param", func(t *testing.T) {
		t.Run("len", func(t *testing.T) {
			fc2 := fc.clone()
			fc2.params[4] = &funcParam{}
			assert.False(t, fc2.Equal(fc))
		})

		t.Run("parameter does not have same index", func(t *testing.T) {
			fc2 := fc.clone()
			fc2.params[4] = fc2.params[3]
			delete(fc2.params, 3)
			assert.False(t, fc2.Equal(fc))
		})

		t.Run("parameter does have same index", func(t *testing.T) {
			fc2 := fc.clone()
			fc2.Param(0, ByName("a2"))
			assert.False(t, fc2.Equal(fc))
			assert.False(t, fc.Equal(fc2))
		})
	})

	t.Run("fakeParam", func(t *testing.T) {
		t.Run("len", func(t *testing.T) {
			fc2 := fc.clone()
			fc2.fakeParams[4] = &funcParam{}
			assert.False(t, fc2.Equal(fc))
		})

		t.Run("parameter does not have same index", func(t *testing.T) {
			fc2 := fc.clone()
			fc2.fakeParams[4] = &funcParam{}
			fc3 := fc.clone()
			fc3.fakeParams[5] = &funcParam{}

			assert.False(t, fc2.Equal(fc3))
		})

		t.Run("parameter does have same index", func(t *testing.T) {
			fc2 := fc.clone()
			fc2.Param(4, ByName("aa"))
			fc3 := fc.clone()
			fc3.Param(4, ByName("bb"))
			assert.False(t, fc2.Equal(fc3))
			assert.False(t, fc3.Equal(fc2))
		})
	})

	t.Run("baseConsumer", func(t *testing.T) {
		t.Run("not nil", func(t *testing.T) {
			fc2 := fc.clone()
			fc2.SetLocation(location.GetCallLocation(0))
			assert.False(t, fc2.Equal(fc))
		})

		t.Run("nil", func(t *testing.T) {
			fc2 := fc.clone()
			fc2.baseConsumer = nil
			assert.False(t, fc2.Equal(fc))
			assert.False(t, fc.Equal(fc2))

			fc3 := fc.clone()
			fc3.baseConsumer = nil
			assert.True(t, fc3.Equal(fc2))
		})
	})
}
