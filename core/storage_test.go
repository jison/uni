package core

import (
	"math/rand"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/jison/uni/core/model"
	"github.com/jison/uni/core/valuer"
	"github.com/jison/uni/internal/errors"
	"github.com/stretchr/testify/assert"
)

func Test_scopeStorage_Get(t *testing.T) {
	scope1 := model.NewScope("scope1")
	scope2 := model.NewScope("scope2", scope1)
	scope3 := model.NewScope("scope3")

	ss := newScopeStorage()
	ss1, _ := ss.Enter(scope1)
	ss2, _ := ss1.Enter(scope2)

	node1 := valuer.Identity()
	node2 := valuer.Identity()
	node3 := valuer.Identity()

	_, _ = ss2.GetOrElse(node1, scope1, func(storage ScopeBaseStorage) valuer.Value {
		return valuer.SingleValue(reflect.ValueOf(1))
	}).AsSingle()
	_, _ = ss2.GetOrElse(node2, scope2, func(storage ScopeBaseStorage) valuer.Value {
		return valuer.SingleValue(reflect.ValueOf(2))
	}).AsSingle()

	t.Run("in current scope", func(t *testing.T) {
		t.Run("have value", func(t *testing.T) {
			val, ok := ss2.Get(node2, scope2)
			assert.True(t, ok)
			v, ok2 := val.AsSingle()
			assert.True(t, ok2)
			assert.Equal(t, reflect.ValueOf(2), v)
		})

		t.Run("have not value", func(t *testing.T) {
			_, ok := ss2.Get(node3, scope2)
			assert.False(t, ok)
		})

	})

	t.Run("in parent scope", func(t *testing.T) {
		val, ok := ss2.Get(node1, scope1)
		assert.True(t, ok)
		v, ok2 := val.AsSingle()
		assert.True(t, ok2)
		assert.Equal(t, reflect.ValueOf(1), v)
	})

	t.Run("have not value", func(t *testing.T) {
		_, ok := ss2.Get(node3, scope3)
		assert.False(t, ok)
	})
}

