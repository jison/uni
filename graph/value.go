package graph

import "reflect"

var nilRValue = reflect.ValueOf(nil)
var errorType = reflect.TypeOf((*error)(nil)).Elem()

type vertexValue struct {
	val reflect.Value
	err error
}

var _ VertexValue = &vertexValue{}

func (v *vertexValue) Value() reflect.Value {
	return v.val
}

func (v *vertexValue) IsNil() bool {
	if !v.val.CanInterface() {
		return false
	}

	return v.val.Interface() == nil
}

func (v *vertexValue) Error() error {
	return v.err
}

func (v *vertexValue) IsError() bool {
	return v.err != nil
}

func errorValue(err error) VertexValue {
	return &vertexValue{val: nilRValue, err: err}
}

func valueOf(val reflect.Value) VertexValue {
	return &vertexValue{val: val, err: nil}
}
