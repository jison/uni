package model

import (
	"fmt"
	"github.com/jison/uni/internal/errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testModuleIterator(t *testing.T, it ModuleIterator, modules []Module) {
	t.Run("iterate", func(t *testing.T) {
		m1 := map[Module]struct{}{}
		m2 := map[Module]struct{}{}

		for _, m := range modules {
			m1[m] = struct{}{}
		}

		r := it.Iterate(func(m Module) bool {
			m2[m] = struct{}{}
			return true
		})

		assert.True(t, r)
		assert.Equal(t, m1, m2)
	})

	t.Run("interrupt", func(t *testing.T) {
		if len(modules) == 0 {
			n := 0
			r := it.Iterate(func(_ Module) bool {
				n += 1
				return false
			})
			assert.True(t, r)
			assert.Equal(t, 0, n)
		} else {
			var half []Module
			r := it.Iterate(func(m Module) bool {
				half = append(half, m)
				return len(half) < len(modules)/2
			})

			assert.False(t, r)

			expected := len(modules) / 2
			if expected == 0 {
				expected = 1
			}
			assert.Equal(t, expected, len(half))
		}
	})
}

func testProviderIterator(t *testing.T, it ProviderIterator, providers []Provider) {
	t.Run("iterate", func(t *testing.T) {
		m1 := map[Provider]struct{}{}
		m2 := map[Provider]struct{}{}

		for _, p := range providers {
			m1[p] = struct{}{}
		}

		r := it.Iterate(func(p Provider) bool {
			m2[p] = struct{}{}
			return true
		})

		assert.True(t, r)
		assert.Equal(t, m1, m2)
	})

	t.Run("interrupt", func(t *testing.T) {
		if len(providers) == 0 {
			n := 0
			r := it.Iterate(func(_ Provider) bool {
				n += 1
				return false
			})
			assert.True(t, r)
			assert.Equal(t, 0, n)
		} else {
			var half []Provider
			r := it.Iterate(func(p Provider) bool {
				half = append(half, p)
				return len(half) < len(providers)/2
			})

			assert.False(t, r)

			expected := len(providers) / 2
			if expected == 0 {
				expected = 1
			}
			assert.Equal(t, expected, len(half))
		}
	})
}

func Test_moduleSet(t *testing.T) {
	m1 := NewModule()
	m2 := NewModule()
	m3 := NewModule()

	tests := []struct {
		name string
		set  moduleSet
		want []Module
	}{
		{"nil", nil, []Module{}},
		{"0", moduleSet{}, []Module{}},
		{"1", moduleSet{m1: {}}, []Module{m1}},
		{"n", moduleSet{m1: {}, m2: {}, m3: {}}, []Module{m1, m2, m3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testModuleIterator(t, tt.set, tt.want)
		})
	}
}

func Test_providerSet(t *testing.T) {
	p1 := Value(1).Provider()
	p2 := Value(2).Provider()
	p3 := Value(3).Provider()

	tests := []struct {
		name string
		set  providerSet
		want []Provider
	}{
		{"nil", nil, []Provider{}},
		{"1", providerSet{}, []Provider{}},
		{"2", providerSet{p1: {}}, []Provider{p1}},
		{"n", providerSet{p1: {}, p2: {}, p3: {}}, []Provider{p1, p2, p3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testProviderIterator(t, tt.set, tt.want)
		})
	}
}

func Test_module_Module(t *testing.T) {
	type testStruct struct {
		a int
		b string
	}
	testFunc := func(a int, b string) (*testStruct, error) {
		return &testStruct{a, b}, nil
	}

	m := newModule(nil, []ProviderBuilder{
		Value(1, Name("value")),
		Struct(testStruct{}, Name("struct")),
		Func(testFunc, Return(0, Name("return"))),
		nil,
	})

	t.Run("Components", func(t *testing.T) {
		t.Run("components from providers", func(t *testing.T) {
			coms := m.AllComponents().ToArray()
			assert.Equal(t, 3, len(coms))
			meetValue := false
			meetStruct := false
			meetFunc := false
			for _, com := range coms {
				if com.Name() == "value" {
					meetValue = true
					assert.Equal(t, TypeOf(0), com.Type())
				} else if com.Name() == "struct" {
					meetStruct = true
					assert.Equal(t, TypeOf(testStruct{}), com.Type())
				} else if com.Name() == "return" {
					meetFunc = true
					assert.Equal(t, TypeOf(&testStruct{}), com.Type())
				}
			}
			assert.True(t, meetValue)
			assert.True(t, meetStruct)
			assert.True(t, meetFunc)
		})

		t.Run("components from sub module", func(t *testing.T) {
			m2 := newModule([]Module{m}, nil)
			coms := m2.AllComponents().ToArray()
			assert.Equal(t, 3, len(coms))
			meetValue := false
			meetStruct := false
			meetFunc := false
			for _, com := range coms {
				if com.Name() == "value" {
					meetValue = true
					assert.Equal(t, TypeOf(0), com.Type())
				} else if com.Name() == "struct" {
					meetStruct = true
					assert.Equal(t, TypeOf(testStruct{}), com.Type())
				} else if com.Name() == "return" {
					meetFunc = true
					assert.Equal(t, TypeOf(&testStruct{}), com.Type())
				}
			}
			assert.True(t, meetValue)
			assert.True(t, meetStruct)
			assert.True(t, meetFunc)
		})
	})
}

