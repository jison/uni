package valuer

import (
	"reflect"
	"sync"

	"github.com/jison/uni/internal/errors"

	"github.com/jison/uni/internal/reflecting"
)

type Value interface {
	AsSingle() (reflect.Value, bool)
	AsArray() ([]reflect.Value, bool)
	AsError() (error, bool)

	Initialized() bool
}

type baseValue struct{}

func (b baseValue) AsSingle() (reflect.Value, bool) {
	return reflect.Value{}, false
}

func (b baseValue) AsArray() ([]reflect.Value, bool) {
	return nil, false
}

func (b baseValue) AsError() (error, bool) {
	return nil, false
}

func (b baseValue) Initialized() bool {
	return true
}

func SingleValue(val reflect.Value) Value {
	if err, ok := reflecting.AsError(val); ok {
		return ErrorValue(err)
	} else {
		return &singleValue{val: val}
	}
}

type singleValue struct {
	baseValue
	val reflect.Value
}

var _ Value = &singleValue{}

func (s *singleValue) AsSingle() (reflect.Value, bool) {
	return s.val, true
}

func ErrorValue(err error) Value {
	if err == nil {
		return SingleValue(reflect.ValueOf(nil))
	}
	return &errorValue{err: err}
}

type errorValue struct {
	baseValue
	err error
}

var _ Value = &errorValue{}

func (e *errorValue) AsError() (error, bool) {
	return e.err, true
}

func ArrayValue(arr []reflect.Value) Value {
	return &arrayValue{arr: arr}
}

type arrayValue struct {
	baseValue
	arr []reflect.Value
}

var _ Value = &arrayValue{}

func (a *arrayValue) AsArray() ([]reflect.Value, bool) {
	return a.arr, true
}

func LazyValue(f func() Value) Value {
	return &lazyValue{f: f}
}

type lazyValue struct {
	f    func() Value
	once sync.Once
	val  Value
}

func (l *lazyValue) getValue() Value {
	l.once.Do(func() {
		if l.f != nil {
			l.val = l.f()
		} else {
			l.val = ErrorValue(errors.Newf("function is nil in lazyValue"))
		}

	})
	return l.val
}

func (l *lazyValue) AsSingle() (reflect.Value, bool) {
	return l.getValue().AsSingle()
}

func (l *lazyValue) AsArray() ([]reflect.Value, bool) {
	return l.getValue().AsArray()
}

func (l *lazyValue) AsError() (error, bool) {
	return l.getValue().AsError()
}

func (l *lazyValue) Initialized() bool {
	return l.val != nil
}

func ValuesOf(vals ...interface{}) []Value {
	var r []Value
	for _, v := range vals {
		var val Value
		rVal := reflect.ValueOf(v)
		if rVal.Kind() == reflect.Array || rVal.Kind() == reflect.Slice {
			var items []reflect.Value
			for i := 0; i < rVal.Len(); i++ {
				items = append(items, reflect.ValueOf(rVal.Index(i).Interface()))
			}
			val = ArrayValue(items)
		} else {
			val = SingleValue(rVal)
		}

		r = append(r, val)
	}
	return r
}
