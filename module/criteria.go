package module

import "reflect"

func NewCriteria(rtype reflect.Type, opts ...criteriaOption) Criteria {
	c := &criteria{rtype: rtype}
	for _, opt := range opts {
		opt.applyCriteria(c)
	}
	return c
}

type criteria struct {
	rtype reflect.Type
	tags  []Symbol
	name  string
}

var _ Criteria = &criteria{}

func (c *criteria) Type() reflect.Type {
	return c.rtype
}

func (c *criteria) Name() string {
	return c.name
}

func (c *criteria) Tags() []Symbol {
	return c.tags
}

type criteriaOption interface {
	applyCriteria(*criteria)
}

type nameOption string

var _ criteriaOption = nameOption("")
var _ parameterOption = nameOption("")
var _ componentOption = nameOption("")

func (o nameOption) applyCriteria(c *criteria) {
	c.name = string(o)
}

func (o nameOption) applyParameter(p *parameter) {
	p.name = string(o)
}

func (o nameOption) applyComponent(c *component) {
	c.name = string(o)
}

func Name(n string) nameOption {
	return nameOption(n)
}

type tagsOption struct {
	tags []Symbol
}

var _ criteriaOption = &tagsOption{}
var _ parameterOption = &tagsOption{}
var _ componentOption = &tagsOption{}

func (o *tagsOption) applyCriteria(c *criteria) {
	c.tags = append(c.tags, o.tags...)
}

func (o *tagsOption) applyParameter(p *parameter) {
	p.tags = append(p.tags, o.tags...)
}

func (o *tagsOption) applyComponent(c *component) {
	for _, t := range o.tags {
		c.tags[t] = struct{}{}
	}
}

func Tags(tags ...Symbol) tagsOption {
	return tagsOption{tags: tags}
}