func moduleIteratorToSet(mi ModuleIterator) map[Module]struct{} {
	ms := map[Module]struct{}{}
	mi.Iterate(func(m Module) bool {
		ms[m] = struct{}{}
		return true
	})
	return ms
}

func providerIteratorToSet(pi ProviderIterator) map[Provider]struct{} {
	ps := map[Provider]struct{}{}
	pi.Iterate(func(p Provider) bool {
		ps[p] = struct{}{}
		return true
	})
	return ps
}

func Test_module_SubModules(t *testing.T) {
	t.Run("0", func(t *testing.T) {
		m1 := newModule(nil, nil)
		ms := moduleIteratorToSet(m1.SubModules())
		assert.Equal(t, 0, len(ms))
	})

	t.Run("1", func(t *testing.T) {
		m1 := newModule(nil, nil)
		m2 := newModule([]Module{m1}, nil)
		ms := moduleIteratorToSet(m2.SubModules())
		assert.Equal(t, 1, len(ms))
		assert.Equal(t, map[Module]struct{}{m1: {}}, ms)
	})

	t.Run("n", func(t *testing.T) {
		m1 := newModule(nil, nil)
		m2 := newModule(nil, nil)
		m3 := newModule(nil, nil)
		m4 := newModule([]Module{m1, m2, m3}, nil)
		ms := moduleIteratorToSet(m4.SubModules())
		assert.Equal(t, 3, len(ms))
		assert.Equal(t, map[Module]struct{}{m1: {}, m2: {}, m3: {}}, ms)
	})
}

func Test_module_Providers(t *testing.T) {
	t.Run("no provider", func(t *testing.T) {
		m1 := newModule(nil, nil)
		ps := providerIteratorToSet(m1.Providers())
		assert.Equal(t, 0, len(ps))
	})

	t.Run("provider from the module itself", func(t *testing.T) {
		pb1 := Value(123, Name("name1"))
		pb2 := Value(456, Name("name2"))

		m1 := newModule(nil, []ProviderBuilder{
			pb1, pb2,
		})
		ps := providerIteratorToSet(m1.Providers())
		assert.Equal(t, 2, len(ps))

		ps2 := []Provider{pb1.Provider(), pb2.Provider()}

		runCount := 0
		m1.Providers().Iterate(func(p Provider) bool {
			runCount += 1
			found := false
			for _, p2 := range ps2 {
				if p.Equal(p2) {
					found = true
					break
				}
			}
			assert.True(t, found)
			return true
		})
		assert.Equal(t, 2, runCount)
	})

	t.Run("provider from sub modules", func(t *testing.T) {
		pb1 := Value(123, Name("name1"))
		pb2 := Value(456, Name("name2"))
		pb3 := Value(789, Name("name3"))
		pb4 := Value(987, Name("name4"))

		m1 := newModule(nil, []ProviderBuilder{
			pb1, pb2,
		})
		m2 := newModule([]Module{m1}, []ProviderBuilder{
			pb3, pb4,
		})

		ps := providerIteratorToSet(m2.Providers())
		assert.Equal(t, 2, len(ps))

		ps2 := []Provider{pb3.Provider(), pb4.Provider()}

		runCount := 0
		m2.Providers().Iterate(func(p Provider) bool {
			runCount += 1
			found := false
			for _, p2 := range ps2 {
				if p.Equal(p2) {
					found = true
					break
				}
			}
			assert.True(t, found)
			return true
		})
		assert.Equal(t, 2, runCount)
	})
}

