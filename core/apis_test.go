package core

import (
	"context"
	"github.com/jison/uni/core/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

//func TestLoad(t *testing.T) {
//	t.Run("load", func(t *testing.T) {
//		m, scope1, _ := buildModuleForContainerTest()
//		c, err := newContainer(m, nil)
//		assert.Nil(t, err)
//
//		c1, _ := c.EnterScope(scope1)
//		err = Load(c1, model.NewCriteria(testStruct{}), model.NewCriteria(&testStruct{}))
//		assert.Nil(t, err)
//
//		c2 := c1.(*container)
//
//		runCount := 0
//		m.AllComponents().Filter(func(com model.Component) bool {
//			return com.Name() == "name3" || com.Name() == "name4"
//		}).Iterate(func(com model.Component) bool {
//			runCount += 1
//			_, ok := c2.storage.Get(com.Valuer(), com.Provider().Scope())
//			assert.True(t, ok)
//			return true
//		})
//
//		assert.Equal(t, 2, runCount)
//	})
//}
//
//func TestLoadAll(t *testing.T) {
//	t.Run("load all", func(t *testing.T) {
//		m, scope1, _ := buildModuleForContainerTest()
//		c, err := newContainer(m, nil)
//		assert.Nil(t, err)
//
//		c1, _ := c.EnterScope(scope1)
//		err = LoadAll(c1)
//		assert.Nil(t, err)
//
//		c2 := c1.(*container)
//
//		runCount := 0
//		m.AllComponents().Filter(func(com model.Component) bool {
//			return com.Name() == "name3" || com.Name() == "name4"
//		}).Iterate(func(com model.Component) bool {
//			runCount += 1
//			_, ok := c2.storage.Get(com.Valuer(), com.Provider().Scope())
//			assert.True(t, ok)
//			return true
//		})
//
//		assert.Equal(t, 2, runCount)
//	})
//}

func TestFuncOf(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		m, scope1, scope2 := buildModuleForContainerTest()
		c, _ := newContainer(m, nil)
		c1, _ := c.EnterScope(scope1)
		c2, _ := c1.EnterScope(scope2)

		runCount := 0
		ret, err := FuncOf(c2, func(a testStruct3, b *testStruct3) (int, error) {
			runCount += 1

			assert.Equal(t, a.a, b.a)
			assert.Equal(t, a.b, b.b)
			assert.Equal(t, a.ts2, b.ts2)
			return a.a, nil
		})

		assert.Equal(t, 1, runCount)
		assert.Equal(t, 123, ret.([]interface{})[0])
		assert.Nil(t, err)
	})

	t.Run("container is nil", func(t *testing.T) {
		var c Container
		_, err := FuncOf(c, func(a testStruct3, b *testStruct3) (int, error) {
			return a.a, nil
		})
		assert.NotNil(t, err)
	})
}

func TestStructOf(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		m, scope1, scope2 := buildModuleForContainerTest()
		c, _ := newContainer(m, nil)
		c1, _ := c.EnterScope(scope1)
		c2, _ := c1.EnterScope(scope2)

		ret, err := StructOf(c2, testStruct{}, model.InScope(scope1))
		assert.Nil(t, err)
		assert.Equal(t, testStruct{
			a: 123,
			b: "abc",
		}, ret)
	})

	t.Run("container is nil", func(t *testing.T) {
		var c Container
		_, err := StructOf(c, testStruct{})
		assert.NotNil(t, err)
	})
}

func TestValueOf(t *testing.T) {
	t.Run("only one value", func(t *testing.T) {
		m, scope1, scope2 := buildModuleForContainerTest()
		c, _ := newContainer(m, nil)
		c1, _ := c.EnterScope(scope1)
		c2, _ := c1.EnterScope(scope2)

		ret, err := ValueOf(c2, testStruct3{})
		assert.Nil(t, err)
		ts3 := ret.(testStruct3)

		assert.Equal(t, 123, ts3.a)
		assert.Equal(t, "abc", ts3.b)
	})

	t.Run("container is nil", func(t *testing.T) {
		var c Container
		_, err := ValueOf(c, 0)
		assert.NotNil(t, err)
	})
}

func TestEnterScope(t *testing.T) {
	t.Run("can enter", func(t *testing.T) {
		m, scope1, scope2 := buildModuleForContainerTest()
		c, _ := NewContainer(m)
		c1, err1 := EnterScope(c, scope1)
		assert.Nil(t, err1)
		assert.Equal(t, scope1, c1.Scope())

		c2, err2 := EnterScope(c1, scope2)
		assert.Nil(t, err2)
		assert.Equal(t, scope2, c2.Scope())
	})

	t.Run("container is nil", func(t *testing.T) {
		scope1 := model.NewScope("scope1")
		var c Container
		_, err := EnterScope(c, scope1)
		assert.NotNil(t, err)
	})
}

