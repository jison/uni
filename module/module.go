package module

import (
	"reflect"
	"runtime"
	"sync"
)

type moduleComponentList map[ComponentList]struct{}

func (l moduleComponentList) Each(f func(Component)) {
	for c := range l {
		c.Each(f)
	}
}

func (l moduleComponentList) Merge(ol ComponentList) moduleComponentList {
	switch r := ol.(type) {

	case moduleComponentList:
		modules := make(map[ComponentList]struct{}, len(l)+len(r))
		for c := range l {
			modules[c] = struct{}{}
		}
		for c := range r {
			modules[c] = struct{}{}
		}
		return moduleComponentList(modules)

	case *emptyComponentList:
		return l
	default:
		modules := make(map[ComponentList]struct{}, len(l)+1)
		for c := range l {
			modules[c] = struct{}{}
		}
		modules[r] = struct{}{}
		return moduleComponentList(modules)
	}
}

func asModuleComponents(l ComponentList) moduleComponentList {
	if mcl, ok := l.(moduleComponentList); ok {
		return mcl
	}
	return moduleComponentList(map[ComponentList]struct{}{l: {}})
}

type module struct {
	subModules map[Module]struct{}
	providers  map[Provider]struct{}

	_componentsInited sync.Once
	_components       ComponentList
}

var _ Module = &module{}

func (m *module) Components() ComponentList {
	m._componentsInited.Do(func() {
		arr := make([]Component, 0)
		for p := range m.providers {
			p.Components().Each(func(c Component) {
				arr = append(arr, c)
			})
		}
		mcl := asModuleComponents(arrayComponentList(arr))

		for subModule := range m.subModules {
			mcl = mcl.Merge(subModule.Components())
		}

		m._components = mcl
	})

	return m._components
}

// uni.NewModule(
// 	uni.ProvideModule(nil),
// 	uni.ProvideFunc(
// 		func(){},
// 		uni.WithName(),
// 		uni.Param(0, uni.WithName(""), uni.Optional()),
// 		uni.Component(0, uni.WithName(""), uni.As(interface{}))
// 	)
// )

func NewModule(opts ...moduleOption) Module {
	m := &module{}
	for _, opt := range opts {
		opt.applyModule(m)
	}
	return m
}

type moduleOption interface {
	applyModule(*module)
}

type provideModule struct {
	subModule Module
}

func (pm *provideModule) applyModule(m *module) {
	m.subModules[m] = struct{}{}
}

func ProvideModule(subModule Module) *provideModule {
	return &provideModule{subModule: subModule}
}

func (p *provider) applyModule(m *module) {
	m.providers[p] = struct{}{}
}

func ProvideFunc(f interface{}, opts ...providerOption) *provider {
	rval := reflect.ValueOf(f)
	p := providerOfFunc(rval)

	for _, o := range opts {
		o.applyProvider(p)
	}
	return p
}

func ProvideStruct(s interface{}, opts ...providerOption) *provider {
	rval := reflect.ValueOf(s)
	p := providerOfStruct(rval)

	for _, o := range opts {
		o.applyProvider(p)
	}
	return p
}

func ProvideValue(v interface{}, opts ...providerOption) *provider {
	rval := reflect.ValueOf(v)
	p := providerOfValue(rval)

	for _, o := range opts {
		o.applyProvider(p)
	}
	return p
}

type ModuleBuilder interface {
	ProvideFunc(f interface{}, opts ...providerOption) ModuleBuilder
	ProvideStruct(s interface{}, opts ...providerOption) ModuleBuilder
	ProvideValue(v interface{}, opts ...providerOption) ModuleBuilder

	ProvideModule(subModule Module) ModuleBuilder
	Module() Module
}

func NewModuleBuilder() ModuleBuilder {
	return &moduleBuilder{opts: make([]moduleOption, 0)}
}

type moduleBuilder struct {
	opts []moduleOption
}

func (mb *moduleBuilder) ProvideModule(m Module) ModuleBuilder {
	mb.opts = append(mb.opts, ProvideModule(m))
	return mb
}

func (mb *moduleBuilder) ProvideFunc(f interface{}, opts ...providerOption) ModuleBuilder {
	mb.opts = append(mb.opts, ProvideFunc(f, opts...))
	return mb
}

func (mb *moduleBuilder) ProvideStruct(s interface{}, opts ...providerOption) ModuleBuilder {
	mb.opts = append(mb.opts, ProvideStruct(s, opts...))
	return mb
}

func (mb *moduleBuilder) ProvideValue(v interface{}, opts ...providerOption) ModuleBuilder {
	mb.opts = append(mb.opts, ProvideValue(v, opts...))
	return mb
}

func (mb *moduleBuilder) Module() Module {
	return NewModule(mb.opts...)
}

func locationOfCaller() location {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return location{}
	}

	outside := runtime.FuncForPC(pc)
	outside.FileLine(pc)
	funcName := outside.Name()
	return location{file: outside.}
}
