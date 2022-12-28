package core

import (
	"github.com/jison/uni/core/model"
	"github.com/jison/uni/core/valuer"
	"github.com/jison/uni/internal/errors"
)

type Executor interface {
	Execute() (interface{}, error)
}

type executor struct {
	graph     DependenceGraph
	cycleInfo DependenceCycleInfo
	storage   ScopeBaseStorage
	node      Node
}

func (e *executor) Execute() (interface{}, error) {
	val := e.getValueOfNode(e.node, e.storage, NewPath(e.graph))
	return val.Interface()
}

func (e *executor) getValueOfNode(node Node, storage ScopeBaseStorage, stack Path) valuer.Value {
	cycles := e.cycleInfo.CyclesOfNode(node)
	if len(cycles) > 0 {
		err := errors.Newf("there are cycles in the dependence path. %v", cycles)
		return valuer.ErrorValue(err)
	}

	nodeScope := e.scopeOfNode(node)

	return storage.GetOrElse(node, nodeScope, func(s ScopeBaseStorage) valuer.Value {
		nodeStack := stack.Append(node)
		var params []valuer.Value
		e.graph.InputNodesTo(node).Each(func(inputNode Node) {
			params = append(params, e.getValueOfNode(inputNode, s, nodeStack))
		})
		nodeVal := node.Value(params)

		provider, isProvider := e.graph.ProviderOfNode(node)
		if isProvider {
			return nodeVal
		}
		err, isErr := nodeVal.AsError()
		if !isErr {
			return nodeVal
		}
		providerErr := errors.Newf("%+v", provider).AddErrors(err)
		return valuer.ErrorValue(providerErr)
	})
}

func (e *executor) scopeOfNode(node Node) model.Scope {
	if provider, ok := e.graph.ProviderOfNode(node); ok {
		return provider.Scope()
	}

	if com, ok := e.graph.ComponentOfNode(node); ok {
		return com.Provider().Scope()
	}

	return nil
}

type errorExecutor struct {
	err error
}

func (e *errorExecutor) Execute() (interface{}, error) {
	return nil, e.err
}

func newExecutor(
	dg DependenceGraph,
	storage ScopeBaseStorage,
	consumer model.Consumer,
) Executor {
	if err := consumer.Validate(); err != nil {
		return newExecutorWithError(err)
	}

	g, consumerNode := dg.Derive(consumer)

	return &executor{
		graph:     g,
		cycleInfo: dg.CycleInfo(),
		storage:   storage,
		node:      consumerNode,
	}
}

func newExecutorWithError(err error) Executor {
	return &errorExecutor{err: err}
}
