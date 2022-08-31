package core

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/jison/uni/core/model"
	"github.com/jison/uni/core/valuer"
	"github.com/jison/uni/graph"
	"github.com/jison/uni/internal/errors"
)

type Node valuer.Valuer

type nodeAttrKey int

const (
	nodeAttrKeyProvider   nodeAttrKey = 1
	nodeAttrKeyConsumer   nodeAttrKey = 2
	nodeAttrKeyComponent  nodeAttrKey = 3
	nodeAttrKeyDependency nodeAttrKey = 4
)

func (k nodeAttrKey) String() string {
	switch k {
	case nodeAttrKeyProvider:
		return "provider"
	case nodeAttrKeyConsumer:
		return "consumer"
	case nodeAttrKeyComponent:
		return "component"
	case nodeAttrKeyDependency:
		return "dependency"
	}
	return ""
}

type graphNodeIterator struct {
	gi graph.NodeAndAttrsIterator
}

func (c *graphNodeIterator) Iterate(f func(Node) bool) bool {
	return c.gi.Iterate(func(node graph.Node, _ graph.AttrsView) bool {
		valNode, ok := node.(Node)
		if ok {
			if !f(valNode) {
				return false
			}
		}
		return true
	})
}

type DependenceGraph interface {
	Graph() graph.DirectedGraphView

	DependencyOfNode(node Node) (model.Dependency, bool)
	ComponentOfNode(node Node) (model.Component, bool)
	ConsumerOfNode(node Node) (model.Consumer, bool)
	ProviderOfNode(node Node) (model.Provider, bool)

	NodeOfDependency(dep model.Dependency) (Node, bool)
	NodeOfComponent(com model.Component) (Node, bool)
	NodeOfConsumer(consumer model.Consumer) (Node, bool)
	NodeOfProvider(provider model.Provider) (Node, bool)

	Derive(consumer model.Consumer) (DependenceGraph, Node)

	CycleInfo() DependenceCycleInfo

	Nodes() NodeCollection
	InputNodesTo(node Node) NodeCollection
	InputComponentsToDependency(dep model.Dependency) model.ComponentCollection
	InputComponentsTo(com model.Component) model.ComponentCollection

	Validate() error
	MissingError() error
	UncertainError() error
	CycleError() error
}

type dependenceGraph struct {
	parent *dependenceGraph

	graph      graph.DirectedGraph
	repository model.ComponentRepository

	nodeByComponent  map[model.Component]Node
	nodeByConsumer   map[model.Consumer]Node
	nodeByDependency map[model.Dependency]Node

	missingDependencies   []model.Dependency
	uncertainDependencies []model.Dependency

	cycleInfoInitOnce sync.Once
	cycleInfo         DependenceCycleInfo
}

func (dg *dependenceGraph) Graph() graph.DirectedGraphView {
	return dg.graph
}

func (dg *dependenceGraph) NodeOfComponent(com model.Component) (Node, bool) {
	if dg.parent != nil {
		if n, ok := dg.parent.NodeOfComponent(com); ok {
			return n, true
		}
	}

	n, ok := dg.nodeByComponent[com]
	return n, ok
}

func (dg *dependenceGraph) NodeOfProvider(provider model.Provider) (Node, bool) {
	if dg.parent != nil {
		if n, ok := dg.parent.NodeOfConsumer(provider); ok {
			return n, true
		}
	}

	n, ok := dg.nodeByConsumer[provider]
	return n, ok
}

func (dg *dependenceGraph) NodeOfConsumer(consumer model.Consumer) (Node, bool) {
	if dg.parent != nil {
		if n, ok := dg.parent.NodeOfConsumer(consumer); ok {
			return n, true
		}
	}

	n, ok := dg.nodeByConsumer[consumer]
	return n, ok
}

func (dg *dependenceGraph) NodeOfDependency(dep model.Dependency) (Node, bool) {
	if dg.parent != nil {
		if n, ok := dg.parent.NodeOfDependency(dep); ok {
			return n, ok
		}
	}

	n, ok := dg.nodeByDependency[dep]
	return n, ok
}

