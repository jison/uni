package model

import (
	"fmt"
	"reflect"

	"github.com/jison/uni/core/valuer"
	"github.com/jison/uni/internal/errors"
	"github.com/jison/uni/internal/location"
	"github.com/jison/uni/internal/reflecting"
)

type criteriaAsDependency struct {
	Criteria
	consumer Consumer
	val      valuer.Valuer
}

func (c *criteriaAsDependency) Consumer() Consumer {
	return c.consumer
}

func (c *criteriaAsDependency) Optional() bool {
	return false
}

func (c *criteriaAsDependency) IsCollector() bool {
	return false
}

func (c *criteriaAsDependency) Valuer() valuer.Valuer {
	return c.val
}

func (c *criteriaAsDependency) Validate() error {
	if reflecting.IsErrorType(c.Type()) {
		return errors.Newf("can not inject with `error` type")
	}

	return nil
}

func (c *criteriaAsDependency) clone() *criteriaAsDependency {
	if c == nil {
		return nil
	}

	return &criteriaAsDependency{
		Criteria: c.Criteria,
		consumer: c.consumer,
		val:      c.val.Clone(),
	}
}

func (c *criteriaAsDependency) Equal(other interface{}) bool {
	o, ok := other.(*criteriaAsDependency)
	if !ok {
		return false
	}

	if o == nil || c == nil {
		return o == nil && c == nil
	}

	if c.Criteria == nil || o.Criteria == nil {
		return c.Criteria == nil && o.Criteria == nil
	}
	if !c.Criteria.Equal(o.Criteria) {
		return false
	}

	if c.val == nil || o.val == nil {
		return c.val == nil && o.val == nil
	}
	if !c.val.Equal(o.val) {
		return false
	}

	return true
}

func (c *criteriaAsDependency) Format(fs fmt.State, r rune) {
	if fs.Flag('+') && r == 'v' {
		_, _ = fmt.Fprintf(fs, "%+v", c.Criteria)
	} else {
		_, _ = fmt.Fprintf(fs, "%v", c.Criteria)
	}
}

type criteriaAsDepList []*criteriaAsDependency

func (dl criteriaAsDepList) Iterate(f func(Dependency) bool) bool {
	for _, dep := range dl {
		if !f(dep) {
			return false
		}
	}
	return true
}

func (dl criteriaAsDepList) Format(fs fmt.State, r rune) {
	formatDependencyIterator(dl, fs, r)
}

type loadCriteriaConsumer struct {
	*baseConsumer
	dependencies criteriaAsDepList
}

var _ Consumer = &loadCriteriaConsumer{}

func (c *loadCriteriaConsumer) Dependencies() DependencyIterator {
	return c.dependencies
}

func (c *loadCriteriaConsumer) Validate() error {
	errs := errors.Empty()

	for _, dep := range c.dependencies {
		if err := dep.Validate(); err != nil {
			errs = errs.AddErrors(err)
		}
	}

	if errs.HasError() {
		return errs
	}

	return nil
}

func (c *loadCriteriaConsumer) Format(f fmt.State, r rune) {
	_, _ = fmt.Fprintf(f, "Load %v", c.dependencies)

	if f.Flag('+') && r == 'v' {
		_, _ = fmt.Fprintf(f, " at %+v", c.Location())
	}
}

func (c *loadCriteriaConsumer) clone() *loadCriteriaConsumer {
	if c == nil {
		return nil
	}

	cloned := &loadCriteriaConsumer{
		baseConsumer: c.baseConsumer.clone(),
	}

	for _, dep := range c.dependencies {
		clonedDep := dep.clone()
		clonedDep.consumer = cloned

		cloned.dependencies = append(cloned.dependencies, clonedDep)
	}

	return cloned
}

func (c *loadCriteriaConsumer) Equal(other interface{}) bool {
	o, ok := other.(*loadCriteriaConsumer)
	if !ok {
		return false
	}

	if c == nil || o == nil {
		return c == nil && o == nil
	}

	if len(c.dependencies) != len(o.dependencies) {
		return false
	}

	for i, d := range c.dependencies {
		d2 := o.dependencies[i]
		if !d.Equal(d2) {
			return false
		}
	}

	if c.baseConsumer != nil {
		if !c.baseConsumer.Equal(o.baseConsumer) {
			return false
		}
	} else if o.baseConsumer != nil {
		return false
	}

	return true
}

type LoadCriteriaConsumerBuilder interface {
	AddCriteria(cri CriteriaBuilder) LoadCriteriaConsumerBuilder
	SetScope(scope Scope) LoadCriteriaConsumerBuilder
	SetLocation(loc location.Location) LoadCriteriaConsumerBuilder
	UpdateCallLocation(loc location.Location) LoadCriteriaConsumerBuilder
	Consumer() Consumer
}

