package model

import (
	"fmt"
	"reflect"
	"sort"
)

type TypeVal interface{}

func TypeOf(v TypeVal) reflect.Type {
	var t reflect.Type

	switch vt := v.(type) {
	case reflect.Type:
		t = vt
	case reflect.Value:
		t = vt.Type()
	default:
		t = reflect.TypeOf(v)
	}

	if t != nil && t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Interface {
		t = t.Elem()
	}

	return t
}

type TypeSet interface {
	Has(t reflect.Type) bool
	Len() int
	Iterate(func(p reflect.Type) bool) bool
	Equal(other TypeSet) bool
}

func newTypeSet(ts ...reflect.Type) *typeSet {
	set := &typeSet{}
	for _, t := range ts {
		set.Add(t)
	}

	return set
}

type typeSet struct {
	indexByType map[reflect.Type]int
	lastIndex   int
}

func (ts *typeSet) Has(t reflect.Type) bool {
	if ts == nil {
		return false
	}

	_, ok := ts.indexByType[t]
	return ok
}

func (ts *typeSet) Len() int {
	if ts == nil {
		return 0
	}

	return len(ts.indexByType)
}

func (ts *typeSet) Iterate(f func(p reflect.Type) bool) bool {
	if ts == nil {
		return true
	}

	for t := range ts.indexByType {
		if !f(t) {
			return false
		}
	}
	return true
}

func (ts *typeSet) Equal(other TypeSet) bool {
	if ts.Len() != other.Len() {
		return false
	}

	return other.Iterate(func(t reflect.Type) bool {
		return ts.Has(t)
	})
}

func (ts *typeSet) Add(t reflect.Type) {
	if ts == nil || t == nil {
		return
	}

	if ts.indexByType == nil {
		ts.indexByType = map[reflect.Type]int{}
	}
	if _, ok := ts.indexByType[t]; ok {
		return
	}

	ts.indexByType[t] = ts.lastIndex
	ts.lastIndex += 1
}

func (ts *typeSet) Del(t reflect.Type) {
	if ts == nil {
		return
	}

	delete(ts.indexByType, t)
}

func (ts *typeSet) types() []reflect.Type {
	if ts == nil {
		return nil
	}

	var types []reflect.Type
	for t := range ts.indexByType {
		types = append(types, t)
	}

	sort.Slice(types, func(i, j int) bool {
		return ts.indexByType[types[i]] < ts.indexByType[types[j]]
	})

	return types
}

func (ts *typeSet) clone() *typeSet {
	if ts == nil {
		return nil
	}

	set := &typeSet{
		indexByType: map[reflect.Type]int{},
		lastIndex:   ts.lastIndex,
	}

	for t, i := range ts.indexByType {
		set.indexByType[t] = i
	}

	return set
}

func (ts *typeSet) Format(fs fmt.State, r rune) {
	var typeFormat string
	if fs.Flag('+') && r == 'v' {
		typeFormat = "%+v"
	} else {
		typeFormat = "%v"
	}

	_, _ = fmt.Fprint(fs, "{")
	first := true
	for _, t := range ts.types() {
		if !first {
			_, _ = fmt.Fprintf(fs, ", ")
		} else {
			first = false
		}

		_, _ = fmt.Fprintf(fs, typeFormat, t)
	}
	_, _ = fmt.Fprint(fs, "}")
}
