package model

import (
	"fmt"
	"reflect"

	"github.com/jison/uni/core/valuer"

	"github.com/jison/uni/internal/errors"
	"github.com/jison/uni/internal/reflecting"
)

type Component interface {
	Provider() Provider
	Ignored() bool
	Hidden() bool
	Type() reflect.Type
	As() TypeSet
	Name() string
	Tags() SymbolSet
	Valuer() valuer.Valuer
	Validate() error
	Equal(interface{}) bool
}

type ComponentBuilder interface {
	SetIgnore(ignore bool) ComponentBuilder
	SetHidden(hidden bool) ComponentBuilder
	AddAs(ifs ...TypeVal) ComponentBuilder
	SetName(name string) ComponentBuilder
	AddTags(tags ...Symbol) ComponentBuilder
	Component() Component
}

type ComponentOption interface {
	ApplyComponent(ComponentBuilder)
}

type component struct {
	provider Provider
	ignored  bool
	hidden   bool
	rType    reflect.Type
	val      valuer.Valuer
	as       *typeSet
	name     string
	tags     *symbolSet
}

var _ Component = &component{}

func (c *component) Provider() Provider {
	return c.provider
}

func (c *component) Ignored() bool {
	return c.ignored
}

func (c *component) Hidden() bool {
	return c.hidden
}

func (c *component) Type() reflect.Type {
	return c.rType
}

func (c *component) As() TypeSet {
	return c.as
}

func (c *component) Name() string {
	return c.name
}

func (c *component) Tags() SymbolSet {
	return c.tags
}

func (c *component) Valuer() valuer.Valuer {
	return c.val
}

func (c *component) Validate() error {
	errs := errors.Empty()

	if c.rType == nil {
		errs = errs.AddErrorf("can not provide component of nil")
	}

	if reflecting.IsErrorType(c.rType) {
		errs = errs.AddErrorf("type of component can not be `error`")
	}

	c.as.Iterate(func(i reflect.Type) bool {
		if reflecting.IsErrorType(i) {
			errs = errs.AddErrorf("[%v] in `as` has implemented error, can not as error interface", i)
		} else if i.Kind() != reflect.Interface {
			errs = errs.AddErrorf("[%v] in `as` is not an interface", i)
		} else if c.rType != nil && !c.rType.Implements(i) {
			errs = errs.AddErrorf("[%v] does not implement [%v] in `as`", c.rType, i)
		}
		return true
	})

	if errs.HasError() {
		return errs
	}

	return nil
}

func (c *component) Equal(other interface{}) bool {
	o, ok := other.(*component)
	if !ok {
		return false
	}

	if c == nil || o == nil {
		return c == nil && o == nil
	}

	if c.Ignored() != o.Ignored() {
		return false
	}
	if c.Hidden() != o.Hidden() {
		return false
	}
	if c.Type() != o.Type() {
		return false
	}
	if c.As() != nil {
		if !c.As().Equal(o.As()) {
			return false
		}
	} else if o.As() != nil {
		return false
	}

	if c.Name() != o.Name() {
		return false
	}

	if c.Tags() != nil {
		if !c.Tags().Equal(o.Tags()) {
			return false
		}
	} else if o.As() != nil {
		return false
	}

	if c.Valuer() != nil {
		if !c.Valuer().Equal(o.Valuer()) {
			return false
		}
	} else if o.Valuer() != nil {
		return false
	}

	return true
}

func (c *component) clone() *component {
	if c == nil {
		return nil
	}

	var val2 valuer.Valuer
	if c.val != nil {
		val2 = c.val.Clone()
	}

	cloned := &component{
		provider: c.provider,
		ignored:  c.ignored,
		hidden:   c.hidden,
		rType:    c.rType,
		val:      val2,
		as:       c.as.clone(),
		name:     c.name,
		tags:     c.tags.clone(),
	}

	return cloned
}

func (c *component) Format(f fmt.State, r rune) {
	isVerbose := (f.Flag('+') || f.Flag('#')) && r == 'v'

	_, _ = fmt.Fprintf(f, "Component[%v]", c.rType)
	firstAttr := true
	prefix := func() {
		if firstAttr {
			_, _ = fmt.Fprint(f, "{")
			firstAttr = false
		} else {
			_, _ = fmt.Fprint(f, ", ")
		}
	}

	if c.name != "" {
		prefix()
		_, _ = fmt.Fprintf(f, "name=%q", c.name)
	}
	if c.tags != nil && c.tags.Len() > 0 {
		prefix()
		if isVerbose {
			_, _ = fmt.Fprintf(f, "tags=%+v", c.tags)
		} else {
			_, _ = fmt.Fprintf(f, "tags=%v", c.tags)
		}
	}
	if c.as != nil && c.as.Len() > 0 {
		prefix()
		_, _ = fmt.Fprintf(f, "as=%v", c.as)
	}
	if c.ignored {
		prefix()
		_, _ = fmt.Fprintf(f, "ignored")
	}
	if c.hidden {
		prefix()
		_, _ = fmt.Fprintf(f, "hidden")
	}

	if !firstAttr {
		_, _ = fmt.Fprint(f, "}")
	}
}

func (c *component) SetIgnore(ignore bool) ComponentBuilder {
	c.ignored = ignore
	return c
}

func (c *component) SetHidden(hidden bool) ComponentBuilder {
	c.hidden = hidden
	return c
}

func (c *component) AddAs(ifs ...TypeVal) ComponentBuilder {
	if c.as == nil {
		c.as = newTypeSet()
	}
	for _, i := range ifs {
		c.as.Add(TypeOf(i))
	}
	return c
}

func (c *component) SetName(name string) ComponentBuilder {
	c.name = name
	return c
}

func (c *component) AddTags(tags ...Symbol) ComponentBuilder {
	if c.tags == nil {
		c.tags = newSymbolSet()
	}

	for _, tag := range tags {
		c.tags.Add(tag)
	}
	return c
}

func (c *component) Component() Component {
	return c.clone()
}

// Iterate single component as a component iterator
func (c *component) Iterate(f func(c Component) bool) bool { return f(c) }

func (o IgnoreOption) ApplyComponent(b ComponentBuilder) {
	b.SetIgnore(o.ignore)
}

func (o HiddenOption) ApplyComponent(b ComponentBuilder) {
	b.SetHidden(o.hidden)
}

func (o AsOption) ApplyComponent(b ComponentBuilder) {
	b.AddAs(o.ifs...)
}

func (o NameOption) ApplyComponent(b ComponentBuilder) {
	b.SetName(string(o))
}

func (o TagsOption) ApplyComponent(b ComponentBuilder) {
	if len(o.tags) == 0 {
		return
	}
	b.AddTags(o.tags...)
}