func (dg *dependenceGraph) addNodeOfComponent(com model.Component) Node {
	comNode, ok := dg.NodeOfComponent(com)
	if ok {
		return comNode
	}

	// add component node first
	// to prevent infinity loop in case of there is a cyclic graph
	comNode = com.Valuer()
	dg.graph.AddNodeWithAttrs(comNode, graph.Attrs{nodeAttrKeyComponent: com})
	dg.nodeByComponent[com] = comNode

	providerNode := dg.addNodeOfProvider(com.Provider())
	graph.AddEdge(dg.graph, providerNode, comNode)

	return comNode
}

func (dg *dependenceGraph) addNodeOfProvider(provider model.Provider) Node {
	consumerNode := dg.addNodeOfConsumer(provider)
	if nodeAttrs, ok := dg.graph.NodeAttrs(consumerNode); ok {
		nodeAttrs.Del(nodeAttrKeyConsumer)
		nodeAttrs.Set(nodeAttrKeyProvider, provider)
	}

	return consumerNode
}

func (dg *dependenceGraph) addNodeOfConsumer(consumer model.Consumer) Node {
	consumerNode, ok := dg.NodeOfConsumer(consumer)
	if ok {
		return consumerNode
	}

	consumerNode = consumer.Valuer()
	dg.graph.AddNodeWithAttrs(consumerNode, graph.Attrs{nodeAttrKeyConsumer: consumer})

	consumer.Dependencies().Iterate(func(dep model.Dependency) bool {
		depNode := dg.addNodeOfDependency(dep)
		graph.AddEdge(dg.graph, depNode, consumerNode)
		return true
	})

	dg.nodeByConsumer[consumer] = consumerNode
	return consumerNode
}

func (dg *dependenceGraph) addNodeOfDependency(dep model.Dependency) Node {
	depNode, ok := dg.NodeOfDependency(dep)
	if ok {
		return depNode
	}

	depNode = dep.Valuer()
	dg.graph.AddNodeWithAttrs(depNode, graph.Attrs{nodeAttrKeyDependency: dep})

	inputNode := dg.getInputNodeOfDependency(dep)
	graph.AddEdge(dg.graph, inputNode, depNode)

	dg.nodeByDependency[dep] = depNode
	return depNode
}

func (dg *dependenceGraph) getInputNodeOfDependency(dep model.Dependency) Node {
	components := dg.repository.ComponentsMatchDependency(dep)
	if dep.IsCollector() {
		return dg.addCollectorNodeOf(dep, components)
	}

	comArr := components.ToArray()
	if len(comArr) == 0 {
		return dg.addMissingNodeOf(dep)
	} else if len(comArr) == 1 {
		return dg.addNodeOfComponent(comArr[0])
	} else {
		return dg.addOneOfNodeOf(dep, components)
	}
}

func (dg *dependenceGraph) addCollectorNodeOf(dep model.Dependency, coms model.ComponentCollection) Node {
	collectorValuer := valuer.Collector(dep.Type())
	coms.Each(func(com model.Component) {
		comNode := dg.addNodeOfComponent(com)
		graph.AddEdge(dg.graph, comNode, collectorValuer)
	})
	return collectorValuer
}

func (dg *dependenceGraph) addMissingNodeOf(dep model.Dependency) Node {
	if dep.Optional() {
		val := reflect.Zero(dep.Type())
		return valuer.Const(val)
	} else {
		dg.missingDependencies = append(dg.missingDependencies, dep)

		v := valuer.Error(&missingError{dep})
		return v
	}
}

func (dg *dependenceGraph) addOneOfNodeOf(dep model.Dependency, coms model.ComponentCollection) Node {
	dg.uncertainDependencies = append(dg.uncertainDependencies, dep)

	oneOfNode := valuer.OneOf()
	coms.Each(func(com model.Component) {
		candidateNode := dg.addNodeOfComponent(com)
		graph.AddEdge(dg.graph, candidateNode, oneOfNode)
	})

	return oneOfNode
}

