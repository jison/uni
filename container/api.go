package container

import (
	"reflect"

	"github.com/jison/uni/module"
)

type Container interface {
	ValuesOf(criteria module.Criteria) ([]reflect.Value, error)
	Invoke(interface{})
}

type Option interface {
	Apply(*options)
}

type options struct {
	errorWhenMissingProvide bool
}
