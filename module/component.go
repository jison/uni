package module

import (
	"reflect"

	"github.com/jison/uni/internal/errors"
)

type component struct {
	provider *provider
	ignored  bool
	hidden   bool
	index    int
	rtype    reflect.Type
	as       map[reflect.Type]struct{}
	name     string
	tags     map[Symbol]struct{}
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

func (c *component) Index() int {
	return c.index
}

func (c *component) Type() reflect.Type {
	return c.rtype
}

func (c *component) As() []reflect.Type {
	r := make([]reflect.Type, 0, len(c.as))
	for i := range c.as {
		r = append(r, i)
	}

	return r
}

func (c *component) AllTypes() []reflect.Type {
	r := make([]reflect.Type, 0, len(c.as)+1)
	r = append(r, c.rtype)
	for i := range c.as {
		r = append(r, i)
	}

	return r
}

func (c *component) Name() string {
	return c.name
}

func (c *component) Tags() []Symbol {
	r := make([]Symbol, 0, len(c.tags))
	for t := range c.tags {
		r = append(r, t)
	}

	return r
}

func (c *component) HasTag(t Symbol) bool {
	_, ok := c.tags[t]
	return ok
}

func (c *component) Match(cri Criteria) bool {
	if _, ok := c.as[cri.Type()]; !ok && cri.Type() != c.rtype {
		return false
	}

	if cri.Name() != "" {
		if c.name != cri.Name() {
			return false
		}
	}

	if len(cri.Tags()) > 0 {
		for _, tag := range cri.Tags() {
			if _, ok := c.tags[tag]; !ok {
				return false
			}
		}
	}

	if c.Hidden() && cri.Name() == "" && len(cri.Tags()) == 0 {
		return false
	}

	return true
}

func (c *component) Validate() error {
	errs := make([]error, 0)
	for i := range c.as {
		if i.Kind() != reflect.Interface {
			errs = append(errs, errors.New("%s in `as` must be interface", i))
		}
		if !c.rtype.Implements(i) {
			errs = append(errs, errors.New("%s does not implement %s in `as`", c.rtype, i))
		}
	}

	if len(errs) > 0 {
		return errors.Merge(errs...)
	}

	return nil
}

type componentOption interface {
	applyComponent(*component)
}

type ignoreOption struct{}

func (o ignoreOption) applyComponent(c *component) {
	c.ignored = true
}

func Ignore() ignoreOption {
	return ignoreOption{}
}

type hideOption struct{}

func (o hideOption) applyComponent(c *component) {
	c.hidden = true
}

func Hide() hideOption {
	return hideOption{}
}

type asOption struct {
	interfaces []reflect.Type
}

func (o *asOption) applyComponent(c *component) {
	for _, i := range o.interfaces {
		c.as[i] = struct{}{}
	}
}

func As(ifs ...interface{}) *asOption {
	arr := make([]reflect.Type, 0)
	for _, i := range ifs {
		switch t := i.(type) {
		case reflect.Type:
			arr = append(arr, t)
		case reflect.Value:
			if t.Kind() == reflect.Ptr {
				rtype := t.Type().Elem()
				arr = append(arr, rtype)
			}
		default:
			rtype := reflect.TypeOf(i)
			if rtype.Kind() == reflect.Ptr {
				rtype = rtype.Elem()
				arr = append(arr, rtype)
			}
		}
	}

	return &asOption{interfaces: arr}
}
