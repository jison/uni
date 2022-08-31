package model

import (
	"fmt"
	"reflect"

	"github.com/jison/uni/core/valuer"

	"github.com/jison/uni/internal/errors"
	"github.com/jison/uni/internal/reflecting"
)

type Dependency interface {
	Criteria

	Consumer() Consumer
	Optional() bool
	IsCollector() bool
	Valuer() valuer.Valuer
	Validate() error
	Equal(interface{}) bool
}

type DependencyBuilder interface {
	SetOptional(optional bool) DependencyBuilder
	SetAsCollector(asCollector bool) DependencyBuilder
	SetName(name string) DependencyBuilder
	AddTags(tags ...Symbol) DependencyBuilder
	Dependency() Dependency
}

type DependencyOption interface {
	ApplyDependency(DependencyBuilder)
}

type dependency struct {
	//provider    Provider
	consumer    Consumer
	optional    bool
	isCollector bool
	rType       reflect.Type
	val         valuer.Valuer
	name        string
	tags        *symbolSet
}

var _ Dependency = &dependency{}

func (d *dependency) Consumer() Consumer {
	return d.consumer
}

func (d *dependency) Optional() bool {
	return d.optional
}

func (d *dependency) IsCollector() bool {
	return d.isCollector
}

func (d *dependency) Type() reflect.Type {
	if d.isCollector && d.rType.Kind() == reflect.Slice {
		return d.rType.Elem()
	}
	return d.rType
}

func (d *dependency) Name() string {
	return d.name
}

func (d *dependency) Tags() SymbolSet {
	return d.tags
}

func (d *dependency) Valuer() valuer.Valuer {
	return d.val
}

func (d *dependency) Validate() error {
	errs := errors.Empty()

	if d.rType == nil {
		return errs.AddErrorf("can not inject nil type")
	}

	if reflecting.IsErrorType(d.Type()) {
		errs = errs.AddErrorf("can not inject `error` type")
	}
	if d.IsCollector() && d.rType.Kind() != reflect.Slice {
		errs = errs.AddErrorf(
			"[%v] can not marked as collector, only `Slice` type dependency can be collector",
			d.rType)
	}
	if errs.HasError() {
		return errs
	}

	return nil
}

func (d *dependency) Equal(other interface{}) bool {
	o, ok := other.(*dependency)
	if !ok {
		return false
	}
	if d == nil || o == nil {
		return d == nil && o == nil
	}

	if d.Type() != o.Type() {
		return false
	}
	if d.Name() != o.Name() {
		return false
	}
	if d.Tags() != nil {
		if !d.Tags().Equal(o.Tags()) {
			return false
		}
	} else if o.Tags() != nil {
		return false
	}

	if d.Optional() != o.Optional() {
		return false
	}
	if d.IsCollector() != o.IsCollector() {
		return false
	}
	if d.Valuer() != nil {
		if !d.Valuer().Equal(o.Valuer()) {
			return false
		}
	} else if o.Valuer() != nil {
		return false
	}

	return true
}

func (d *dependency) clone() *dependency {
	if d == nil {
		return nil
	}

	var val2 valuer.Valuer
	if d.val != nil {
		val2 = d.val.Clone()
	}

	cloned := &dependency{
		consumer:    d.consumer,
		optional:    d.optional,
		isCollector: d.isCollector,
		rType:       d.rType,
		val:         val2,
		name:        d.name,
		tags:        d.tags.clone(),
	}

	return cloned
}

func (d *dependency) Format(fs fmt.State, r rune) {
	isVerbose := fs.Flag('+') && r == 'v'

	_, _ = fmt.Fprintf(fs, "Dependency[%v]", d.Type())

	firstAttr := true
	prefix := func() {
		if firstAttr {
			_, _ = fmt.Fprint(fs, "{")
			firstAttr = false
		} else {
			_, _ = fmt.Fprint(fs, ", ")
		}
	}

	if d.name != "" {
		prefix()
		_, _ = fmt.Fprintf(fs, "name=%q", d.name)
	}
	if d.tags != nil && d.tags.Len() > 0 {
		prefix()
		if isVerbose {
			_, _ = fmt.Fprintf(fs, "tags=%+v", d.tags)
		} else {
			_, _ = fmt.Fprintf(fs, "tags=%v", d.tags)
		}
	}
	if d.optional {
		prefix()
		_, _ = fmt.Fprintf(fs, "optional")
	}
	if d.isCollector {
		prefix()
		_, _ = fmt.Fprintf(fs, "asCollector")
	}

	if !firstAttr {
		_, _ = fmt.Fprint(fs, "}")
	}
}

func (d *dependency) SetOptional(optional bool) DependencyBuilder {
	d.optional = optional
	return d
}

func (d *dependency) SetAsCollector(asCollector bool) DependencyBuilder {
	d.isCollector = asCollector
	return d
}

func (d *dependency) SetName(name string) DependencyBuilder {
	d.name = name
	return d
}

func (d *dependency) AddTags(tags ...Symbol) DependencyBuilder {
	if d.tags == nil {
		d.tags = newSymbolSet()
	}

	for _, tag := range tags {
		d.tags.Add(tag)
	}

	return d
}

func (d *dependency) Dependency() Dependency {
	return d.clone()
}

// Iterate single dependency as DependencyIterator
func (d *dependency) Iterate(f func(Dependency) bool) bool {
	return f(d)
}

func (o OptionalOption) ApplyDependency(b DependencyBuilder) {
	b.SetOptional(bool(o))
}

func (o AsCollectorOption) ApplyDependency(b DependencyBuilder) {
	b.SetAsCollector(bool(o))
}

func (o ByNameOption) ApplyDependency(b DependencyBuilder) {
	b.SetName(string(o))
}

func (o ByTagsOption) ApplyDependency(b DependencyBuilder) {
	b.AddTags(o.tags...)
}
