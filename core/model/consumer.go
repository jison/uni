package model

import (
	"github.com/jison/uni/core/valuer"
	"github.com/jison/uni/internal/location"
)

type Consumer interface {
	Dependencies() DependencyIterator
	Valuer() valuer.Valuer
	Scope() Scope
	Location() location.Location
	Validate() error
	Equal(interface{}) bool
}

type ConsumerBuilder interface {
	Consumer() Consumer
}

type baseConsumer struct {
	scope Scope
	loc   location.Location
	val   valuer.Valuer
}

func (c *baseConsumer) Valuer() valuer.Valuer {
	return c.val
}

func (c *baseConsumer) Scope() Scope {
	if c.scope == nil {
		return GlobalScope
	}
	return c.scope
}

func (c *baseConsumer) SetScope(s Scope) {
	c.scope = s
}

func (c *baseConsumer) Location() location.Location {
	return c.loc
}

func (c *baseConsumer) SetLocation(loc location.Location) {
	c.loc = loc
}

func (c *baseConsumer) clone() *baseConsumer {
	var val2 valuer.Valuer
	if c.val != nil {
		val2 = c.val.Clone()
	}

	cloned := &baseConsumer{
		scope: c.scope,
		loc:   c.loc,
		val:   val2,
	}
	return cloned
}

func (c *baseConsumer) Equal(other interface{}) bool {
	o, ok := other.(*baseConsumer)
	if !ok {
		return false
	}

	if c == nil || o == nil {
		return c == nil && o == nil
	}

	if c.Location() != o.Location() {
		return false
	}

	if c.Scope() != o.Scope() {
		return false
	}

	if c.Valuer() == nil || o.Valuer() == nil {
		return c.Valuer() == nil && o.Valuer() == nil
	}
	if !c.Valuer().Equal(o.Valuer()) {
		return false
	}

	return true
}
