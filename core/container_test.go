package core

import (
	"fmt"
	"github.com/jison/uni/internal/errors"
	"testing"

	"github.com/jison/uni/core/model"
	"github.com/stretchr/testify/assert"
)

func buildModuleForContainerTest() (model.Module, model.Scope, model.Scope) {
	scope1 := model.NewScope("scope1")
	scope2 := model.NewScope("scope2", scope1)

	return model.NewModule(
		model.Value(123, model.Name("name1")),
		model.WithScope(scope1)(
			model.Value("abc", model.Name("name2")),
			model.Struct(testStruct{}, model.Name("name3")),
			model.Func(func(ts testStruct) *testStruct { return &ts },
				model.Return(0, model.Name("name4")),
			),
		),
		model.WithScope(scope2)(
			model.Struct(&testStruct2{}, model.Name("name5"),
				model.Field("ts3", model.ByName("nameNotExist"), model.Optional(true)),
			),
			model.Struct(&testStruct3{}, model.Name("name6")),
			model.Func(
				func(ts *testStruct3) testStruct3 {
					return *ts
				},
				model.Return(0, model.Name("name7")),
			),
		),
	), scope1, scope2
}

func Test_NewContainer(t *testing.T) {
	t.Run("no options", func(t *testing.T) {
		m := model.NewModule(model.Value(123))
		c, err := NewContainer(m)
		assert.Nil(t, err)
		assert.NotNil(t, c)
	})

	t.Run("ignore missing", func(t *testing.T) {
		m := model.NewModule(model.Func(func(int) string { return "" }))
		c, err := NewContainer(m)
		assert.NotNil(t, err)

		c, err2 := NewContainer(m, IgnoreMissing())
		assert.Nil(t, err2)
		assert.NotNil(t, c)
	})

	t.Run("ignore uncertainties", func(t *testing.T) {
		m := model.NewModule(
			model.Value(123),
			model.Value(456),
			model.Func(func(int) string { return "" }),
		)
		c, err := NewContainer(m)
		assert.NotNil(t, err)

		c, err2 := NewContainer(m, IgnoreUncertain())
		assert.Nil(t, err2)
		assert.NotNil(t, c)
	})

	t.Run("ignore cycle", func(t *testing.T) {
		m := model.NewModule(
			model.Func(func(string) int { return 0 }),
			model.Func(func(int) string { return "" }),
		)
		c, err := NewContainer(m)
		assert.NotNil(t, err)

		c, err2 := NewContainer(m, IgnoreCycle())
		assert.Nil(t, err2)
		assert.NotNil(t, c)
	})
}

func Test_newContainer(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		m := model.NewModule(model.Value(123))
		c, err := newContainer(m, nil)
		assert.Nil(t, err)
		assert.NotNil(t, c)
	})

	t.Run("error in module", func(t *testing.T) {
		type testInterface interface {
			a()
		}
		m := model.NewModule(model.Value(123, model.As((*testInterface)(nil))))
		_, err := newContainer(m, nil)
		fmt.Printf("%v\n", err)
		assert.NotNil(t, err)
	})

	t.Run("check dependence", func(t *testing.T) {
		m, _, _, _ := buildTestModule()
		_, err := newContainer(m, nil)
		assert.NotNil(t, err)
	})

	t.Run("no check dependence", func(t *testing.T) {
		m, _, _, _ := buildTestModule()
		_, err := newContainer(m, &ContainerOptions{
			ignoreMissing:   true,
			ignoreUncertain: true,
			ignoreCycle:     true,
		})
		assert.Nil(t, err)
	})
}

func Test_container_Load(t *testing.T) {
	t.Run("load", func(t *testing.T) {
		m, scope1, _ := buildModuleForContainerTest()
		c, err := newContainer(m, nil)
		assert.Nil(t, err)

		c1, _ := c.EnterScope(scope1)
		err = c1.Load(model.NewCriteria(testStruct{}), model.NewCriteria(&testStruct{}))
		assert.Nil(t, err)

		c2 := c1.(*container)

		runCount := 0
		m.AllComponents().Filter(func(com model.Component) bool {
			return com.Name() == "name3" || com.Name() == "name4"
		}).Iterate(func(com model.Component) bool {
			runCount += 1
			_, ok := c2.storage.Get(com.Valuer(), com.Provider().Scope())
			assert.True(t, ok)
			return true
		})

		assert.Equal(t, 2, runCount)
	})

	t.Run("criteria do not match any component in the scope", func(t *testing.T) {
		m, scope1, _ := buildModuleForContainerTest()
		c, err := newContainer(m, nil)
		assert.Nil(t, err)

		c1, _ := c.EnterScope(scope1)
		err = c1.Load(model.NewCriteria(testStruct{}), model.NewCriteria(testStruct2{}))
		assert.NotNil(t, err)
	})
}

