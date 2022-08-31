package core

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jison/uni/core/valuer"
	"github.com/stretchr/testify/assert"

	"github.com/jison/uni/core/model"
	"github.com/jison/uni/graph"
)

type testStruct struct {
	a int
	b string
}

type testStruct2 struct {
	ts3 *testStruct3
	a   int
	b   string
}

type testStruct3 struct {
	ts2 *testStruct2
	a   int
	b   string
}

type testStruct4 struct {
	ti testInterface
	a  int
	b  string
}

type testInterface interface{}

func buildTestModule() (model.Module, model.Scope, model.Scope, model.Scope) {
	var scope1 = model.NewScope("scope1")
	var scope2 = model.NewScope("scope2", scope1)
	var scope3 = model.NewScope("scope3")

	return model.NewModule(
		model.Value(123, model.Name("name1")),
		model.Value(456, model.InScope(scope1), model.Name("name2")),
		model.Value(789, model.InScope(scope3), model.Name("name3")),
		model.Value("abc", model.Name("name4")),
		model.Value("def", model.InScope(scope2), model.Name("name5")),
		model.Struct(testStruct{}, model.InScope(scope3), model.Name("name6")),
		model.Struct(&testStruct2{}, model.InScope(scope2), model.Name("name7"),
			model.Field("a", model.ByName("name2")),
			model.Field("b", model.ByName("name4")),
		),
		model.Struct(&testStruct3{}, model.InScope(scope2), model.As((*testInterface)(nil)), model.Name("name8"),
			model.Field("a", model.ByName("name1")),
			model.Field("b", model.ByName("name5")),
		),
		model.Func(func(ti testInterface, a int, b string) *testStruct4 {
			return &testStruct4{ti, a, b}
		}, model.Return(0, model.Name("name9"))),
		model.Struct(&testStruct4{}, model.InScope(scope2), model.As((*testInterface)(nil)), model.Name("name10")),
		model.Func(func(tis []testInterface) []testInterface {
			return tis
		}, model.InScope(scope2),
			model.Param(0, model.AsCollector(true)),
			model.Return(0, model.Name("name11")),
		),
	), scope1, scope2, scope3
}

func buildTestGraphWithoutError() (DependenceGraph, model.Scope, model.Scope, model.Scope) {
	m, scope1, scope2, scope3 := buildTestModule()

	excludedComs := map[string]struct{}{
		"name3": {}, "name7": {}, "name8": {}, "name9": {}, "name10": {}, "name11": {}}

	coms := m.AllComponents().Filter(func(com model.Component) bool {
		_, ok := excludedComs[com.Name()]
		return !ok
	})
	rep := model.NewRepository(coms)
	return newDependenceGraph(rep), scope1, scope2, scope3
}

func buildTestGraphWithMissingDependencyError() (DependenceGraph, model.Scope, model.Scope, model.Scope) {
	m, scope1, scope2, scope3 := buildTestModule()

	excludedComs := map[string]struct{}{
		"name3": {}, "name7": {}, "name8": {}, "name10": {}, "name11": {}}

	coms := m.AllComponents().Filter(func(com model.Component) bool {
		_, ok := excludedComs[com.Name()]
		return !ok
	})
	rep := model.NewRepository(coms)
	return newDependenceGraph(rep), scope1, scope2, scope3
}

func buildTestGraphWithNotUniqueDependencyError() (DependenceGraph, model.Scope, model.Scope, model.Scope) {
	m, scope1, scope2, scope3 := buildTestModule()

	excludedComs := map[string]struct{}{
		"name7": {}, "name8": {}, "name9": {}, "name10": {}, "name11": {}}

	coms := m.AllComponents().Filter(func(com model.Component) bool {
		_, ok := excludedComs[com.Name()]
		return !ok
	})
	rep := model.NewRepository(coms)
	return newDependenceGraph(rep), scope1, scope2, scope3
}

func buildTestGraphWithCycleError() (DependenceGraph, model.Scope, model.Scope, model.Scope) {
	m, scope1, scope2, scope3 := buildTestModule()

	excludedComs := map[string]struct{}{
		"name3": {}, "name9": {}, "name10": {}, "name11": {}}

	coms := m.AllComponents().Filter(func(com model.Component) bool {
		_, ok := excludedComs[com.Name()]
		return !ok
	})
	rep := model.NewRepository(coms)
	return newDependenceGraph(rep), scope1, scope2, scope3
}

