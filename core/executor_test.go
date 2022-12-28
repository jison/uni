package core

import (
	"testing"

	"github.com/jison/uni/core/model"
	"github.com/stretchr/testify/assert"
)

func Test_executor(t *testing.T) {
	module, scope1, scope2, scope3 := buildTestModule()
	rep := model.NewRepository(module.AllComponents())
	g := newDependenceGraph(rep)

	t.Run("Func", func(t *testing.T) {
		consumer := model.FuncConsumer(
			func(a []int) int {
				s := 0
				for _, i := range a {
					s += i
				}
				return s
			},
			model.InScope(scope2),
			model.Param(0, model.AsCollector(true)),
		).Consumer()

		ss0 := newScopeStorage()
		ss1, _ := ss0.Enter(scope1)
		ss2, _ := ss1.Enter(scope2)

		exe := newExecutor(g, ss2, consumer)
		val, err := exe.Execute()
		assert.Nil(t, err)
		assert.Equal(t, 579, val.([]interface{})[0])
	})

	t.Run("Value", func(t *testing.T) {
		consumer := model.ValueConsumer(
			model.TypeOf(0),
			model.ByName("name3"),
			model.InScope(scope3),
		).Consumer()

		ss0 := newScopeStorage()
		ss3, _ := ss0.Enter(scope3)

		exe := newExecutor(g, ss3, consumer)
		val, err := exe.Execute()
		assert.Nil(t, err)
		assert.Equal(t, 789, val)
	})

	t.Run("Struct", func(t *testing.T) {
		consumer := model.StructConsumer(
			model.TypeOf(testStruct{}),
			model.InScope(scope3),
			model.Field("a", model.ByName("name3")),
			model.Field("b", model.ByName("name4")),
		).Consumer()

		ss0 := newScopeStorage()
		ss3, _ := ss0.Enter(scope3)

		exe := newExecutor(g, ss3, consumer)
		val, err := exe.Execute()
		assert.Nil(t, err)
		assert.Equal(t, testStruct{
			a: 789,
			b: "abc",
		}, val)
	})

	t.Run("consumer with error", func(t *testing.T) {
		consumer := model.StructConsumer(
			model.TypeOf(12), // not struct type
			model.InScope(scope3),
		).Consumer()

		ss0 := newScopeStorage()
		ss3, _ := ss0.Enter(scope3)

		exe := newExecutor(g, ss3, consumer)
		_, err := exe.Execute()
		assert.NotNil(t, err)
	})

	t.Run("not unique dependencies", func(t *testing.T) {
		consumer := model.StructConsumer(
			model.TypeOf(testStruct{}),
			model.InScope(scope3),
		).Consumer()

		ss0 := newScopeStorage()
		ss3, _ := ss0.Enter(scope3)

		exe := newExecutor(g, ss3, consumer)
		val, err := exe.Execute()
		assert.Nil(t, err)
		s := val.(testStruct)
		assert.Equal(t, "abc", s.b)
		assert.Contains(t, []int{123, 789}, s.a)
	})

	t.Run("missing dependency", func(t *testing.T) {
		consumer := model.StructConsumer(
			model.TypeOf(testStruct2{}),
		).Consumer()

		ss0 := newScopeStorage()

		exe := newExecutor(g, ss0, consumer)
		_, err := exe.Execute()
		//t.Logf("%v\n", err)
		assert.NotNil(t, err)
	})

	t.Run("cycle", func(t *testing.T) {
		consumer := model.ValueConsumer(
			model.TypeOf(&testStruct3{}),
			model.InScope(scope2),
		).Consumer()

		ss0 := newScopeStorage()
		ss1, _ := ss0.Enter(scope1)
		ss2, _ := ss1.Enter(scope2)

		exe := newExecutor(g, ss2, consumer)
		_, err := exe.Execute()
		//t.Logf("%v\n", err)
		assert.NotNil(t, err)
	})
}
