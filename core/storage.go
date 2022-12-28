package core

import (
	"sync"

	"github.com/jison/uni/core/valuer"

	"github.com/jison/uni/core/model"
	"github.com/jison/uni/internal/errors"
)

type ScopeBaseStorage interface {
	GetOrElse(node Node, scope model.Scope, getter func(ScopeBaseStorage) valuer.Value) valuer.Value
}

func newScopeStorage() *scopeStorage {
	return &scopeStorage{
		parent:      nil,
		scope:       model.GlobalScope,
		valueByNode: &sync.Map{},
		mutexByNode: &sync.Map{},
	}
}

type scopeStorage struct {
	parent      *scopeStorage
	scope       model.Scope
	valueByNode *sync.Map // map[Node]valuer.Value
	mutexByNode *sync.Map // map[Node]*sync.Mutex
}

func (s *scopeStorage) Scope() model.Scope {
	return s.scope
}

func (s *scopeStorage) Get(node Node, scope model.Scope) (valuer.Value, bool) {
	if s.scope == scope {
		value, ok := s.valueByNode.Load(node)
		if ok {
			return value.(valuer.Value), true
		} else {
			return nil, false
		}
	} else if s.parent != nil {
		return s.parent.Get(node, scope)
	} else {
		return nil, false
	}
}

func (s *scopeStorage) GetOrElse(node Node, scope model.Scope,
	valSupplier func(ScopeBaseStorage) valuer.Value) valuer.Value {
	if node == nil {
		return valuer.ErrorValue(errors.Newf("node is nil"))
	}

	if scope == nil {
		if valSupplier == nil {
			return valuer.ErrorValue(errors.Newf("valSupplier is nil"))
		}
		return valSupplier(s)
	}

	if s.scope == scope {
		value, ok := s.valueByNode.Load(node)
		if ok {
			return value.(valuer.Value)
		} else if valSupplier == nil {
			return valuer.ErrorValue(errors.Newf("valSupplier is nil"))
		} else {
			return valuer.LazyValue(func() valuer.Value {
				nodeMutex := s.getMutexByNode(node)
				return s.update(nodeMutex, node, valSupplier)
			})
		}
	} else if s.parent != nil {
		return s.parent.GetOrElse(node, scope, valSupplier)
	} else {
		return valuer.ErrorValue(errors.Newf("this scope `%v` is not entered in the context", scope))
	}
}

func (s *scopeStorage) update(mutex *sync.Mutex, node Node,
	valSupplier func(ScopeBaseStorage) valuer.Value) valuer.Value {
	mutex.Lock()
	defer mutex.Unlock()

	nodeVal, ok := s.valueByNode.Load(node)
	if ok {
		return nodeVal.(valuer.Value)
	}

	val := valSupplier(s)
	if _, isErr := val.AsError(); !isErr {
		s.valueByNode.Store(node, val)
	}

	return val
}

func (s *scopeStorage) getMutexByNode(node Node) *sync.Mutex {
	if mutex, ok := s.mutexByNode.Load(node); ok {
		return mutex.(*sync.Mutex)
	}

	mutex, _ := s.mutexByNode.LoadOrStore(node, &sync.Mutex{})
	return mutex.(*sync.Mutex)
}

func (s *scopeStorage) Enter(scope model.Scope) (*scopeStorage, error) {
	if scope == nil || !scope.CanEnterDirectlyFrom(s.scope) {
		return nil, errors.Newf("%+v can not enter from %+v directly", scope, s.scope)
	}

	newS := &scopeStorage{
		parent:      s,
		scope:       scope,
		valueByNode: &sync.Map{},
		mutexByNode: &sync.Map{},
	}
	return newS, nil
}

func (s *scopeStorage) Leave() *scopeStorage {
	return s.parent
}
