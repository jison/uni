package module

import (
	"reflect"

	"github.com/jison/uni/internal/errors"
)

func providerOfError(err error) *provider {
	return &provider{
		funcVal:       reflect.ValueOf(err),
		parameterList: nil,
		componentList: nil,
	}
}

func providerOfFunc(v reflect.Value) *provider {
	funcType := v.Type()
	if funcType.Kind() != reflect.Func {
		return providerOfError(errors.New("param is not a function"))
	}

	pl := buildFuncParamList(funcType)
	cl := buildFuncComponentList(funcType)

	p := &provider{
		funcVal:       v,
		parameterList: pl,
		componentList: cl,
	}
	for _, param := range pl {
		param.provider = p
	}
	for _, c := range cl {
		c.provider = p
	}

	return p
}

func buildFuncParamList(rtype reflect.Type) parameterList {
	numIn := rtype.NumIn()

	pl := make([]*parameter, numIn)
	for i := 0; i < numIn; i++ {
		p := &parameter{}
		p.index = i
		p.rtype = rtype.In(i)

		pl[i] = p
	}

	return pl
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

func isErrorType(rtype reflect.Type) bool {
	return rtype.Implements(errorType)
}

func buildFuncComponentList(rtype reflect.Type) map[int]*component {
	numOut := rtype.NumOut()

	cm := make(map[int]*component, 0)
	for i := 0; i < numOut; i++ {
		cType := rtype.Out(i)
		if isErrorType(cType) {
			continue
		}
		c := &component{}
		c.index = i
		c.rtype = cType
		cm[i] = c
	}

	return cm
}

func providerOfStruct(s reflect.Value) *provider {
	rtype := s.Type()
	if rtype.Kind() != reflect.Struct {
		return providerOfError(errors.New("param is not a struct"))
	}

	pl := buildStructFieldList(rtype)
	c := &component{
		index: 0,
		rtype: rtype,
	}

	funcVal := reflect.ValueOf(func(params ...interface{}) (interface{}, error) {
		if len(params) != rtype.NumField() {
			return nil, errors.Bugf("len of param does not match the struct needs")
		}

		val := reflect.New(rtype).Elem()
		for i := 0; i < rtype.NumField(); i++ {
			val.Field(i).Set(reflect.ValueOf(params[i]))
		}

		return val.Interface(), nil
	})

	p := &provider{
		funcVal:       funcVal,
		parameterList: pl,
		componentList: map[int]*component{0: c},
	}

	for _, param := range pl {
		param.provider = p
	}
	c.provider = p

	return p
}

func buildStructFieldList(rtype reflect.Type) parameterList {
	numField := rtype.NumField()

	pl := make([]*parameter, numField)
	for i := 0; i < numField; i++ {
		p := &parameter{}
		p.index = i
		p.rtype = rtype.Field(i).Type

		pl[i] = p
	}

	return pl
}

func providerOfValue(v reflect.Value) *provider {
	rtype := v.Type()

	if isErrorType(rtype) {
		return providerOfError(errors.New("can not provide error value"))
	}

	c := &component{
		index: 0,
		rtype: rtype,
	}

	funcVal := reflect.ValueOf(func() (interface{}, error) {
		return v.Interface(), nil
	})

	p := &provider{
		funcVal:       funcVal,
		parameterList: nil,
		componentList: map[int]*component{0: c},
	}

	c.provider = p

	return p
}
