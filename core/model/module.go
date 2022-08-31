package model

import (
	"reflect"

	"github.com/jison/uni/internal/errors"
)

type ModuleIterator interface {
	Iterate(func(Module) bool) bool
}

type moduleSet map[Module]struct{}

func (s moduleSet) Iterate(f func(Module) bool) bool {
	for m := range s {
		if !f(m) {
			return false
		}
	}
	return true
}

type ProviderIterator interface {
	Iterate(func(Provider) bool) bool
}

type providerSet map[Provider]struct{}

func (s providerSet) Iterate(f func(Provider) bool) bool {
	for p := range s {
		if !f(p) {
			return false
		}
	}
	return true
}

type Module interface {
	SubModules() ModuleIterator
	Providers() ProviderIterator

	AllModules() ModuleIterator
	AllProviders() ProviderIterator
	AllComponents() ComponentCollection

	Validate() error
}

type module struct {
	subModules moduleSet
	providers  providerSet
}

func newModule(modules []Module, pbs []ProviderBuilder) *module {
	ms := moduleSet{}
	for _, m := range modules {
		if m == nil {
			continue
		}
		ms[m] = struct{}{}
	}

	ps := providerSet{}
	for _, pb := range pbs {
		if pb == nil {
			continue
		}
		p := pb.Provider()
		if p == nil {
			continue
		}
		ps[p] = struct{}{}
	}

	return &module{
		subModules: ms,
		providers:  ps,
	}
}

var _ Module = &module{}

func (m *module) SubModules() ModuleIterator {
	return m.subModules
}

func (m *module) Providers() ProviderIterator {
	return m.providers
}

func (m *module) AllModules() ModuleIterator {
	ms := moduleSet{}

	var addSubModules func(sm Module)
	addSubModules = func(sm Module) {
		ms[sm] = struct{}{}
		sm.SubModules().Iterate(func(m Module) bool {
			addSubModules(m)
			return true
		})
	}
	addSubModules(m)

	return ms
}

func (m *module) AllProviders() ProviderIterator {
	ps := providerSet{}

	m.AllModules().Iterate(func(m Module) bool {
		return m.Providers().Iterate(func(p Provider) bool {
			ps[p] = struct{}{}
			return true
		})
	})

	return ps
}

func (m *module) AllComponents() ComponentCollection {
	cs := EmptyComponents()

	m.AllProviders().Iterate(func(p Provider) bool {
		cs = CombineComponents(cs, p.Components())
		return true
	})

	return cs
}

func (m *module) validateDuplicateComponentName() error {
	type dupNameKey struct {
		rType reflect.Type
		name  string
	}

	componentsWithSameName := map[dupNameKey]componentSet{}
	m.AllComponents().Each(func(c Component) {
		if c.Name() != "" {
			key := dupNameKey{c.Type(), c.Name()}
			if comSet, ok := componentsWithSameName[key]; ok {
				comSet.Add(c)
			} else {
				comSet = newComponentSet()
				comSet.Add(c)
				componentsWithSameName[key] = comSet
			}
		}
	})

	errs := errors.Empty()
	for k, v := range componentsWithSameName {
		if len(v) > 1 {
			err := errors.Empty()

			v.Each(func(com Component) {
				err = err.AddErrorf("%v at %v", com, com.Provider().Location())
			})

			err = err.WithMainf("duplicate name %q of type `%v` in components", k.name, k.rType)
			errs = errs.AddErrors(err)
		}
	}
	if errs.HasError() {
		return errs
	}

	return nil
}

func (m *module) validateProviders() error {
	providersByPackage := map[string]map[Provider]struct{}{}

	m.AllProviders().Iterate(func(p Provider) bool {
		pkg := p.Location().PkgName()

		providers := providersByPackage[pkg]
		if providers == nil {
			providers = map[Provider]struct{}{}
			providersByPackage[pkg] = providers
		}

		providers[p] = struct{}{}
		return true
	})

	errs := errors.Empty()
	for pkg, providers := range providersByPackage {
		pkgErrs := errors.Empty()
		for p := range providers {
			err := p.Validate()
			if err != nil {
				var structErr errors.StructError
				if errors.As(err, &structErr) {
					err = structErr.WithMainf("provider %+v", p)
				}

				pkgErrs = pkgErrs.AddErrors(err)
			}
		}

		if pkgErrs.HasError() {
			pkgErrs = pkgErrs.WithMainf("%v", pkg)
			errs = errs.AddErrors(pkgErrs)
		}
	}

	if errs.HasError() {
		return errs.WithMainf("there are errors at these package")
	}
	return nil
}

func (m *module) Validate() error {
	errs := errors.Empty()

	if err := m.validateDuplicateComponentName(); err != nil {
		errs = errs.AddErrors(err)
	}

	if err := m.validateProviders(); err != nil {
		errs = errs.AddErrors(err)
	}

	if errs.HasError() {
		return errs
	}

	return nil
}

func NewModule(opts ...ModuleOption) Module {
	mb := &moduleBuilder{}
	for _, o := range opts {
		if o == nil {
			continue
		}
		o.ApplyModule(mb)
	}
	return mb.Module()
}

type ModuleBuilder interface {
	AddModule(m Module) ModuleBuilder
	AddProvider(p ProviderBuilder) ModuleBuilder
	Module() Module
}

func NewModuleBuilder() ModuleBuilder {
	return &moduleBuilder{}
}

type moduleBuilder struct {
	modules          map[Module]struct{}
	providerBuilders map[ProviderBuilder]struct{}
}

func (mb *moduleBuilder) AddModule(m Module) ModuleBuilder {
	if m == nil {
		return mb
	}

	if mb.modules == nil {
		mb.modules = map[Module]struct{}{}
	}
	mb.modules[m] = struct{}{}
	return mb
}

func (mb *moduleBuilder) AddProvider(p ProviderBuilder) ModuleBuilder {
	if p == nil {
		return mb
	}

	if mb.providerBuilders == nil {
		mb.providerBuilders = map[ProviderBuilder]struct{}{}
	}
	mb.providerBuilders[p] = struct{}{}
	return mb
}

func (mb *moduleBuilder) Module() Module {
	var modules []Module
	for m := range mb.modules {
		modules = append(modules, m)
	}

	var providers []ProviderBuilder
	for pb := range mb.providerBuilders {
		providers = append(providers, pb)
	}

	return newModule(modules, providers)
}

type ModuleOption interface {
	ApplyModule(builder ModuleBuilder)
}

type moduleOption func(builder ModuleBuilder)

func (mo moduleOption) ApplyModule(builder ModuleBuilder) {
	mo(builder)
}

func SubModule(subModule Module) ModuleOption {
	return moduleOption(func(builder ModuleBuilder) {
		builder.AddModule(subModule)
	})
}

func Provide(p ProviderBuilder) ModuleOption {
	return moduleOption(func(builder ModuleBuilder) {
		builder.AddProvider(p)
	})
}