func (dg *dependenceGraph) Derive(consumer model.Consumer) (DependenceGraph, Node) {
	derived := &dependenceGraph{
		parent:           dg,
		graph:            graph.DeriveDirectedGraph(dg.graph),
		repository:       dg.repository,
		nodeByComponent:  map[model.Component]Node{},
		nodeByConsumer:   map[model.Consumer]Node{},
		nodeByDependency: map[model.Dependency]Node{},
	}

	consumerNode := derived.addNodeOfConsumer(consumer)

	return derived, consumerNode
}

func (dg *dependenceGraph) attrOfNode(node Node, key nodeAttrKey) (interface{}, bool) {
	var attrs graph.AttrsView
	var ok bool
	if attrs, ok = dg.graph.NodeAttrs(node); !ok {
		return nil, false
	}

	var val interface{}
	if val, ok = attrs.Get(key); !ok {
		return nil, false
	}

	return val, true
}

func (dg *dependenceGraph) DependencyOfNode(node Node) (model.Dependency, bool) {
	val, ok := dg.attrOfNode(node, nodeAttrKeyDependency)
	if !ok {
		return nil, false
	}

	res, ok := val.(model.Dependency)
	return res, ok
}

func (dg *dependenceGraph) ComponentOfNode(node Node) (model.Component, bool) {
	val, ok := dg.attrOfNode(node, nodeAttrKeyComponent)
	if !ok {
		return nil, false
	}

	res, ok := val.(model.Component)
	return res, ok
}

func (dg *dependenceGraph) ConsumerOfNode(node Node) (model.Consumer, bool) {
	var val interface{}
	var ok bool
	var con model.Consumer

	val, ok = dg.attrOfNode(node, nodeAttrKeyProvider)
	if ok {
		con, ok = val.(model.Consumer)
		if ok {
			return con, true
		}
	}

	val, ok = dg.attrOfNode(node, nodeAttrKeyConsumer)
	if ok {
		con, ok = val.(model.Consumer)
		if ok {
			return con, true
		}
	}

	return nil, false
}

func (dg *dependenceGraph) ProviderOfNode(node Node) (model.Provider, bool) {
	val, ok := dg.attrOfNode(node, nodeAttrKeyProvider)
	if !ok {
		return nil, false
	}

	res, ok := val.(model.Provider)
	return res, ok
}

func (dg *dependenceGraph) CycleInfo() DependenceCycleInfo {
	if dg.parent != nil && len(dg.nodeByComponent) == 0 {
		return dg.parent.CycleInfo()
	}

	dg.cycleInfoInitOnce.Do(func() {
		dg.cycleInfo = buildCycleInfoOf(dg)
	})

	return dg.cycleInfo
}

func (dg *dependenceGraph) Nodes() NodeCollection {
	return NewNodeCollection(&graphNodeIterator{dg.graph.Nodes()})
}

func (dg *dependenceGraph) InputNodesTo(node Node) NodeCollection {
	gi := graph.GetNodesInDirectionMatch(node,
		func(node graph.Node) graph.NodeAndAttrsIterator {
			return graph.PredecessorsOf(dg.graph, node)
		},
		func(gNode graph.Node, attrs graph.AttrsView) bool {
			_, ok := gNode.(valuer.Valuer)
			return ok
		},
	)

	return NewNodeCollection(&graphNodeIterator{gi})
}

func (dg *dependenceGraph) InputComponentsToDependency(dep model.Dependency) model.ComponentCollection {
	depNode, depExist := dg.NodeOfDependency(dep)
	if !depExist {
		return model.EmptyComponents()
	}

	getComponentFromGraphNode := func(gNode graph.Node) (model.Component, bool) {
		var ok bool
		var valNode Node
		if valNode, ok = gNode.(valuer.Valuer); !ok {
			return nil, false
		}

		var com model.Component
		if com, ok = dg.ComponentOfNode(valNode); !ok {
			return nil, false
		}

		return com, true
	}

	gi := graph.GetNodesInDirectionMatch(depNode,
		func(node graph.Node) graph.NodeAndAttrsIterator {
			return graph.PredecessorsOf(dg.graph, node)
		},
		func(gNode graph.Node, attrs graph.AttrsView) bool {
			_, ok := getComponentFromGraphNode(gNode)
			return ok
		},
	)

	return model.ComponentsOfIterator(model.FuncComponentIterator(func(f func(model.Component) bool) bool {
		return gi.Iterate(func(gNode graph.Node, _ graph.AttrsView) bool {
			com, ok := getComponentFromGraphNode(gNode)
			if ok {
				if !f(com) {
					return false
				}
			}
			return true
		})
	}))
}

