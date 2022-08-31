package model

import (
	"fmt"
	"reflect"

	"github.com/jison/uni/core/valuer"
	"github.com/jison/uni/internal/errors"
	"github.com/jison/uni/internal/location"
	"github.com/jison/uni/internal/reflecting"
)

type structField struct {
	*dependency
	ignored bool
	field   reflect.StructField
}

var _ Dependency = &structField{}

func (sf *structField) Format(f fmt.State, r rune) {
	_, _ = fmt.Fprintf(f, "%v at field `%s`", sf.dependency, sf.field.Name)

	if f.Flag('+') && r == 'v' {
		_, _ = fmt.Fprintf(f, " of %+v", sf.Consumer())
	}
}

func (sf *structField) Equal(other interface{}) bool {
	o, ok := other.(*structField)
	if !ok {
		return false
	}
	if sf == nil || o == nil {
		return sf == nil && o == nil
	}

	if sf.ignored != o.ignored {
		return false
	}

	if !reflect.DeepEqual(sf.field, o.field) {
		return false
	}

	if sf.dependency != nil {
		if !sf.dependency.Equal(o.dependency) {
			return false
		}
	} else if o.dependency != nil {
		return false
	}

	return true
}

func (sf *structField) clone() *structField {
	return &structField{
		dependency: sf.dependency.clone(),
		ignored:    sf.ignored,
		field:      sf.field,
	}
}

type fieldByName map[string]*structField

func (m fieldByName) Iterate(f func(Dependency) bool) bool {
	for _, field := range m {
		if field.ignored {
			continue
		}
		if !f(field) {
			return false
		}
	}
	return true
}

type structConsumer struct {
	*baseConsumer
	sType      reflect.Type
	fields     fieldByName
	fakeFields fieldByName
}

var _ Consumer = &structConsumer{}

func (sc *structConsumer) Dependencies() DependencyIterator {
	return sc.fields
}

func (sc *structConsumer) Validate() error {
	errs := errors.Empty()

	if !reflecting.IsKindOrPtrOfKind(sc.sType, reflect.Struct) {
		return errors.Newf("[%v] is not type of struct or type of pointer of struct", sc.sType)
	}
	if reflecting.IsErrorType(sc.sType) {
		errs = errs.AddErrorf("[%v] is implementing error, can not provide error value", sc.sType)
	}

	for fieldName := range sc.fakeFields {
		errs = errs.AddErrorf("field `%v` is nonexistent", fieldName)
	}

	for _, field := range sc.fields {
		if field.ignored {
			continue
		}
		err := field.dependency.Validate()
		if err != nil {
			var structErr errors.StructError
			if errors.As(err, &structErr) {
				err = structErr.WithMainf("field `%v`", field.name)
			}

			errs = errs.AddErrors(err)
		}
	}

	if errs.HasError() {
		return errs
	}

	return nil
}

func (sc *structConsumer) Format(f fmt.State, r rune) {
	if f.Flag('+') && r == 'v' {
		_, _ = fmt.Fprintf(f, "StructConsumer[%v] at %v", sc.sType, sc.Location())
	} else {
		_, _ = fmt.Fprintf(f, "StructConsumer[%v]", sc.sType)
	}
}

func (sc *structConsumer) clone() *structConsumer {
	cloned := &structConsumer{
		baseConsumer: sc.baseConsumer.clone(),
		sType:        sc.sType,
		fields:       fieldByName{},
		fakeFields:   fieldByName{},
	}

	cloneFields := func(oldFields, newFields fieldByName) {
		for name, field := range oldFields {
			f := field.clone()
			f.dependency.consumer = cloned
			newFields[name] = f
		}
	}

	cloneFields(sc.fields, cloned.fields)
	cloneFields(sc.fakeFields, cloned.fakeFields)

	return cloned
}

func (sc *structConsumer) Equal(other interface{}) bool {
	o, ok := other.(*structConsumer)
	if !ok {
		return false
	}

	if sc.sType != o.sType {
		return false
	}

	if len(sc.fields) != len(o.fields) {
		return false
	}

	for name, f := range sc.fields {
		if f2, fok := o.fields[name]; !fok {
			return false
		} else if !f.Equal(f2) {
			return false
		}
	}

	if len(sc.fakeFields) != len(o.fakeFields) {
		return false
	}

	for name, f := range sc.fakeFields {
		if f2, fok := o.fakeFields[name]; !fok {
			return false
		} else if !f.Equal(f2) {
			return false
		}
	}

	if sc.baseConsumer != nil {
		if !sc.baseConsumer.Equal(o.baseConsumer) {
			return false
		}
	} else if o.baseConsumer != nil {
		return false
	}

	return true
}