func Test_module_AllModules(t *testing.T) {
	t.Run("all modules", func(t *testing.T) {
		m1 := newModule(nil, nil)
		m2 := newModule([]Module{m1}, nil)
		m3 := newModule([]Module{m2}, nil)

		ms := moduleSet{m1: {}, m2: {}, m3: {}}
		assert.Equal(t, ms, m3.AllModules())
	})
}

func Test_module_AllProviders(t *testing.T) {
	t.Run("all providers", func(t *testing.T) {
		pb1 := Value(12, Name("name1"))
		pb2 := Value(34, Name("name2"))
		pb3 := Value(56, Name("name3"))
		pb4 := Value(78, Name("name4"))

		m1 := newModule(nil, []ProviderBuilder{
			pb1, pb2,
		})
		m2 := newModule([]Module{m1}, []ProviderBuilder{
			pb3, pb4,
		})

		ps := providerIteratorToSet(m2.Providers())
		assert.Equal(t, 2, len(ps))

		ps2 := []Provider{pb1.Provider(), pb2.Provider(), pb3.Provider(), pb4.Provider()}

		runCount := 0
		m2.AllProviders().Iterate(func(p Provider) bool {
			runCount += 1
			found := false
			for _, p2 := range ps2 {
				if p.Equal(p2) {
					found = true
					break
				}
			}
			assert.True(t, found)
			return true
		})
		assert.Equal(t, len(ps2), runCount)
	})
}

func Test_module_AllComponents(t *testing.T) {
	pb1 := Value(12, Name("name1"))
	pb2 := Value(34, Name("name2"))
	pb3 := Value(56, Name("name3"))
	pb4 := Value(78, Name("name4"))

	m1 := newModule(nil, []ProviderBuilder{
		pb1, pb2,
	})
	m2 := newModule([]Module{m1}, []ProviderBuilder{
		pb3, pb4,
	})

	cs := newComponentSet()
	ps := []Provider{pb1.Provider(), pb2.Provider(), pb3.Provider(), pb4.Provider()}
	for _, p := range ps {
		p.Components().Each(func(c Component) {
			cs.Add(c)
		})
	}

	cs2 := m2.AllComponents().ToSet()
	assert.Equal(t, cs.Len(), cs2.Len())

	cs2.Each(func(c Component) {
		found := !cs.Iterate(func(c2 Component) bool {
			return !c2.Equal(c)
		})
		assert.True(t, found)
	})
}

func Test_module_Validate(t *testing.T) {
	type testStruct struct {
		a int
		b string
	}
	testFunc := func(a int, b string) (*testStruct, error) {
		return &testStruct{a, b}, nil
	}

	t.Run("no error", func(t *testing.T) {
		m := NewModule(
			Value(1, Name("value")),
			Struct(testStruct{}, Name("struct")),
			Func(testFunc, Return(0, Name("return"))),
		)
		err := m.Validate()
		assert.Nil(t, err)
	})

	t.Run("errors in provider with no component", func(t *testing.T) {
		m := NewModule(
			Value(1, Name("value")),
			Struct(testStruct{}, Name("struct")),
			Func(testFunc, Return(0, Name("return"))),
			Func(func() {}),
		)
		err := m.Validate()
		assert.NotNil(t, err)
	})

	t.Run("errors in provider", func(t *testing.T) {
		type testInterface interface {
			a()
		}

		m := NewModule(
			Value(1, Name("value")),
			Struct(testStruct{}, Name("struct")),
			Func(testFunc, Return(0, Name("return"))),
			Value(errors.Newf("this is an error")),
			Value("a", As((*testInterface)(nil))),
			Func(func() {}),
		)
		err := m.Validate()
		fmt.Printf("%+v\n", err)
		assert.NotNil(t, err)
	})

	t.Run("duplicate component name with same type", func(t *testing.T) {
		m := NewModule(
			Value(1, Name("name1")),
			Value(2, Name("name1")),
			Struct(testStruct{}, Name("struct")),
			Func(testFunc, Return(0, Name("return"))),
			Func(func() {}),
		)
		err := m.Validate()
		fmt.Printf("%+v\n", err)
		assert.NotNil(t, err)
	})

	t.Run("many errors", func(t *testing.T) {
		type testInterface interface {
			a()
		}
		tag1 := NewSymbol("tag1")

		m := NewModule(
			Value(1, Name("name1")),
			Value(2, Name("name1")),
			Struct(testStruct{},
				Field("c", ByName("c field")),
				Name("struct"),
				As((*testInterface)(nil)),
			),
			Func(testFunc,
				Param(0, ByTags(tag1)),
				Return(0, Name("return"), As((*error)(nil))),
				Return(1, As((*error)(nil))),
			),

			Value(errors.Newf("this is an error")),
			Value("a", As((*testInterface)(nil))),
			Func(func() {}),
		)
		err := m.Validate()
		fmt.Printf("%+v\n", err)
		assert.NotNil(t, err)
	})
}