func (c *loadCriteriaConsumer) AddCriteria(cri CriteriaBuilder) LoadCriteriaConsumerBuilder {
	dep := &criteriaAsDependency{
		Criteria: cri.Criteria(),
		consumer: c,
		val:      valuer.Identity(),
	}
	c.dependencies = append(c.dependencies, dep)

	return c
}

func (c *loadCriteriaConsumer) SetScope(scope Scope) LoadCriteriaConsumerBuilder {
	c.baseConsumer.SetScope(scope)
	return c
}

func (c *loadCriteriaConsumer) SetLocation(loc location.Location) LoadCriteriaConsumerBuilder {
	c.baseConsumer.SetLocation(loc)
	return c
}

func (c *loadCriteriaConsumer) UpdateCallLocation(loc location.Location) LoadCriteriaConsumerBuilder {
	if c.Location() == nil {
		if loc == nil {
			loc = location.GetCallLocation(3).Callee()
		}
		c.SetLocation(loc)
	}
	return c
}

func (c *loadCriteriaConsumer) Consumer() Consumer {
	return c.clone()
}

func loadCriteriaConsumerOf(criteriaList ...CriteriaBuilder) *loadCriteriaConsumer {
	c := &loadCriteriaConsumer{
		baseConsumer: &baseConsumer{
			val: valuer.Collector(TypeOf((*interface{})(nil))),
		},
	}

	for _, cri := range criteriaList {
		if cri == nil {
			continue
		}
		c.AddCriteria(cri)
	}
	return c
}

func LoadCriteriaConsumer(criteriaList ...CriteriaBuilder) LoadCriteriaConsumerBuilder {
	return loadCriteriaConsumerOf(criteriaList...).UpdateCallLocation(nil)
}

var wildcardType = reflecting.AnyType

func IsWildCardType(t reflect.Type) bool {
	return t == wildcardType
}

type loadAllConsumer struct {
	*baseConsumer
	dep *dependency
}

func (l *loadAllConsumer) Dependencies() DependencyIterator {
	return l.dep
}

func (l *loadAllConsumer) Validate() error {
	return nil
}

func (l *loadAllConsumer) Format(f fmt.State, r rune) {
	_, _ = fmt.Fprintf(f, "LoadAll")

	if f.Flag('+') && r == 'v' {
		_, _ = fmt.Fprintf(f, " at %+v", l.Location())
	}
}

func (l *loadAllConsumer) clone() *loadAllConsumer {
	if l == nil {
		return nil
	}

	cloned := &loadAllConsumer{
		baseConsumer: l.baseConsumer.clone(),
		dep:          l.dep.clone(),
	}

	cloned.dep.consumer = cloned
	return cloned
}

func (l *loadAllConsumer) Equal(other interface{}) bool {
	o, ok := other.(*loadAllConsumer)
	if !ok {
		return false
	}

	if l == nil || o == nil {
		return l == nil && o == nil
	}

	if l.baseConsumer != nil {
		if !l.baseConsumer.Equal(o.baseConsumer) {
			return false
		}
	} else if o.baseConsumer != nil {
		return false
	}

	if l.dep != nil {
		if !l.dep.Equal(o.dep) {
			return false
		}
	} else if o.dep != nil {
		return false
	}

	return true
}

type LoadAllConsumerBuilder interface {
	SetScope(scope Scope) LoadAllConsumerBuilder
	SetLocation(loc location.Location) LoadAllConsumerBuilder
	UpdateCallLocation(loc location.Location) LoadAllConsumerBuilder
	Consumer() Consumer
}

func (l *loadAllConsumer) SetScope(scope Scope) LoadAllConsumerBuilder {
	l.baseConsumer.SetScope(scope)
	return l
}

func (l *loadAllConsumer) SetLocation(loc location.Location) LoadAllConsumerBuilder {
	l.baseConsumer.SetLocation(loc)
	return l
}

func (l *loadAllConsumer) UpdateCallLocation(loc location.Location) LoadAllConsumerBuilder {
	if l.Location() == nil {
		if loc == nil {
			loc = location.GetCallLocation(3).Callee()
		}
		l.SetLocation(loc)
	}
	return l
}

func (l *loadAllConsumer) Consumer() Consumer {
	return l.clone()
}

func loadAllConsumerOf(scope Scope) *loadAllConsumer {
	c := &loadAllConsumer{
		baseConsumer: &baseConsumer{
			val: valuer.Identity(),
		},
		dep: &dependency{
			rType:       reflect.SliceOf(wildcardType),
			val:         valuer.Identity(),
			isCollector: true,
		},
	}
	c.dep.consumer = c
	if scope == nil {
		c.SetScope(GlobalScope)
	} else {
		c.SetScope(scope)
	}
	return c
}

func LoadAllConsumer(scope Scope) LoadAllConsumerBuilder {
	c := loadAllConsumerOf(scope).UpdateCallLocation(nil)
	return c
}
