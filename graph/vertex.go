package graph

import (
	"reflect"

	"github.com/jison/uni/internal/errors"
)

type funcVertex struct {
	funcVal     reflect.Value
	resultIndex int
}

func NewFuncVertex(funcVal reflect.Value, index int) Vertex {
	return &funcVertex{funcVal: funcVal, resultIndex: index}
}

func (v *funcVertex) Provide(varr []VertexValue) VertexValue {
	errs := make([]error, 0)
	vals := make([]reflect.Value, 0, len(varr))
	for _, v := range varr {
		if v.IsError() {
			errs = append(errs, v.Error())
		} else {
			vals = append(vals, v.Value())
		}
	}
	if len(errs) > 0 {
		return errorValue(errors.Merge(errs...))
	}

	rval, err := callFunc(v.funcVal, vals)
	if err != nil {
		return errorValue(err)
	}

	if v.resultIndex < 0 {
		return valueOf(reflect.ValueOf(rval))
	} else if v.resultIndex >= len(rval) {
		return errorValue(errors.Bugf("resultIndex %d is out of range, len of result is %d", v.resultIndex, len(rval)))
	} else {
		return valueOf(rval[v.resultIndex])
	}
}

func callFunc(funcVal reflect.Value, params []reflect.Value) ([]reflect.Value, error) {
	funcType := funcVal.Type()
	if funcType.Kind() != reflect.Func {
		return nil, errors.Bugf("funcVal is not a function")
	}
	if funcType.NumIn() != len(params) {
		return nil, errors.Bugf("function need %d parameters, but provided %d", funcType.NumIn(), len(params))
	}

	rvals := funcVal.Call(params)

	errs := make([]error, 0)
	for i := 0; i < funcType.NumOut(); i++ {
		outType := funcType.Out(i)
		if outType.Implements(errorType) {
			errVal := rvals[i]
			var err error
			if !errVal.CanInterface() {
				err = errors.New("error can not convert to interface{}")
			} else {
				err = errVal.Interface().(error)
			}
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return nil, errors.Merge(errs...)
	}

	return rvals, nil
}

type indexedVertex struct {
	index int
}

func NewIndexedVertex(index int) Vertex {
	return &indexedVertex{index}
}

func (v *indexedVertex) Provide(varr []VertexValue) VertexValue {
	if len(varr) == 0 {
		return errorValue(errors.New("varr is empty"))
	}

	val := varr[0]
	if val.IsError() {
		return val
	}

	rval := val.Value()

	if rval.Kind() != reflect.Slice && rval.Kind() != reflect.Array {
		return errorValue(errors.New("varr[0] is not an array"))
	}

	if rval.Len() <= v.index {
		return errorValue(errors.New("varr[0] len is %d, and index %d is out of range", rval.Len(), v.index))
	}

	return valueOf(rval.Index(v.index))
}

type valueVertex struct {
	val reflect.Value
}

func NewValueVertex(val reflect.Value) Vertex {
	return &valueVertex{val: val}
}

func (v *valueVertex) Provide(vl []VertexValue) VertexValue {
	return valueOf(v.val)
}

type errorVertex struct {
	err error
}

func NewErrorVertex(err error) Vertex {
	return &errorVertex{err: err}
}

func (v *errorVertex) Provide(vl []VertexValue) VertexValue {
	return errorValue(v.err)
}

type oneOfVertex struct {
}

func NewOneOfVertex() Vertex {
	return &oneOfVertex{}
}

func (v *oneOfVertex) Provide(varr []VertexValue) VertexValue {
	for _, v := range varr {
		if !v.IsError() {
			return v
		}
	}
	return errorValue(errors.New("all value of varr are errors"))
}

type arrayVertex struct{}

func NewArrayVertex() Vertex {
	return &arrayVertex{}
}

func (v *arrayVertex) Provide(varr []VertexValue) VertexValue {
	errs := make([]error, 0)
	rarr := make([]interface{}, 0, len(varr))
	for _, v := range varr {
		if v.IsError() {
			errs = append(errs, v.Error())
		}
		if !v.Value().CanInterface() {
			errs = append(errs, errors.New("v.CanInterface() is false"))
		}
		rarr = append(rarr, v.Value().Interface())
	}
	if len(errs) > 0 {
		return errorValue(errors.Merge(errs...))
	}

	return valueOf(reflect.ValueOf(rarr))
}