func TestModuleBuilder(t *testing.T) {
	type testStruct struct {
		a int
		b string
	}
	testFunc := func(a int, b string) (*testStruct, error) {
		return &testStruct{a, b}, nil
	}

	t.Run("AddProvider", func(t *testing.T) {
		mb := NewModuleBuilder()
		mb.AddProvider(Value(1, Name("value")))
		mb.AddProvider(Struct(testStruct{}, Name("struct")))
		mb.AddProvider(Func(testFunc, Return(0, Name("return"))))
		mb.AddProvider(nil)

		m := mb.Module()
		coms := m.AllComponents().ToArray()
		assert.Equal(t, 3, len(coms))
		meetValue := false
		meetStruct := false
		meetFunc := false
		for _, com := range coms {
			if com.Name() == "value" {
				meetValue = true
				assert.Equal(t, TypeOf(0), com.Type())
			} else if com.Name() == "struct" {
				meetStruct = true
				assert.Equal(t, TypeOf(testStruct{}), com.Type())
			} else if com.Name() == "return" {
				meetFunc = true
				assert.Equal(t, TypeOf(&testStruct{}), com.Type())
			}
		}
		assert.True(t, meetValue)
		assert.True(t, meetStruct)
		assert.True(t, meetFunc)
	})

	t.Run("AddModule", func(t *testing.T) {
		mb := NewModuleBuilder()
		mb.AddProvider(Value(1, Name("value")))
		mb.AddProvider(Struct(testStruct{}, Name("struct")))
		mb.AddProvider(Func(testFunc, Return(0, Name("return"))))
		m := mb.Module()

		mb2 := NewModuleBuilder()
		mb2.AddModule(m)
		mb2.AddModule(nil)
		m2 := mb2.Module()

		coms := m2.AllComponents().ToArray()
		assert.Equal(t, 3, len(coms))
		meetValue := false
		meetStruct := false
		meetFunc := false
		for _, com := range coms {
			if com.Name() == "value" {
				meetValue = true
				assert.Equal(t, TypeOf(0), com.Type())
			} else if com.Name() == "struct" {
				meetStruct = true
				assert.Equal(t, TypeOf(testStruct{}), com.Type())
			} else if com.Name() == "return" {
				meetFunc = true
				assert.Equal(t, TypeOf(&testStruct{}), com.Type())
			}
		}
		assert.True(t, meetValue)
		assert.True(t, meetStruct)
		assert.True(t, meetFunc)
	})

	t.Run("Module", func(t *testing.T) {
		mb := NewModuleBuilder()
		mb.AddProvider(Value(1, Name("value")))
		mb.AddProvider(Struct(testStruct{}, Name("struct")))
		mb.AddProvider(Func(testFunc, Return(0, Name("return"))))
		m := mb.Module()

		mb2 := NewModuleBuilder()
		mb2.AddModule(m)
		mb2.AddProvider(Value("abc", Name("string")))
		mb2.AddModule(nil)
		m2 := mb2.Module()

		coms := m2.AllComponents().ToArray()
		assert.Equal(t, 4, len(coms))
		meetValue := false
		meetStruct := false
		meetFunc := false
		meetString := false
		for _, com := range coms {
			if com.Name() == "value" {
				meetValue = true
				assert.Equal(t, TypeOf(0), com.Type())
			} else if com.Name() == "struct" {
				meetStruct = true
				assert.Equal(t, TypeOf(testStruct{}), com.Type())
			} else if com.Name() == "return" {
				meetFunc = true
				assert.Equal(t, TypeOf(&testStruct{}), com.Type())
			} else if com.Name() == "string" {
				meetString = true
				assert.Equal(t, TypeOf(""), com.Type())
			}
		}
		assert.True(t, meetValue)
		assert.True(t, meetStruct)
		assert.True(t, meetFunc)
		assert.True(t, meetString)
	})
}