type StructConsumerBuilder interface {
	Field(fieldName string, opts ...DependencyOption) StructConsumerBuilder
	IgnoreFields(predicate func(field reflect.StructField) bool) StructConsumerBuilder
	SetScope(scope Scope) StructConsumerBuilder
	SetLocation(loc location.Location) StructConsumerBuilder
	UpdateCallLocation(loc location.Location) StructConsumerBuilder
	Consumer() Consumer
}

func (sc *structConsumer) Field(fieldName string, opts ...DependencyOption) StructConsumerBuilder {
	var field *structField
	var ok bool
	if field, ok = sc.fields[fieldName]; !ok {
		if field, ok = sc.fakeFields[fieldName]; !ok {
			field = &structField{
				dependency: &dependency{
					consumer: sc,
					rType:    reflect.TypeOf(nil),
					val:      valuer.Field(fieldName),
				},
				field: reflect.StructField{},
			}
			sc.fakeFields[fieldName] = field
		}
	}

	for _, o := range opts {
		if o == nil {
			continue
		}
		o.ApplyDependency(field)
	}

	return sc
}

func (sc *structConsumer) IgnoreFields(predicate func(field reflect.StructField) bool) StructConsumerBuilder {
	if predicate == nil {
		return sc
	}

	for _, field := range sc.fields {
		if predicate(field.field) {
			field.ignored = true
		}
	}
	return sc
}

func (sc *structConsumer) SetScope(scope Scope) StructConsumerBuilder {
	sc.baseConsumer.SetScope(scope)
	return sc
}

func (sc *structConsumer) SetLocation(loc location.Location) StructConsumerBuilder {
	sc.baseConsumer.SetLocation(loc)
	return sc
}

func (sc *structConsumer) UpdateCallLocation(loc location.Location) StructConsumerBuilder {
	if sc.Location() == nil {
		if loc == nil {
			loc = location.GetCallLocation(3).Callee()
		}
		sc.SetLocation(loc)
	}
	return sc
}

func (sc *structConsumer) Consumer() Consumer {
	return sc.clone()
}

func StructConsumer(t TypeVal, opts ...StructConsumerOption) StructConsumerBuilder {
	sc := structConsumerOf(TypeOf(t), opts...).UpdateCallLocation(nil)
	return sc
}

func structConsumerOf(t reflect.Type, opts ...StructConsumerOption) *structConsumer {
	sc := &structConsumer{
		baseConsumer: &baseConsumer{
			val: valuer.Struct(t),
		},
		sType:      t,
		fields:     fieldByName{},
		fakeFields: fieldByName{},
	}

	if t != nil {
		structType := t
		if structType.Kind() == reflect.Ptr {
			structType = structType.Elem()
		}
		if structType.Kind() == reflect.Struct {
			for i := 0; i < structType.NumField(); i++ {
				field := structType.Field(i)
				sc.fields[field.Name] = &structField{
					dependency: &dependency{
						consumer: sc,
						val:      valuer.Field(field.Name),
						rType:    field.Type,
					},
					field: field,
				}
			}
		}
	}

	for _, o := range opts {
		if o == nil {
			continue
		}
		o.ApplyStructConsumer(sc)
	}

	return sc
}

type StructConsumerOption interface {
	ApplyStructConsumer(StructConsumerBuilder)
}

func (o FieldOption) ApplyStructConsumer(b StructConsumerBuilder) {
	b.Field(o.name, o.opts...)
}

func (o IgnoreFieldsOption) ApplyStructConsumer(b StructConsumerBuilder) {
	b.IgnoreFields(o.predicate)
}

func (o ScopeOption) ApplyStructConsumer(b StructConsumerBuilder) {
	b.SetScope(o.scope)
}

func (o LocationOption) ApplyStructConsumer(b StructConsumerBuilder) {
	b.SetLocation(o.Location)
}

func (o UpdateCallLocationOption) ApplyStructConsumer(b StructConsumerBuilder) {
	b.UpdateCallLocation(o.Location)
}