func TestLeaveScope(t *testing.T) {
	t.Run("leave", func(t *testing.T) {
		m, scope1, scope2 := buildModuleForContainerTest()
		c, _ := NewContainer(m)
		c1, _ := EnterScope(c, scope1)
		c2, _ := EnterScope(c1, scope2)

		c3 := LeaveScope(c2)
		assert.Equal(t, scope1, c3.Scope())
		c4 := LeaveScope(c3)
		assert.Equal(t, model.GlobalScope, c4.Scope())
		c5 := LeaveScope(c4)
		assert.Equal(t, model.GlobalScope, c5.Scope())
	})

	t.Run("container is nil", func(t *testing.T) {
		var c Container
		c2 := LeaveScope(c)
		assert.Nil(t, c2)
	})
}

func TestWithContainer(t *testing.T) {
	m, _, _ := buildModuleForContainerTest()
	c, _ := NewContainer(m)

	c1 := ContainerOfCtx(context.TODO())
	assert.Nil(t, c1)

	ctx := WithContainerCtx(context.TODO(), c)
	c2 := ContainerOfCtx(ctx)
	assert.NotNil(t, c2)
	assert.Same(t, c, c2)
}

func TestFuncOfCtx(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		m, scope1, scope2 := buildModuleForContainerTest()
		c, _ := NewContainer(m)
		ctx := WithContainerCtx(context.TODO(), c)
		ctx1, _ := EnterScopeCtx(ctx, scope1)
		ctx2, _ := EnterScopeCtx(ctx1, scope2)

		runCount := 0

		ret, err := FuncOfCtx(ctx2, func(a testStruct3, b *testStruct3) (int, error) {
			runCount += 1

			assert.Equal(t, a.a, b.a)
			assert.Equal(t, a.b, b.b)
			assert.Equal(t, a.ts2, b.ts2)
			return a.a, nil
		})

		assert.Equal(t, 1, runCount)
		assert.Equal(t, 123, ret.([]interface{})[0])
		assert.Nil(t, err)
	})
}

func TestStructOfCtx(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		m, scope1, scope2 := buildModuleForContainerTest()
		c, _ := NewContainer(m)
		ctx := WithContainerCtx(context.TODO(), c)
		ctx1, _ := EnterScopeCtx(ctx, scope1)
		ctx2, _ := EnterScopeCtx(ctx1, scope2)

		ret, err := StructOfCtx(ctx2, testStruct{}, model.InScope(scope1))

		assert.Nil(t, err)
		assert.Equal(t, testStruct{
			a: 123,
			b: "abc",
		}, ret)
	})
}

func TestValueOfCtx(t *testing.T) {
	t.Run("only one value", func(t *testing.T) {
		m, scope1, scope2 := buildModuleForContainerTest()
		c, _ := NewContainer(m)
		ctx := WithContainerCtx(context.TODO(), c)
		ctx1, _ := EnterScopeCtx(ctx, scope1)
		ctx2, _ := EnterScopeCtx(ctx1, scope2)

		ret, err := ValueOfCtx(ctx2, testStruct3{})

		assert.Nil(t, err)
		ts3 := ret.(testStruct3)

		assert.Equal(t, 123, ts3.a)
		assert.Equal(t, "abc", ts3.b)
	})
}

func TestEnterScopeCtx(t *testing.T) {
	t.Run("can enter", func(t *testing.T) {
		m, scope1, scope2 := buildModuleForContainerTest()
		c, _ := NewContainer(m)
		ctx := WithContainerCtx(context.TODO(), c)

		ctx1, err1 := EnterScopeCtx(ctx, scope1)
		assert.Nil(t, err1)
		c1 := ContainerOfCtx(ctx1)
		assert.Equal(t, scope1, c1.Scope())

		ctx2, err2 := EnterScopeCtx(ctx1, scope2)
		assert.Nil(t, err2)
		c2 := ContainerOfCtx(ctx2)
		assert.Equal(t, scope2, c2.Scope())
	})

	t.Run("container is nil", func(t *testing.T) {
		scope1 := model.NewScope("scope1")
		_, err1 := EnterScopeCtx(context.TODO(), scope1)
		assert.NotNil(t, err1)
	})
}

func TestLeaveScopeCtx(t *testing.T) {
	t.Run("leave", func(t *testing.T) {
		m, scope1, scope2 := buildModuleForContainerTest()
		c, _ := NewContainer(m)
		ctx := WithContainerCtx(context.TODO(), c)
		ctx1, _ := EnterScopeCtx(ctx, scope1)
		ctx2, _ := EnterScopeCtx(ctx1, scope2)

		ctx3 := LeaveScopeCtx(ctx2)
		c3 := ContainerOfCtx(ctx3)
		assert.Equal(t, scope1, c3.Scope())
		ctx4 := LeaveScopeCtx(ctx3)
		c4 := ContainerOfCtx(ctx4)
		assert.Equal(t, model.GlobalScope, c4.Scope())
		ctx5 := LeaveScopeCtx(ctx4)
		c5 := ContainerOfCtx(ctx5)
		assert.Equal(t, model.GlobalScope, c5.Scope())
	})

	t.Run("container is nil", func(t *testing.T) {
		ctx := context.TODO()
		ctx2 := LeaveScopeCtx(ctx)
		assert.Same(t, ctx, ctx2)
	})
}