func TestNewModule(t *testing.T) {
	type testStruct struct {
		a int
		b string
	}
	testFunc := func(a int, b string) (*testStruct, error) {
		return &testStruct{a, b}, nil
	}

	t.Run("ProviderBuilder", func(t *testing.T) {
		m := NewModule(
			Value(1, Name("value")),
			Struct(testStruct{}, Name("struct")),
			Func(testFunc, Return(0, Name("return"))),
			nil,
		)

		coms := m.AllComponents().ToArray()
		assert.Equal(t, 3, len(coms))
		meetValue := false
		meetStruct := false
		meetFunc := false
		for _, com := range coms {
			if com.Name() == "value" {
				meetValue = true
				assert.Equal(t, TypeOf(0), com.Type())
			} else if com.Name() == "struct" {
				meetStruct = true
				assert.Equal(t, TypeOf(testStruct{}), com.Type())
			} else if com.Name() == "return" {
				meetFunc = true
				assert.Equal(t, TypeOf(&testStruct{}), com.Type())
			}
		}
		assert.True(t, meetValue)
		assert.True(t, meetStruct)
		assert.True(t, meetFunc)
	})

	t.Run("Provide", func(t *testing.T) {
		m := NewModule(
			Provide(Value(1, Name("value"))),
			Provide(Struct(testStruct{}, Name("struct"))),
			Provide(Func(testFunc, Return(0, Name("return")))),
			Provide(nil),
		)

		coms := m.AllComponents().ToArray()
		assert.Equal(t, 3, len(coms))
		meetValue := false
		meetStruct := false
		meetFunc := false
		for _, com := range coms {
			if com.Name() == "value" {
				meetValue = true
				assert.Equal(t, TypeOf(0), com.Type())
			} else if com.Name() == "struct" {
				meetStruct = true
				assert.Equal(t, TypeOf(testStruct{}), com.Type())
			} else if com.Name() == "return" {
				meetFunc = true
				assert.Equal(t, TypeOf(&testStruct{}), com.Type())
			}
		}
		assert.True(t, meetValue)
		assert.True(t, meetStruct)
		assert.True(t, meetFunc)
	})

	t.Run("SubModule", func(t *testing.T) {
		m := NewModule(
			Value(1, Name("value")),
			Struct(testStruct{}, Name("struct")),
			Func(testFunc, Return(0, Name("return"))),
			nil,
		)
		m2 := NewModule(SubModule(m))

		coms := m2.AllComponents().ToArray()
		assert.Equal(t, 3, len(coms))
		meetValue := false
		meetStruct := false
		meetFunc := false
		for _, com := range coms {
			if com.Name() == "value" {
				meetValue = true
				assert.Equal(t, TypeOf(0), com.Type())
			} else if com.Name() == "struct" {
				meetStruct = true
				assert.Equal(t, TypeOf(testStruct{}), com.Type())
			} else if com.Name() == "return" {
				meetFunc = true
				assert.Equal(t, TypeOf(&testStruct{}), com.Type())
			}
		}
		assert.True(t, meetValue)
		assert.True(t, meetStruct)
		assert.True(t, meetFunc)
	})
}

type nilProviderBuilder struct{}

func (n nilProviderBuilder) Provider() Provider {
	return nil
}

func Test_newModule(t *testing.T) {
	type testStruct struct {
		a int
		b string
	}
	testFunc := func(a int, b string) (*testStruct, error) {
		return &testStruct{a, b}, nil
	}

	t.Run("newModule", func(t *testing.T) {
		m := NewModule(
			Value(1, Name("value")),
			Struct(testStruct{}, Name("struct")),
			nil,
		)
		m2 := newModule(
			[]Module{m, nil},
			[]ProviderBuilder{
				Func(testFunc, Return(0, Name("return"))),
				nilProviderBuilder{},
				nil,
			},
		)
		coms := m2.AllComponents().ToArray()
		assert.Equal(t, 3, len(coms))
		meetValue := false
		meetStruct := false
		meetFunc := false
		for _, com := range coms {
			if com.Name() == "value" {
				meetValue = true
				assert.Equal(t, TypeOf(0), com.Type())
			} else if com.Name() == "struct" {
				meetStruct = true
				assert.Equal(t, TypeOf(testStruct{}), com.Type())
			} else if com.Name() == "return" {
				meetFunc = true
				assert.Equal(t, TypeOf(&testStruct{}), com.Type())
			}
		}
		assert.True(t, meetValue)
		assert.True(t, meetStruct)
		assert.True(t, meetFunc)
	})
}