func Test_container_LoadAll(t *testing.T) {
	t.Run("load all", func(t *testing.T) {
		m, scope1, _ := buildModuleForContainerTest()
		c, err := newContainer(m, nil)
		assert.Nil(t, err)

		c1, _ := c.EnterScope(scope1)
		err = c1.LoadAll()
		assert.Nil(t, err)

		c2 := c1.(*container)

		runCount := 0
		m.AllComponents().Filter(func(com model.Component) bool {
			return com.Name() == "name3" || com.Name() == "name4"
		}).Iterate(func(com model.Component) bool {
			runCount += 1
			_, ok := c2.storage.Get(com.Valuer(), com.Provider().Scope())
			assert.True(t, ok)
			return true
		})

		assert.Equal(t, 2, runCount)
	})
}

func Test_container_FuncOf(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		m, scope1, scope2 := buildModuleForContainerTest()
		c, _ := newContainer(m, nil)
		c1, _ := c.EnterScope(scope1)
		c2, _ := c1.EnterScope(scope2)

		runCount := 0
		exe := c2.FuncOf(func(a testStruct3, b *testStruct3) (int, error) {
			runCount += 1

			assert.Equal(t, a.a, b.a)
			assert.Equal(t, a.b, b.b)
			assert.Equal(t, a.ts2, b.ts2)
			return a.a, nil
		})

		ret, err := exe.Execute()
		assert.Equal(t, 1, runCount)
		assert.Equal(t, 123, ret.([]interface{})[0])
		assert.Nil(t, err)

		ret, err = exe.Execute()
		assert.Equal(t, 2, runCount)
		assert.Equal(t, 123, ret.([]interface{})[0])
		assert.Nil(t, err)
	})

	t.Run("variadic", func(t *testing.T) {
		m, scope1, scope2 := buildModuleForContainerTest()
		c, _ := newContainer(m, nil)
		c1, _ := c.EnterScope(scope1)
		c2, _ := c1.EnterScope(scope2)

		runCount := 0
		exe := c2.FuncOf(func(a string, b ...int) (string, error) {
			runCount += 1

			assert.Equal(t, "abc", a)
			assert.Equal(t, 1, len(b))
			assert.Equal(t, 123, b[0])
			return a, nil
		}, model.Param(1, model.AsCollector(true)))

		ret, err := exe.Execute()
		assert.Equal(t, 1, runCount)
		assert.Equal(t, "abc", ret.([]interface{})[0])
		assert.Nil(t, err)

		ret, err = exe.Execute()
		assert.Equal(t, 2, runCount)
		assert.Equal(t, "abc", ret.([]interface{})[0])
		assert.Nil(t, err)
	})

	t.Run("function return error", func(t *testing.T) {
		m, scope1, scope2 := buildModuleForContainerTest()
		c, _ := newContainer(m, nil)
		c1, _ := c.EnterScope(scope1)
		c2, _ := c1.EnterScope(scope2)

		err := errors.Newf("this is an error")
		exe := c2.FuncOf(func(a testStruct3, b *testStruct3) (int, error) {
			return 123, err
		})
		_, err2 := exe.Execute()
		assert.NotNil(t, err2)
		assert.True(t, errors.Is(err2, err))
	})

	t.Run("error in dependencies", func(t *testing.T) {
		m, scope1, _ := buildModuleForContainerTest()

		err := errors.Newf("this is an error")
		m2 := model.NewModule(
			model.SubModule(m),
			model.Func(
				func(ts3 testStruct) (*testStruct, error) {
					return &ts3, err
				},
				model.Return(0, model.Name("name8")),
				model.InScope(scope1),
			),
		)

		c, _ := newContainer(m2, nil)
		c1, _ := c.EnterScope(scope1)
		exe := c1.FuncOf(func(ts3 *testStruct) {
			// do nothing
		}, model.Param(0, model.ByName("name8")))

		_, err2 := exe.Execute()
		assert.NotNil(t, err2)
		assert.True(t, errors.Is(err2, err))
	})

	t.Run("error in dependence graph", func(t *testing.T) {
		m, scope1, _ := buildModuleForContainerTest()

		err := errors.Newf("this is an error")
		m2 := model.NewModule(
			model.SubModule(m),
			model.Func(
				func(ts3 testStruct3) (*testStruct, error) {
					return nil, err
				},
				model.Return(0, model.Name("name8")),
				model.InScope(scope1),
			),
		)

		c, _ := newContainer(m2, &ContainerOptions{
			ignoreMissing:   true,
			ignoreUncertain: true,
			ignoreCycle:     true,
		})
		c1, _ := c.EnterScope(scope1)
		exe := c1.FuncOf(func(ts3 *testStruct) {
			// do nothing
		}, model.Param(0, model.ByName("name8")))
		_, err2 := exe.Execute()
		assert.NotNil(t, err2)
		assert.False(t, errors.Is(err2, err))
	})
}