func buildTestGraph() (DependenceGraph, model.ComponentRepository) {
	m, _, _, _ := buildTestModule()
	rep := model.NewRepository(m.AllComponents())
	return newDependenceGraph(rep), rep
}

func comByName(rep model.ComponentRepository, name string) model.Component {
	var com model.Component
	rep.AllComponents().Iterate(func(c model.Component) bool {
		if c.Name() == name {
			com = c
			return false
		}
		return true
	})
	return com
}

func verifyDependencyNode(t *testing.T, g graph.DirectedGraphView, dep model.Dependency, agentNode Node,
	comNames []string) {
	verifyInputNode := func(node Node, comNames []string) {
		names := map[string]struct{}{}
		for _, name := range comNames {
			names[name] = struct{}{}
		}

		inputNodes := graph.NodesWithAttrsFrom(graph.PredecessorsOf(g, node))
		assert.Equal(t, len(comNames), len(inputNodes))

		inputNames := map[string]struct{}{}
		inputNodes.Iterate(func(gn graph.Node, attrs graph.AttrsView) bool {
			comVal, attrOk := attrs.Get(nodeAttrKeyComponent)
			assert.True(t, attrOk)
			com, comOk := comVal.(model.Component)
			assert.True(t, comOk)
			assert.Same(t, gn, com.Valuer())
			inputNames[com.Name()] = struct{}{}
			return true
		})

		assert.Equal(t, names, inputNames)
	}

	depNode := dep.Valuer()
	if agentNode == nil {
		verifyInputNode(depNode, comNames)
	} else {
		inputNodes := graph.NodesWithAttrsFrom(graph.PredecessorsOf(g, depNode))
		assert.Equal(t, 1, len(inputNodes))
		var agent Node
		inputNodes.Iterate(func(gn graph.Node, attrs graph.AttrsView) bool {
			var agentOk bool
			agent, agentOk = gn.(Node)
			assert.True(t, agentOk)
			assert.Equal(t, agentNode, agent)
			return false
		})

		verifyInputNode(agent, comNames)
	}
}

type depVerifyInfo struct {
	t         reflect.Type
	agentNode Node
	comNames  []string
}

func verifyComponentNode(t *testing.T, g graph.DirectedGraphView, com model.Component, depInfo []depVerifyInfo) {
	assert.NotNil(t, com)

	comNode := com.Valuer()
	comAttrs, nodeOk := g.NodeAttrs(comNode)
	assert.True(t, nodeOk)
	com2, nodeAttrOk := comAttrs.Get(nodeAttrKeyComponent)
	assert.True(t, nodeAttrOk)
	assert.Same(t, com, com2)

	provideNodes := graph.NodesWithAttrsFrom(graph.PredecessorsOf(g, comNode))
	assert.Equal(t, 1, len(provideNodes))
	var providerNode Node
	provideNodes.Iterate(func(gn graph.Node, attrs graph.AttrsView) bool {
		n, ok := gn.(Node)
		assert.True(t, ok)
		providerNode = n
		provider, proAttrOk := attrs.Get(nodeAttrKeyProvider)
		assert.True(t, proAttrOk)
		assert.Same(t, com.Provider(), provider)
		assert.Same(t, gn, com.Provider().Valuer())

		return false
	})
	g.NodeAttrs(providerNode)

	dependencyNodes := graph.NodesWithAttrsFrom(graph.PredecessorsOf(g, providerNode))
	assert.Equal(t, len(depInfo), len(dependencyNodes))
	dependencyNodes.Iterate(func(gn graph.Node, attrs graph.AttrsView) bool {
		depAttr, depAttrOk := attrs.Get(nodeAttrKeyDependency)
		assert.True(t, depAttrOk)
		dep, depOk := depAttr.(model.Dependency)
		assert.True(t, depOk)
		assert.Same(t, com.Provider(), dep.Consumer())
		assert.Same(t, gn, dep.Valuer())
		meet := false
		for _, info := range depInfo {
			if info.t == dep.Type() {
				meet = true
				verifyDependencyNode(t, g, dep, info.agentNode, info.comNames)
			}
		}
		assert.True(t, meet)

		return true
	})
}

