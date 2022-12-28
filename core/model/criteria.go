package model

import (
	"fmt"
	"reflect"
)

type Criteria interface {
	Type() reflect.Type
	Name() string
	Tags() SymbolSet
	Equal(interface{}) bool
}

type CriteriaBuilder interface {
	SetName(string) CriteriaBuilder
	AddTags(...Symbol) CriteriaBuilder
	Criteria() Criteria
}

type CriteriaOption interface {
	ApplyCriteria(CriteriaBuilder)
}

func NewCriteria(t TypeVal, opts ...CriteriaOption) CriteriaBuilder {
	c := &criteria{rType: TypeOf(t)}
	for _, opt := range opts {
		opt.ApplyCriteria(c)
	}
	return c
}

type criteria struct {
	rType reflect.Type
	tags  *symbolSet
	name  string
}

var _ Criteria = &criteria{}

func (c *criteria) Type() reflect.Type {
	return c.rType
}

func (c *criteria) Name() string {
	return c.name
}

func (c *criteria) Tags() SymbolSet {
	return c.tags
}

func (c *criteria) Equal(other interface{}) bool {
	o, ok := other.(*criteria)
	if !ok {
		return false
	}
	if c == nil || o == nil {
		return c == nil && o == nil
	}

	if c.Type() != o.Type() {
		return false
	}
	if c.Name() != o.Name() {
		return false
	}
	if !c.Tags().Equal(o.Tags()) {
		return false
	}

	return true
}

func (c *criteria) Format(fs fmt.State, r rune) {
	isVerbose := fs.Flag('+') && r == 'v'

	_, _ = fmt.Fprint(fs, "{")
	_, _ = fmt.Fprintf(fs, "type=%v", c.rType)

	if c.name != "" {
		_, _ = fmt.Fprintf(fs, ", name=%q", c.name)
	}
	if c.tags != nil && c.tags.Len() > 0 {
		if isVerbose {
			_, _ = fmt.Fprintf(fs, ", tags=%+v", c.tags)
		} else {
			_, _ = fmt.Fprintf(fs, ", tags=%v", c.tags)
		}
	}
	_, _ = fmt.Fprint(fs, "}")
}

func (c *criteria) clone() *criteria {
	if c == nil {
		return nil
	}

	c2 := &criteria{
		rType: c.rType,
		tags:  c.tags.clone(),
		name:  c.name,
	}

	return c2
}

func (c *criteria) SetName(name string) CriteriaBuilder {
	c.name = name
	return c
}

func (c *criteria) AddTags(tags ...Symbol) CriteriaBuilder {
	if c.tags == nil {
		c.tags = newSymbolSet()
	}

	for _, t := range tags {
		c.tags.Add(t)
	}
	return c
}

func (c *criteria) Criteria() Criteria {
	return c.clone()
}

func (o ByNameOption) ApplyCriteria(cb CriteriaBuilder) {
	cb.SetName(string(o))
}

func (o ByTagsOption) ApplyCriteria(cb CriteriaBuilder) {
	cb.AddTags(o.tags...)
}