func Test_container_StructOf(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		m, scope1, scope2 := buildModuleForContainerTest()
		c, _ := newContainer(m, nil)
		c1, _ := c.EnterScope(scope1)
		c2, _ := c1.EnterScope(scope2)

		exe := c2.StructOf(testStruct{}, model.InScope(scope1))
		ret, err := exe.Execute()

		assert.Nil(t, err)
		assert.Equal(t, testStruct{
			a: 123,
			b: "abc",
		}, ret)
	})
}

func Test_container_ValueOf(t *testing.T) {
	t.Run("only one value", func(t *testing.T) {
		m, scope1, scope2 := buildModuleForContainerTest()
		c, _ := newContainer(m, nil)
		c1, _ := c.EnterScope(scope1)
		c2, _ := c1.EnterScope(scope2)

		exe := c2.ValueOf(testStruct3{})
		ret, err := exe.Execute()

		assert.Nil(t, err)
		ts3 := ret.(testStruct3)

		assert.Equal(t, 123, ts3.a)
		assert.Equal(t, "abc", ts3.b)
	})

	t.Run("one of value", func(t *testing.T) {
		m := model.NewModule(
			model.Value(123),
			model.Value(456),
		)
		c, _ := newContainer(m, nil)
		exe := c.ValueOf(model.TypeOf(0))
		ret, err := exe.Execute()
		assert.Nil(t, err)
		assert.True(t, ret == 123 || ret == 456)
	})

	t.Run("collect value", func(t *testing.T) {
		m := model.NewModule(
			model.Value(123),
			model.Value(456),
		)
		c, _ := newContainer(m, nil)
		exe := c.ValueOf(model.TypeOf([]int(nil)), model.AsCollector(true))
		ret, err := exe.Execute()
		assert.Nil(t, err)
		arr := ret.([]int)
		assert.Equal(t, 2, len(arr))
		assert.Contains(t, arr, 123)
		assert.Contains(t, arr, 456)
	})
}

func Test_container_Scope(t *testing.T) {
	m, scope1, _ := buildModuleForContainerTest()
	c, _ := newContainer(m, nil)
	assert.Equal(t, model.GlobalScope, c.Scope())
	c1, _ := c.EnterScope(scope1)
	assert.Equal(t, scope1, c1.Scope())
}

func Test_container_EnterScope(t *testing.T) {
	t.Run("can enter", func(t *testing.T) {
		m, scope1, scope2 := buildModuleForContainerTest()
		c, _ := newContainer(m, nil)
		c1, err1 := c.EnterScope(scope1)
		assert.Nil(t, err1)
		assert.Equal(t, scope1, c1.Scope())

		c2, err2 := c1.EnterScope(scope2)
		assert.Nil(t, err2)
		assert.Equal(t, scope2, c2.Scope())
	})

	t.Run("can not enter", func(t *testing.T) {
		scope3 := model.NewScope("scope3")

		m, scope1, _ := buildModuleForContainerTest()
		c, _ := newContainer(m, nil)
		c1, _ := c.EnterScope(scope1)

		_, err := c1.EnterScope(scope3)
		assert.NotNil(t, err)
	})

	t.Run("different scope value", func(t *testing.T) {
		scope1 := model.NewScope("scope1")
		counter := 0
		m1 := model.NewModule(model.Func(func() int {
			counter += 1
			return counter
		}, model.InScope(scope1)))

		c, _ := newContainer(m1, nil)
		c1, _ := c.EnterScope(scope1)

		exe := c1.ValueOf(model.TypeOf(0))

		ret, _ := exe.Execute()
		assert.Equal(t, 1, ret.(int))
		ret2, _ := exe.Execute()
		assert.Equal(t, 1, ret2.(int))

		c2, _ := c1.LeaveScope().EnterScope(scope1)
		exe2 := c2.ValueOf(model.TypeOf(0))
		ret3, _ := exe2.Execute()
		assert.Equal(t, 2, ret3.(int))
		ret4, _ := exe2.Execute()
		assert.Equal(t, 2, ret4.(int))
	})
}

func Test_container_LeaveScope(t *testing.T) {
	t.Run("leave", func(t *testing.T) {
		m, scope1, scope2 := buildModuleForContainerTest()
		c, _ := newContainer(m, nil)
		c1, _ := c.EnterScope(scope1)
		c2, _ := c1.EnterScope(scope2)

		c3 := c2.LeaveScope()
		assert.Equal(t, scope1, c3.Scope())
		c4 := c3.LeaveScope()
		assert.Equal(t, model.GlobalScope, c4.Scope())
		c5 := c4.LeaveScope()
		assert.Equal(t, model.GlobalScope, c5.Scope())
	})
}