func Test_dependenceGraph_Graph(t *testing.T) {
	g, rep := buildTestGraph()

	var name10Dep model.Dependency
	comByName(rep, "name9").Provider().Dependencies().Iterate(func(dep model.Dependency) bool {
		if dep.Type() == model.TypeOf((*testInterface)(nil)) {
			name10Dep = dep
			return false
		}
		return true
	})
	_ = name10Dep

	tests := []struct {
		comName string
		depInfo []depVerifyInfo
	}{
		{"name1", []depVerifyInfo{}},
		{"name2", []depVerifyInfo{}},
		{"name3", []depVerifyInfo{}},
		{"name4", []depVerifyInfo{}},
		{"name5", []depVerifyInfo{}},
		{"name6", []depVerifyInfo{
			{model.TypeOf(0), valuer.OneOf(), []string{"name1", "name3"}},
			{model.TypeOf(""), nil, []string{"name4"}},
		}},
		{"name7", []depVerifyInfo{
			{model.TypeOf(0), nil, []string{"name2"}},
			{model.TypeOf(""), nil, []string{"name4"}},
			{model.TypeOf(&testStruct3{}), nil, []string{"name8"}},
		}},
		{"name8", []depVerifyInfo{
			{model.TypeOf(0), nil, []string{"name1"}},
			{model.TypeOf(""), nil, []string{"name5"}},
			{model.TypeOf(&testStruct2{}), nil, []string{"name7"}},
		}},
		{"name9", []depVerifyInfo{
			{model.TypeOf(0), nil, []string{"name1"}},
			{model.TypeOf(""), nil, []string{"name4"}},
			{model.TypeOf((*testInterface)(nil)), valuer.Error(&missingError{name10Dep}),
				[]string{}},
		}},
		{"name10", []depVerifyInfo{
			{model.TypeOf(0), valuer.OneOf(), []string{"name1", "name2"}},
			{model.TypeOf(""), valuer.OneOf(), []string{"name4", "name5"}},
			{model.TypeOf((*testInterface)(nil)), nil, []string{"name8"}},
		}},
		{"name11", []depVerifyInfo{
			{model.TypeOf((*testInterface)(nil)), valuer.Collector(model.TypeOf((*testInterface)(nil))),
				[]string{"name8", "name10"}},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.comName, func(t *testing.T) {
			com := comByName(rep, tt.comName)
			verifyComponentNode(t, g.Graph(), com, tt.depInfo)
		})
	}

	cycles := graph.FindCycles(g.Graph())
	assert.Equal(t, 1, len(cycles))
	assert.Equal(t, 6, len(cycles[0]))
	nodesInCycle := map[Node]struct{}{}
	for _, gn := range cycles[0] {
		node, ok := gn.(Node)
		assert.True(t, ok)
		nodesInCycle[node] = struct{}{}
	}

	inCycle := func(n Node) bool {
		_, ok := nodesInCycle[n]
		return ok
	}

	assert.True(t, inCycle(comByName(rep, "name7").Valuer()))
	assert.True(t, inCycle(comByName(rep, "name7").Provider().Valuer()))
	assert.True(t, inCycle(comByName(rep, "name8").Valuer()))
	assert.True(t, inCycle(comByName(rep, "name8").Provider().Valuer()))
}

func Test_dependenceGraph_NodeOfComponent(t *testing.T) {
	comNames := []string{
		"name1", "name2", "name3", "name4", "name5", "name6", "name7", "name8",
		"name9", "name10", "name11",
	}

	g, rep := buildTestGraph()

	for _, name := range comNames {
		com := comByName(rep, name)
		comNode, ok := g.NodeOfComponent(com)
		assert.True(t, ok)
		assert.Same(t, com.Valuer(), comNode)
	}
}

func Test_dependenceGraph_NodeOfProvider(t *testing.T) {
	comNames := []string{
		"name1", "name2", "name3", "name4", "name5", "name6", "name7", "name8",
		"name9", "name10", "name11",
	}

	g, rep := buildTestGraph()

	for _, name := range comNames {
		com := comByName(rep, name)
		proNode, ok := g.NodeOfProvider(com.Provider())
		assert.True(t, ok)
		assert.Same(t, com.Provider().Valuer(), proNode)
	}
}

func Test_dependenceGraph_NodeOfConsumer(t *testing.T) {
	comNames := []string{
		"name1", "name2", "name3", "name4", "name5", "name6", "name7", "name8",
		"name9", "name10", "name11",
	}

	g, rep := buildTestGraph()

	for _, name := range comNames {
		com := comByName(rep, name)
		proNode, ok := g.NodeOfConsumer(com.Provider())
		assert.True(t, ok)
		assert.Same(t, com.Provider().Valuer(), proNode)
	}

	con := model.FuncConsumer(func(ti testInterface) {
		t.Logf("%v\n", ti)
	}).Consumer()
	dg, _ := g.Derive(con)

	conNode, ok := dg.NodeOfConsumer(con)
	assert.True(t, ok)
	assert.Same(t, con.Valuer(), conNode)
}

func Test_dependenceGraph_NodeOfDependency(t *testing.T) {
	comNames := []string{
		"name1", "name2", "name3", "name4", "name5", "name6", "name7", "name8",
		"name9", "name10", "name11",
	}

	g, rep := buildTestGraph()

	for _, name := range comNames {
		com := comByName(rep, name)
		com.Provider().Dependencies().Iterate(func(dep model.Dependency) bool {
			depNode, ok := g.NodeOfDependency(dep)
			assert.True(t, ok)
			assert.Same(t, dep.Valuer(), depNode)
			return true
		})
	}
}

func Test_dependenceGraph_DependencyOfNode(t *testing.T) {
	comNames := []string{
		"name1", "name2", "name3", "name4", "name5", "name6", "name7", "name8",
		"name9", "name10", "name11",
	}

	g, rep := buildTestGraph()

	for _, name := range comNames {
		com := comByName(rep, name)

		com.Provider().Dependencies().Iterate(func(dep model.Dependency) bool {
			dep2, ok := g.DependencyOfNode(dep.Valuer())
			assert.True(t, ok)
			assert.Same(t, dep, dep2)
			return true
		})
	}
}

func Test_dependenceGraph_ComponentOfNode(t *testing.T) {
	comNames := []string{
		"name1", "name2", "name3", "name4", "name5", "name6", "name7", "name8",
		"name9", "name10", "name11",
	}

	g, rep := buildTestGraph()

	for _, name := range comNames {
		com := comByName(rep, name)
		com2, ok := g.ComponentOfNode(com.Valuer())
		assert.True(t, ok)
		assert.Same(t, com, com2)
	}
}

func Test_dependenceGraph_ConsumerOfNode(t *testing.T) {
	comNames := []string{
		"name1", "name2", "name3", "name4", "name5", "name6", "name7", "name8",
		"name9", "name10", "name11",
	}

	g, rep := buildTestGraph()

	for _, name := range comNames {
		com := comByName(rep, name)
		con, ok := g.ConsumerOfNode(com.Provider().Valuer())
		assert.True(t, ok)
		assert.Same(t, com.Provider(), con)
	}

	con := model.FuncConsumer(func(ti testInterface) {
		t.Logf("%v\n", ti)
	}).Consumer()
	dg, _ := g.Derive(con)

	con2, ok := dg.ConsumerOfNode(con.Valuer())
	assert.True(t, ok)
	assert.Same(t, con, con2)
}

func Test_dependenceGraph_ProviderOfNode(t *testing.T) {
	comNames := []string{
		"name1", "name2", "name3", "name4", "name5", "name6", "name7", "name8",
		"name9", "name10", "name11",
	}

	g, rep := buildTestGraph()

	for _, name := range comNames {
		com := comByName(rep, name)
		pro, ok := g.ProviderOfNode(com.Provider().Valuer())
		assert.True(t, ok)
		assert.Same(t, com.Provider(), pro)
	}
}

func Test_dependenceGraph_CycleInfo(t *testing.T) {
	g, rep := buildTestGraph()
	cycles := g.CycleInfo().Cycles()
	assert.Equal(t, 1, len(cycles))

	for _, cycle := range cycles {
		verifyCycleIsInGraph(t, g, cycle)
	}

	verifyCyclesOfNode(t, g, comByName(rep, "name7").Valuer(), 1)
	verifyCyclesOfNode(t, g, comByName(rep, "name8").Valuer(), 1)
	verifyComponentsInOneCycle(t, g, rep, "name7", "name8")
}

func Test_dependenceGraph_Nodes(t *testing.T) {
	g, _ := buildTestGraph()

	allNodes := map[Node]struct{}{}
	g.Graph().Nodes().Iterate(func(gn graph.Node, _ graph.AttrsView) bool {
		n, ok := gn.(Node)
		assert.True(t, ok)
		allNodes[n] = struct{}{}
		return true
	})

	assert.Equal(t, len(allNodes), g.Nodes().ToSet().Len())

	g.Nodes().Iterate(func(node Node) bool {
		_, ok := allNodes[node]
		assert.True(t, ok)
		return true
	})
}

func Test_dependenceGraph_InputNodesTo(t *testing.T) {
	g, _ := buildTestGraph()
	g.Nodes().Iterate(func(node Node) bool {
		inputNodes := newNodeSet()
		graph.PredecessorsOf(g.Graph(), node).Iterate(func(gn graph.Node, _ graph.AttrsView) bool {
			inputNodes.Add(gn.(valuer.Valuer))
			return true
		})

		inputNodes2 := g.InputNodesTo(node).ToSet()
		assert.Equal(t, inputNodes, inputNodes2)
		return true
	})
}

func Test_dependenceGraph_InputComponentsToDependency(t *testing.T) {
	g, rep := buildTestGraph()

	tests := []struct {
		comName   string
		typeOfDep reflect.Type
		inputComs []string
	}{
		{"name6", model.TypeOf(""), []string{"name4"}},
		{"name6", model.TypeOf(0), []string{"name1", "name3"}},
		{"name7", model.TypeOf(""), []string{"name4"}},
		{"name7", model.TypeOf(0), []string{"name2"}},
		{"name7", model.TypeOf(&testStruct3{}), []string{"name8"}},
		{"name8", model.TypeOf(""), []string{"name5"}},
		{"name8", model.TypeOf(0), []string{"name1"}},
		{"name8", model.TypeOf(&testStruct2{}), []string{"name7"}},
		{"name9", model.TypeOf(""), []string{"name4"}},
		{"name9", model.TypeOf(0), []string{"name1"}},
		{"name10", model.TypeOf(""), []string{"name4", "name5"}},
		{"name10", model.TypeOf(0), []string{"name1", "name2"}},
		{"name10", model.TypeOf((*testInterface)(nil)), []string{"name8"}},
		{"name11", model.TypeOf((*testInterface)(nil)), []string{"name8", "name10"}},
	}

	for _, tt := range tests {
		t.Run(tt.comName, func(t *testing.T) {
			com := comByName(rep, tt.comName)
			var dep model.Dependency
			com.Provider().Dependencies().Iterate(func(d model.Dependency) bool {
				if d.Type() == tt.typeOfDep {
					dep = d
					return false
				}
				return true
			})

			comSet := map[model.Component]struct{}{}
			for _, name := range tt.inputComs {
				c := comByName(rep, name)
				comSet[c] = struct{}{}
			}

			inputComs := g.InputComponentsToDependency(dep).ToSet()
			assert.Equal(t, len(comSet), inputComs.Len())
			for c := range comSet {
				assert.True(t, inputComs.Contains(c))
			}
		})
	}
}

func Test_dependenceGraph_InputComponents(t *testing.T) {
	g, rep := buildTestGraph()

	tests := []struct {
		comName   string
		inputComs []string
	}{
		{"name6", []string{"name1", "name3", "name4"}},
		{"name7", []string{"name2", "name4", "name8"}},
		{"name8", []string{"name1", "name5", "name7"}},
		{"name9", []string{"name1", "name4"}},
		{"name10", []string{"name1", "name2", "name4", "name5", "name8"}},
		{"name11", []string{"name8", "name10"}},
	}

	for _, tt := range tests {
		t.Run(tt.comName, func(t *testing.T) {
			com := comByName(rep, tt.comName)

			comSet := map[model.Component]struct{}{}
			for _, name := range tt.inputComs {
				c := comByName(rep, name)
				comSet[c] = struct{}{}
			}

			inputComs := g.InputComponentsTo(com).ToSet()
			assert.Equal(t, len(comSet), inputComs.Len())
			for c := range comSet {
				assert.True(t, inputComs.Contains(c))
			}
		})
	}
}

func Test_dependenceGraph_Validate(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		g, _, _, _ := buildTestGraphWithoutError()
		err := g.Validate()
		t.Logf("%v\n", err)
		assert.Nil(t, err)
	})

	t.Run("miss dependency", func(t *testing.T) {
		g, _, _, _ := buildTestGraphWithMissingDependencyError()
		err := g.Validate()
		t.Logf("%v\n", err)
		assert.NotNil(t, err)
	})

	t.Run("not unique dependencies", func(t *testing.T) {
		g, _, _, _ := buildTestGraphWithNotUniqueDependencyError()
		err := g.Validate()
		t.Logf("%v\n", err)
		assert.NotNil(t, err)
	})

	t.Run("cycle", func(t *testing.T) {
		g, _, _, _ := buildTestGraphWithCycleError()
		err := g.Validate()
		t.Logf("%v\n", err)
		assert.NotNil(t, err)
	})
}

