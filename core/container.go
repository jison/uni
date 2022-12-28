package core

import (
	"github.com/jison/uni/core/model"
	"github.com/jison/uni/internal/errors"
)

type Container interface {
	Load(criteriaList ...model.CriteriaBuilder) error
	LoadAll() error
	FuncOf(function interface{}, opts ...model.FuncConsumerOption) Executor
	StructOf(t model.TypeVal, opts ...model.StructConsumerOption) Executor
	ValueOf(t model.TypeVal, opts ...model.ValueConsumerOption) Executor

	ExecutorOf(cb model.ConsumerBuilder) Executor

	Scope() model.Scope
	EnterScope(model.Scope) (Container, error)
	LeaveScope() Container
}

type ContainerOptions struct {
	ignoreMissing   bool
	ignoreUncertain bool
	ignoreCycle     bool
}

type ContainerOption func(*ContainerOptions)

func IgnoreMissing() ContainerOption {
	return func(opts *ContainerOptions) {
		opts.ignoreMissing = true
	}
}

func IgnoreUncertain() ContainerOption {
	return func(opts *ContainerOptions) {
		opts.ignoreUncertain = true
	}
}

func IgnoreCycle() ContainerOption {
	return func(opts *ContainerOptions) {
		opts.ignoreCycle = true
	}
}

func NewContainer(m model.Module, opts ...ContainerOption) (Container, error) {
	containerOpts := &ContainerOptions{}
	for _, opt := range opts {
		opt(containerOpts)
	}

	return newContainer(m, containerOpts)
}

func newContainer(m model.Module, opts *ContainerOptions) (*container, error) {
	if m == nil {
		return nil, errors.Newf("module is nil")
	}

	if err := m.Validate(); err != nil {
		return nil, err
	}

	rep := model.NewRepository(m.AllComponents())
	g := newDependenceGraph(rep)

	errs := errors.Empty()

	if err := g.MissingError(); err != nil && (opts == nil || !opts.ignoreMissing) {
		errs = errs.AddErrors(err)
	}

	if err := g.UncertainError(); err != nil && (opts == nil || !opts.ignoreUncertain) {
		errs = errs.AddErrors(err)
	}

	if err := g.CycleError(); err != nil && (opts == nil || !opts.ignoreCycle) {
		errs = errs.AddErrors(err)
	}

	if errs.HasError() {
		return nil, errs
	}

	return &container{
		graph:   g,
		storage: newScopeStorage(),
	}, nil
}

type container struct {
	graph   DependenceGraph
	storage *scopeStorage
}

func (c *container) Load(criteriaList ...model.CriteriaBuilder) error {
	if c == nil {
		return errors.Newf("container is nil")
	}

	cb := model.LoadCriteriaConsumer(criteriaList...).
		SetScope(c.Scope()).
		UpdateCallLocation(nil)

	e := c.ExecutorOf(cb)
	_, err := e.Execute()
	return err
}

func (c *container) LoadAll() error {
	if c == nil {
		return errors.Newf("container is nil")
	}

	cb := model.LoadAllConsumer(c.Scope()).UpdateCallLocation(nil)
	e := c.ExecutorOf(cb)
	_, err := e.Execute()
	return err
}

func (c *container) FuncOf(function interface{}, opts ...model.FuncConsumerOption) Executor {
	if c == nil {
		return newExecutorWithError(errors.Newf("container is nil"))
	}

	cb := model.FuncConsumer(function, opts...).
		SetScope(c.Scope()).
		UpdateCallLocation(nil)
	return c.ExecutorOf(cb)
}

func (c *container) StructOf(t model.TypeVal, opts ...model.StructConsumerOption) Executor {
	if c == nil {
		return newExecutorWithError(errors.Newf("container is nil"))
	}

	cb := model.StructConsumer(t, opts...).
		SetScope(c.Scope()).
		UpdateCallLocation(nil)
	return c.ExecutorOf(cb)
}

func (c *container) ValueOf(t model.TypeVal, opts ...model.ValueConsumerOption) Executor {
	if c == nil {
		return newExecutorWithError(errors.Newf("container is nil"))
	}

	cb := model.ValueConsumer(t, opts...).
		SetScope(c.Scope()).
		UpdateCallLocation(nil)
	return c.ExecutorOf(cb)
}

func (c *container) ExecutorOf(cb model.ConsumerBuilder) Executor {
	if c == nil {
		return newExecutorWithError(errors.Newf("container is nil"))
	}

	return newExecutor(c.graph, c.storage, cb.Consumer())
}

func (c *container) newContainerWithStorage(storage *scopeStorage) *container {
	return &container{
		graph:   c.graph,
		storage: storage,
	}
}

func (c *container) Scope() model.Scope {
	if c == nil {
		return nil
	}
	return c.storage.Scope()
}

func (c *container) EnterScope(s model.Scope) (Container, error) {
	if c == nil {
		return nil, errors.Newf("container is nil")
	}

	newStorage, err := c.storage.Enter(s)
	if err != nil {
		return nil, err
	}

	return c.newContainerWithStorage(newStorage), nil
}

func (c *container) LeaveScope() Container {
	if c == nil {
		return nil
	}

	oldStorage := c.storage.Leave()
	if oldStorage == nil {
		return c
	}

	return c.newContainerWithStorage(oldStorage)
}