func (dg *dependenceGraph) InputComponentsTo(com model.Component) model.ComponentCollection {
	var coms []model.ComponentIterator
	com.Provider().Dependencies().Iterate(func(d model.Dependency) bool {
		cc := dg.InputComponentsToDependency(d)
		coms = append(coms, cc)
		return true
	})

	return model.CombineComponents(coms...).Distinct()
}

func (dg *dependenceGraph) allMissingDependencies() model.DependencyIterator {
	selfDeps := model.ArrayDependencyIterator(dg.missingDependencies)
	if dg.parent == nil {
		return selfDeps
	}

	return model.CombineDependencyIterators(dg.parent.allMissingDependencies(), selfDeps)
}

func (dg *dependenceGraph) MissingError() error {
	errs := errors.Empty()
	dg.allMissingDependencies().Iterate(func(dep model.Dependency) bool {
		err := errors.Newf("%v in %v at %v", dep, dep.Consumer().Scope(), dep.Consumer().Location())
		errs = errs.AddErrors(err)
		return true
	})

	if errs.HasError() {
		return errs.WithMainf("can not find component match these dependencies")
	}

	return nil
}

func (dg *dependenceGraph) allUncertainDependencies() model.DependencyIterator {
	selfDeps := model.ArrayDependencyIterator(dg.uncertainDependencies)
	if dg.parent == nil {
		return selfDeps
	}

	return model.CombineDependencyIterators(dg.parent.allUncertainDependencies(), selfDeps)
}

func (dg *dependenceGraph) UncertainError() error {
	errs := errors.Empty()
	dg.allUncertainDependencies().Iterate(func(dep model.Dependency) bool {
		notUniqueErr := errors.Empty()

		dg.InputComponentsToDependency(dep).Each(func(com model.Component) {
			notUniqueErr = notUniqueErr.AddErrorf("%v at %v", com, com.Provider().Location())
		})

		if notUniqueErr.HasError() {
			errs = errs.AddErrors(notUniqueErr.WithMainf("%v at %v", dep, dep.Consumer().Location()))
		}

		return true
	})

	if errs.HasError() {
		return errs.WithMainf("these dependencies are more than one component match")
	}

	return nil
}

func (dg *dependenceGraph) CycleError() error {
	errs := errors.Empty()
	cycles := dg.CycleInfo().Cycles()
	for _, cycle := range cycles {
		errs = errs.AddErrorf("%v", cycle)
	}

	if errs.HasError() {
		return errs.WithMainf("there are dependence cycles")
	}

	return nil
}

func (dg *dependenceGraph) Validate() error {
	errs := errors.Empty()

	if err := dg.MissingError(); err != nil {
		errs = errs.AddErrors(err)
	}

	if err := dg.UncertainError(); err != nil {
		errs = errs.AddErrors(err)
	}

	if err := dg.CycleError(); err != nil {
		errs = errs.AddErrors(err)
	}

	if errs.HasError() {
		return errs
	}

	return nil
}

func newDependenceGraph(rep model.ComponentRepository) DependenceGraph {
	g := &dependenceGraph{
		parent:           nil,
		graph:            graph.NewDirectedGraph(),
		repository:       rep,
		nodeByComponent:  map[model.Component]Node{},
		nodeByConsumer:   map[model.Consumer]Node{},
		nodeByDependency: map[model.Dependency]Node{},
	}

	rep.AllComponents().Iterate(func(com model.Component) bool {
		g.addNodeOfComponent(com)
		return true
	})

	return g
}

type missingError struct {
	dependency model.Dependency
}

func (e *missingError) Error() string {
	return fmt.Sprintf("can not find components that match %v in %v", e.dependency,
		e.dependency.Consumer().Scope())
}