func Test_scopeStorage_GetOrElse(t *testing.T) {
	scope1 := model.NewScope("scope1")
	scope2 := model.NewScope("scope2", scope1)
	scope3 := model.NewScope("scope3")

	t.Run("node is nil", func(t *testing.T) {
		ss := newScopeStorage()
		ss1, _ := ss.Enter(scope1)
		ss2, _ := ss1.Enter(scope2)

		val := ss2.GetOrElse(nil, scope2, func(_ ScopeBaseStorage) valuer.Value {
			return valuer.SingleValue(reflect.ValueOf(123))
		})
		err, ok := val.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("valSupplier is nil", func(t *testing.T) {
		ss := newScopeStorage()
		ss1, _ := ss.Enter(scope1)
		ss2, _ := ss1.Enter(scope2)

		node := valuer.Identity()
		val := ss2.GetOrElse(node, scope2, nil)
		err, ok := val.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("scope is nil", func(t *testing.T) {
		ss := newScopeStorage()
		ss1, _ := ss.Enter(scope1)
		ss2, _ := ss1.Enter(scope2)

		node := valuer.Identity()
		runCount := 0
		val := ss2.GetOrElse(node, nil, func(_ ScopeBaseStorage) valuer.Value {
			runCount += 1
			return valuer.SingleValue(reflect.ValueOf(123))
		})
		rVal, ok := val.AsSingle()
		assert.True(t, ok)
		assert.Equal(t, reflect.ValueOf(123), rVal)
		assert.Equal(t, 1, runCount)

		val2 := ss2.GetOrElse(node, nil, func(_ ScopeBaseStorage) valuer.Value {
			runCount += 1
			return valuer.SingleValue(reflect.ValueOf(456))
		})
		rVal2, ok2 := val2.AsSingle()
		assert.True(t, ok2)
		assert.Equal(t, reflect.ValueOf(456), rVal2)
		assert.Equal(t, 2, runCount)
	})

	t.Run("scope is nil and valSupplier is nil", func(t *testing.T) {
		ss := newScopeStorage()
		ss1, _ := ss.Enter(scope1)
		ss2, _ := ss1.Enter(scope2)

		node := valuer.Identity()
		val := ss2.GetOrElse(node, nil, nil)
		err, ok := val.AsError()
		assert.True(t, ok)
		assert.NotNil(t, err)
	})

	t.Run("scope have not entered", func(t *testing.T) {
		t.Run("have not value", func(t *testing.T) {
			ss := newScopeStorage()
			ss1, _ := ss.Enter(scope1)
			ss2, _ := ss1.Enter(scope2)

			node := valuer.Identity()
			val := ss2.GetOrElse(node, scope3, func(_ ScopeBaseStorage) valuer.Value {
				return valuer.SingleValue(reflect.ValueOf(123))
			})
			err, ok := val.AsError()
			assert.True(t, ok)
			assert.NotNil(t, err)
		})
	})

	t.Run("scope have entered before", func(t *testing.T) {
		t.Run("have not value", func(t *testing.T) {
			ss := newScopeStorage()
			ss1, _ := ss.Enter(scope1)
			ss2, _ := ss1.Enter(scope2)

			node := valuer.Identity()
			runCount := 0
			val := ss2.GetOrElse(node, scope1, func(_ ScopeBaseStorage) valuer.Value {
				runCount += 1
				return valuer.SingleValue(reflect.ValueOf(123))
			})
			rVal, ok := val.AsSingle()
			assert.Equal(t, 1, runCount)

			assert.True(t, ok)
			assert.Equal(t, reflect.ValueOf(123), rVal)
		})

		t.Run("return error value", func(t *testing.T) {
			ss := newScopeStorage()
			ss1, _ := ss.Enter(scope1)
			ss2, _ := ss1.Enter(scope2)

			node := valuer.Identity()
			runCount := 0
			val := ss2.GetOrElse(node, scope1, func(_ ScopeBaseStorage) valuer.Value {
				runCount += 1
				return valuer.ErrorValue(errors.Newf("this is an error"))
			})
			err, ok := val.AsError()
			assert.True(t, ok)
			assert.NotNil(t, err)
			assert.Equal(t, 1, runCount)

			val2 := ss2.GetOrElse(node, scope1, func(_ ScopeBaseStorage) valuer.Value {
				runCount += 1
				return valuer.SingleValue(reflect.ValueOf(123))
			})
			rVal, ok := val2.AsSingle()
			assert.Equal(t, 2, runCount)
			assert.True(t, ok)
			assert.Equal(t, reflect.ValueOf(123), rVal)
		})

		t.Run("have value", func(t *testing.T) {
			ss := newScopeStorage()
			ss1, _ := ss.Enter(scope1)
			ss2, _ := ss1.Enter(scope2)

			node := valuer.Identity()
			runCount := 0
			val0 := ss2.GetOrElse(node, scope1, func(_ ScopeBaseStorage) valuer.Value {
				runCount += 1
				return valuer.SingleValue(reflect.ValueOf(123))
			})
			_, _ = val0.AsSingle()
			assert.Equal(t, 1, runCount)

			t.Run("get from the same scope", func(t *testing.T) {
				val1 := ss2.GetOrElse(node, scope1, func(_ ScopeBaseStorage) valuer.Value {
					runCount += 1
					return valuer.SingleValue(reflect.ValueOf(456))
				})
				rVal, ok := val1.AsSingle()
				assert.Equal(t, 1, runCount)
				assert.True(t, ok)
				assert.Equal(t, reflect.ValueOf(123), rVal)
			})

			t.Run("get from different scope", func(t *testing.T) {
				val1 := ss2.GetOrElse(node, scope2, func(_ ScopeBaseStorage) valuer.Value {
					runCount += 1
					return valuer.SingleValue(reflect.ValueOf(456))
				})
				rVal, ok := val1.AsSingle()
				assert.Equal(t, 2, runCount)
				assert.True(t, ok)
				assert.Equal(t, reflect.ValueOf(456), rVal)
			})
		})

		t.Run("get from current scope", func(t *testing.T) {
			ss := newScopeStorage()
			ss1, _ := ss.Enter(scope1)
			ss2, _ := ss1.Enter(scope2)

			node := valuer.Identity()
			runCount := 0
			val := ss2.GetOrElse(node, scope2, func(_ ScopeBaseStorage) valuer.Value {
				runCount += 1
				return valuer.SingleValue(reflect.ValueOf(123))
			})
			rVal, ok := val.AsSingle()
			assert.Equal(t, 1, runCount)
			assert.True(t, ok)
			assert.Equal(t, reflect.ValueOf(123), rVal)
		})
	})

	t.Run("recursive", func(t *testing.T) {
		ss0 := newScopeStorage()
		ss1, _ := ss0.Enter(scope1)
		ss2, _ := ss1.Enter(scope2)

		t.Run("scope is superior", func(t *testing.T) {
			node1 := valuer.Identity()
			node2 := valuer.Identity()
			node3 := valuer.Identity()
			runCount := 0
			val := ss2.GetOrElse(node1, scope2, func(ss ScopeBaseStorage) valuer.Value {
				runCount += 1
				return ss.GetOrElse(node2, scope1, func(ss ScopeBaseStorage) valuer.Value {
					runCount += 1
					return ss.GetOrElse(node3, model.GlobalScope, func(ss ScopeBaseStorage) valuer.Value {
						runCount += 1
						return valuer.SingleValue(reflect.ValueOf(123))
					})
				})
			})
			rVal, ok := val.AsSingle()
			assert.Equal(t, 3, runCount)
			assert.True(t, ok)
			assert.Equal(t, reflect.ValueOf(123), rVal)
		})

		t.Run("scope is inferior", func(t *testing.T) {
			node1 := valuer.Identity()
			node2 := valuer.Identity()
			node3 := valuer.Identity()
			runCount := 0
			val := ss2.GetOrElse(node1, scope1, func(ss ScopeBaseStorage) valuer.Value {
				runCount += 1
				return ss.GetOrElse(node2, scope2, func(ss ScopeBaseStorage) valuer.Value {
					runCount += 1
					return ss.GetOrElse(node3, model.GlobalScope, func(ss ScopeBaseStorage) valuer.Value {
						runCount += 1
						return valuer.SingleValue(reflect.ValueOf(123))
					})
				})
			})
			err, ok := val.AsError()
			assert.True(t, ok)
			assert.NotNil(t, err)
		})
	})

	t.Run("concurrent", func(t *testing.T) {
		ss0 := newScopeStorage()
		ss1, _ := ss0.Enter(scope1)
		ss2, _ := ss1.Enter(scope2)

		t.Run("get one node concurrently", func(t *testing.T) {
			var counter int32
			n := 20
			node := valuer.Identity()
			finalVal := -1
			wg := sync.WaitGroup{}
			for i := 0; i < n; i++ {
				wg.Add(1)
				go func(v int) {
					val := ss2.GetOrElse(node, scope2, func(_ ScopeBaseStorage) valuer.Value {
						atomic.AddInt32(&counter, 1)
						finalVal = v
						return valuer.SingleValue(reflect.ValueOf(v))
					})
					_, _ = val.AsSingle()
					wg.Done()
				}(i)
			}
			wg.Wait()
			assert.Equal(t, 1, int(atomic.LoadInt32(&counter)))
			val := ss2.GetOrElse(node, scope2, func(_ ScopeBaseStorage) valuer.Value {
				return valuer.SingleValue(reflect.ValueOf(n + 1))
			})
			rVal, ok := val.AsSingle()
			assert.True(t, ok)
			assert.Equal(t, reflect.ValueOf(finalVal), rVal)
		})

		t.Run("mass concurrently", func(t *testing.T) {
			type nodeInfo struct {
				node    Node
				scope   model.Scope
				val     int
				counter int32
			}

			randomScope := func() model.Scope {
				scopes := []model.Scope{model.GlobalScope, scope1, scope2}
				return scopes[rand.Intn(3)]
			}

			var infos []*nodeInfo
			nodeCount := 100
			for i := 0; i < nodeCount; i++ {
				infos = append(infos, &nodeInfo{
					node:    valuer.Identity(),
					scope:   randomScope(),
					val:     -1,
					counter: 0,
				})
			}

			runCount := 1000
			wg := sync.WaitGroup{}
			for i := 0; i < runCount; i++ {
				wg.Add(1)
				go func(v int) {
					info := infos[rand.Intn(nodeCount)]
					val := ss2.GetOrElse(info.node, info.scope, func(_ ScopeBaseStorage) valuer.Value {
						atomic.AddInt32(&info.counter, 1)
						info.val = v
						return valuer.SingleValue(reflect.ValueOf(v))
					})
					_, _ = val.AsSingle()
					wg.Done()
				}(i)
			}
			wg.Wait()

			for i := 0; i < nodeCount; i++ {
				info := infos[i]
				assert.Equal(t, 1, int(atomic.LoadInt32(&info.counter)))
				val := ss2.GetOrElse(info.node, info.scope, nil)
				rVal, ok := val.AsSingle()
				assert.True(t, ok)
				assert.Equal(t, info.val, rVal.Interface())
			}
		})
	})
}

func Test_scopeStorage_Enter(t *testing.T) {
	scope1 := model.NewScope("scope1")
	scope2 := model.NewScope("scope2", scope1)
	scope3 := model.NewScope("scope3")

	t.Run("enter nil scope", func(t *testing.T) {
		ss := newScopeStorage()
		ss2, err := ss.Enter(nil)
		assert.NotNil(t, err)
		assert.Nil(t, ss2)
	})

	t.Run("can enter", func(t *testing.T) {
		ss := newScopeStorage()
		ss1, err1 := ss.Enter(scope1)
		assert.Nil(t, err1)
		assert.Equal(t, scope1, ss1.Scope())

		ss2, err2 := ss1.Enter(scope2)
		assert.Nil(t, err2)
		assert.Equal(t, scope2, ss2.Scope())
	})

	t.Run("can not enter", func(t *testing.T) {
		ss := newScopeStorage()
		_, err1 := ss.Enter(scope2)
		assert.NotNil(t, err1)

		ss1, err2 := ss.Enter(scope3)
		assert.Nil(t, err2)
		_, err3 := ss1.Enter(scope1)
		assert.NotNil(t, err3)
	})
}

func Test_scopeStorage_Leave(t *testing.T) {
	scope1 := model.NewScope("scope1")
	scope2 := model.NewScope("scope2", scope1)

	t.Run("leave from global", func(t *testing.T) {
		ss := newScopeStorage()
		ss2 := ss.Leave()
		assert.Nil(t, ss2)
	})

	t.Run("leave from scope", func(t *testing.T) {
		ss := newScopeStorage()
		ss1, _ := ss.Enter(scope1)
		ss2, _ := ss1.Enter(scope2)

		ss3 := ss2.Leave()
		assert.Equal(t, scope1, ss3.Scope())
		ss4 := ss3.Leave()
		assert.Equal(t, model.GlobalScope, ss4.Scope())
	})
}

func Test_scopeStorage_Scope(t *testing.T) {
	scope1 := model.NewScope("scope1")

	t.Run("global", func(t *testing.T) {
		ss := newScopeStorage()
		assert.Equal(t, model.GlobalScope, ss.Scope())
	})

	t.Run("scope", func(t *testing.T) {
		ss := newScopeStorage()
		ss1, _ := ss.Enter(scope1)
		assert.Equal(t, scope1, ss1.Scope())
	})
}