func Test_dependenceGraph_Derive(t *testing.T) {
	m, _, scope2, _ := buildTestModule()
	rep := model.NewRepository(m.AllComponents())
	g := newDependenceGraph(rep)
	con := model.FuncConsumer(func([]testInterface) {}, model.InScope(scope2)).Consumer()
	dg, conNode := g.Derive(con)

	assert.Equal(t, con.Valuer(), conNode)

	t.Run("Graph", func(t *testing.T) {
		_, ok := g.Graph().NodeAttrs(conNode)
		assert.False(t, ok)
		_, ok = dg.Graph().NodeAttrs(conNode)
		assert.True(t, ok)

		con.Dependencies().Iterate(func(dep model.Dependency) bool {
			_, ok = g.Graph().NodeAttrs(dep.Valuer())
			assert.False(t, ok)
			_, ok = dg.Graph().NodeAttrs(dep.Valuer())
			assert.True(t, ok)
			return true
		})

		g.Graph().Nodes().Iterate(func(node graph.Node, _ graph.AttrsView) bool {
			_, ok = dg.Graph().NodeAttrs(node)
			assert.True(t, ok)
			return true
		})
	})

	t.Run("NodeOfComponent", func(t *testing.T) {
		coms := rep.AllComponents()
		coms.Iterate(func(com model.Component) bool {
			node, ok := g.NodeOfComponent(com)
			assert.True(t, ok)
			assert.Same(t, com.Valuer(), node)

			node, ok = dg.NodeOfComponent(com)
			assert.True(t, ok)
			assert.Same(t, com.Valuer(), node)

			return true
		})
	})

	t.Run("NodeOfProvider", func(t *testing.T) {
		coms := rep.AllComponents()
		coms.Iterate(func(com model.Component) bool {
			node, ok := g.NodeOfProvider(com.Provider())
			assert.True(t, ok)
			assert.Same(t, com.Provider().Valuer(), node)

			node, ok = dg.NodeOfProvider(com.Provider())
			assert.True(t, ok)
			assert.Same(t, com.Provider().Valuer(), node)

			return true
		})
	})

	t.Run("NodeOfConsumer", func(t *testing.T) {
		coms := rep.AllComponents()
		coms.Iterate(func(com model.Component) bool {
			node, ok := g.NodeOfConsumer(com.Provider())
			assert.True(t, ok)
			assert.Same(t, com.Provider().Valuer(), node)

			node, ok = dg.NodeOfConsumer(com.Provider())
			assert.True(t, ok)
			assert.Same(t, com.Provider().Valuer(), node)

			return true
		})

		node, ok := g.NodeOfConsumer(con)
		assert.False(t, ok)
		node, ok = dg.NodeOfConsumer(con)
		assert.True(t, ok)
		assert.Equal(t, con.Valuer(), node)
	})

	t.Run("NodeOfDependency", func(t *testing.T) {
		coms := rep.AllComponents()
		coms.Iterate(func(com model.Component) bool {
			com.Provider().Dependencies().Iterate(func(dep model.Dependency) bool {
				node, ok := g.NodeOfDependency(dep)
				assert.True(t, ok)
				assert.Same(t, dep.Valuer(), node)

				node, ok = dg.NodeOfDependency(dep)
				assert.True(t, ok)
				assert.Same(t, dep.Valuer(), node)
				return true
			})
			return true
		})

		con.Dependencies().Iterate(func(dep model.Dependency) bool {
			node, ok := g.NodeOfDependency(dep)
			assert.False(t, ok)

			node, ok = dg.NodeOfDependency(dep)
			assert.True(t, ok)
			assert.Same(t, dep.Valuer(), node)
			return true
		})
	})

	t.Run("DependencyOfNode", func(t *testing.T) {
		coms := rep.AllComponents()
		coms.Iterate(func(com model.Component) bool {
			com.Provider().Dependencies().Iterate(func(dep model.Dependency) bool {
				val, ok := g.DependencyOfNode(dep.Valuer())
				assert.True(t, ok)
				assert.Same(t, dep, val)

				val, ok = dg.DependencyOfNode(dep.Valuer())
				assert.True(t, ok)
				assert.Same(t, dep, val)
				return true
			})
			return true
		})

		con.Dependencies().Iterate(func(dep model.Dependency) bool {
			val, ok := g.DependencyOfNode(dep.Valuer())
			assert.False(t, ok)

			val, ok = dg.DependencyOfNode(dep.Valuer())
			assert.True(t, ok)
			assert.Same(t, dep, val)
			return true
		})
	})

	t.Run("ConsumerOfNode", func(t *testing.T) {
		coms := rep.AllComponents()
		coms.Iterate(func(com model.Component) bool {
			val, ok := g.ConsumerOfNode(com.Provider().Valuer())
			assert.True(t, ok)
			assert.Same(t, com.Provider(), val)

			val, ok = dg.ConsumerOfNode(com.Provider().Valuer())
			assert.True(t, ok)
			assert.Same(t, com.Provider(), val)

			return true
		})

		val, ok := g.ConsumerOfNode(con.Valuer())
		assert.False(t, ok)
		val, ok = dg.ConsumerOfNode(con.Valuer())
		assert.True(t, ok)
		assert.Equal(t, con, val)
	})

	t.Run("ProviderOfNode", func(t *testing.T) {
		coms := rep.AllComponents()
		coms.Iterate(func(com model.Component) bool {
			val, ok := g.ProviderOfNode(com.Provider().Valuer())
			assert.True(t, ok)
			assert.Same(t, com.Provider(), val)

			val, ok = dg.ProviderOfNode(com.Provider().Valuer())
			assert.True(t, ok)
			assert.Same(t, com.Provider(), val)

			return true
		})

		_, ok := g.ProviderOfNode(con.Valuer())
		assert.False(t, ok)
		_, ok = dg.ProviderOfNode(con.Valuer())
		assert.False(t, ok)
	})

	t.Run("ComponentOfNode", func(t *testing.T) {
		coms := rep.AllComponents()
		coms.Iterate(func(com model.Component) bool {
			val, ok := g.ComponentOfNode(com.Valuer())
			assert.True(t, ok)
			assert.Same(t, com, val)

			val, ok = dg.ComponentOfNode(com.Valuer())
			assert.True(t, ok)
			assert.Same(t, com, val)

			return true
		})
	})

	t.Run("CycleInfo", func(t *testing.T) {
		cycles1 := g.CycleInfo().Cycles()
		assert.Equal(t, 1, len(cycles1))
		cycle1 := cycles1[0]

		cycles2 := dg.CycleInfo().Cycles()
		assert.Equal(t, 1, len(cycles2))
		cycle2 := cycles2[0]

		assert.True(t, cycleEqual(cycle1, cycle2))
	})

	t.Run("Nodes", func(t *testing.T) {
		gNodes := g.Nodes().ToSet()
		dgNodes := dg.Nodes().ToSet()

		con.Dependencies().Iterate(func(dep model.Dependency) bool {
			assert.False(t, gNodes.Contains(dep.Valuer()))
			assert.True(t, dgNodes.Contains(dep.Valuer()))
			return true
		})

		assert.False(t, gNodes.Contains(conNode))
		assert.True(t, dgNodes.Contains(conNode))
	})

	t.Run("InputNodesTo", func(t *testing.T) {
		g.Nodes().Iterate(func(node Node) bool {
			inputNodes := g.InputNodesTo(node).ToSet()
			inputNodes2 := dg.InputNodesTo(node).ToSet()

			assert.Equal(t, inputNodes, inputNodes2)
			return true
		})

		inputNodes := g.InputNodesTo(con.Valuer()).ToSet()
		assert.Equal(t, 0, inputNodes.Len())

		inputNodes = dg.InputNodesTo(con.Valuer()).ToSet()
		assert.Equal(t, 1, inputNodes.Len())

		con.Dependencies().Iterate(func(dep model.Dependency) bool {
			inputNodes = g.InputNodesTo(dep.Valuer()).ToSet()
			assert.Equal(t, 0, inputNodes.Len())

			inputNodes = dg.InputNodesTo(dep.Valuer()).ToSet()
			assert.Equal(t, 1, inputNodes.Len())
			return true
		})
	})

	t.Run("InputComponentsToDependency", func(t *testing.T) {

		coms := rep.AllComponents()
		coms.Iterate(func(com model.Component) bool {
			com.Provider().Dependencies().Iterate(func(dep model.Dependency) bool {
				inputComs1 := g.InputComponentsToDependency(dep).ToSet()
				inputComs2 := dg.InputComponentsToDependency(dep).ToSet()

				assert.Equal(t, inputComs1, inputComs2)
				return true
			})
			return true
		})

		con.Dependencies().Iterate(func(dep model.Dependency) bool {
			inputComs := g.InputComponentsToDependency(dep).ToSet()
			assert.Equal(t, 0, inputComs.Len())

			inputComs = dg.InputComponentsToDependency(dep).ToSet()
			assert.Equal(t, 1, inputComs.Len())
			return true
		})
	})

	t.Run("InputComponentsTo", func(t *testing.T) {
		coms := rep.AllComponents()
		coms.Iterate(func(com model.Component) bool {
			inputComs1 := g.InputComponentsTo(com).ToSet()
			inputComs2 := dg.InputComponentsTo(com).ToSet()

			assert.Equal(t, inputComs1, inputComs2)
			return true
		})
	})

	t.Run("Validate", func(t *testing.T) {
		//module, _, scope2, scope3 := buildTestModule()

		t.Run("no errors", func(t *testing.T) {
			g2, _, _, scope3 := buildTestGraphWithoutError()

			con2 := model.FuncConsumer(func(testStruct) {}, model.InScope(scope3)).Consumer()
			dg2, _ := g2.Derive(con2)
			err := dg2.Validate()
			fmt.Printf("%v\n", err)
			assert.Nil(t, err)
		})

		t.Run("error in derived graph", func(t *testing.T) {
			t.Run("miss dependency", func(t *testing.T) {
				g2, _, _, scope3 := buildTestGraphWithMissingDependencyError()

				con2 := model.FuncConsumer(func(testStruct, testInterface) {}, model.InScope(scope3)).Consumer()
				dg2, _ := g2.Derive(con2)
				err := dg2.Validate()
				assert.NotNil(t, err)
			})

			t.Run("not unique dependencies", func(t *testing.T) {
				g2, _, scope22, _ := buildTestGraphWithNotUniqueDependencyError()

				con2 := model.FuncConsumer(func(int) {}, model.InScope(scope22)).Consumer()
				dg2, _ := g2.Derive(con2)

				err := dg2.Validate()
				assert.NotNil(t, err)
			})
		})

		t.Run("error in parent graph", func(t *testing.T) {
			t.Run("miss dependency", func(t *testing.T) {
				g2, _, _, scope3 := buildTestGraphWithMissingDependencyError()

				con2 := model.FuncConsumer(func(testStruct, testInterface) {}, model.InScope(scope3)).Consumer()
				dg2, _ := g2.Derive(con2)
				err := dg2.Validate()
				t.Logf("%v\n", err)
				assert.NotNil(t, err)
			})

			t.Run("not unique dependencies", func(t *testing.T) {
				g2, _, _, scope3 := buildTestGraphWithNotUniqueDependencyError()

				con2 := model.FuncConsumer(func(testStruct, testInterface) {}, model.InScope(scope3)).Consumer()
				dg2, _ := g2.Derive(con2)
				err := dg2.Validate()
				t.Logf("%v\n", err)
				assert.NotNil(t, err)
			})

			t.Run("cycle", func(t *testing.T) {
				g2, _, _, scope3 := buildTestGraphWithCycleError()

				con2 := model.FuncConsumer(func(testStruct, testInterface) {}, model.InScope(scope3)).Consumer()
				dg2, _ := g2.Derive(con2)
				err := dg2.Validate()
				t.Logf("%v\n", err)
				assert.NotNil(t, err)
			})
		})
	})
}
