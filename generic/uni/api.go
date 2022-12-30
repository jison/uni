package uni

import (
	"context"

	"github.com/jison/uni/core"
	"github.com/jison/uni/core/model"
	"github.com/jison/uni/internal/errors"
)

import "reflect"

type Container = core.Container

var IgnoreMissing = core.IgnoreMissing
var IgnoreUncertain = core.IgnoreUncertain
var IgnoreCycle = core.IgnoreCycle
var NewContainer = core.NewContainer

var NewModuleBuilder = model.NewModuleBuilder
var NewModule = model.NewModule

var Module = model.SubModule
var Provide = model.Provide

var Value = model.Value

var Struct = model.Struct
var Field = model.Field
var IgnoreFields = model.IgnoreFields

var Func = model.Func
var Param = model.Param
var Return = model.Return

var AsCollector = model.AsCollector
var Optional = model.Optional

var As = model.As
var Ignore = model.Ignore
var Hide = model.Hide

var Scope = model.InScope
var WithScope = model.WithScope

var TypeOf = model.TypeOf
var Type = model.NewCriteria

var Name = model.Name
var ByName = model.ByName
var Tags = model.Tags
var ByTags = model.ByTags

var NewTag = model.NewSymbol

var NewScope = model.NewScope

var BuildFunc = model.FuncConsumer
var BuildStruct = model.StructConsumer
var BuildValue = model.ValueConsumer

var FuncOf = core.FuncOf
var StructOf = core.StructOf
var ValueOf = core.ValueOf

var FuncOfCtx = core.FuncOfCtx
var StructOfCtx = core.StructOfCtx
var ValueOfCtx = core.ValueOfCtx
var EnterScopeCtx = core.EnterScopeCtx
var LeaveScopeCtx = core.LeaveScopeCtx

func TypeOfT[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func TypeT[P any](opts ...model.CriteriaOption) model.CriteriaBuilder {
	return Type(TypeOfT[P](), opts...)
}

func AsT[T any]() model.ComponentOption {
	return As(TypeOfT[T]())
}

func StructT[T any](opts ...model.StructProviderOption) model.StructProviderBuilder {
	opts = append(opts, model.UpdateCallLocation())
	return Struct(TypeOfT[T](), opts...)
}

func convertTo[T any](val any, err error) (T, error) {
	var t T
	var ok bool
	if err != nil {
		return t, err
	}
	t, ok = val.(T)
	if !ok {
		return t, errors.Newf("can not convert %v as type %v", val, TypeOfT[T]())
	}

	return t, nil
}

func FuncOfT(c Container, f any, opts ...model.FuncConsumerOption) ([]any, error) {
	opts = append(opts, model.UpdateCallLocation())
	val, err := core.FuncOf(c, f, opts...)
	return convertTo[[]any](val, err)
}

func StructOfT[T any](c Container, opts ...model.StructConsumerOption) (T, error) {
	opts = append(opts, model.UpdateCallLocation())
	val, err := StructOf(c, TypeOfT[T](), opts...)
	return convertTo[T](val, err)
}

func ValueOfT[T any](c Container, opts ...model.ValueConsumerOption) (T, error) {
	opts = append(opts, model.UpdateCallLocation())
	val, err := ValueOf(c, TypeOfT[T](), opts...)
	return convertTo[T](val, err)
}

func FuncOfCtxT(ctx context.Context, f any, opts ...model.FuncConsumerOption) ([]any, error) {
	opts = append(opts, model.UpdateCallLocation())
	val, err := core.FuncOfCtx(ctx, f, opts...)
	return convertTo[[]any](val, err)
}

func StructOfCtxT[T any](ctx context.Context, opts ...model.StructConsumerOption) (T, error) {
	opts = append(opts, model.UpdateCallLocation())
	val, err := StructOfCtx(ctx, TypeOfT[T](), opts...)
	return convertTo[T](val, err)
}

func ValueOfCtxT[T any](ctx context.Context, opts ...model.ValueConsumerOption) (T, error) {
	opts = append(opts, model.UpdateCallLocation())
	val, err := ValueOfCtx(ctx, TypeOfT[T](), opts...)
	return convertTo[T](val, err)
}

//goland:noinspection GoUnusedFunction
func suppressUnusedWarningDslGeneric() {
	var _ = IgnoreMissing
	var _ = IgnoreUncertain
	var _ = IgnoreCycle
	var _ = NewContainer
	var _ = NewModuleBuilder
	var _ = NewModule
	var _ = Module
	var _ = Provide
	var _ = Struct
	var _ = Value
	var _ = Field
	var _ = IgnoreFields
	var _ = Func
	var _ = Param
	var _ = Return
	var _ = AsCollector
	var _ = Optional
	var _ = As
	var _ = Ignore
	var _ = Hide
	var _ = Scope
	var _ = WithScope
	var _ = TypeOf
	var _ = Type
	var _ = Name
	var _ = ByName
	var _ = Tags
	var _ = ByTags
	var _ = NewTag
	var _ = NewScope
	var _ = BuildFunc
	var _ = BuildStruct
	var _ = BuildValue
	var _ = FuncOf
	var _ = StructOf
	var _ = ValueOf
	var _ = FuncOfCtx
	var _ = StructOfCtx
	var _ = ValueOfCtx
	var _ = EnterScopeCtx
	var _ = LeaveScopeCtx
	var _ = TypeOfT[any]
	var _ = TypeT[any]
	var _ = AsT[any]
	var _ = StructT[any]
	var _ = FuncOfT
	var _ = StructOfT[any]
	var _ = ValueOfT[any]
	var _ = FuncOfCtxT
	var _ = StructOfCtxT[any]
	var _ = ValueOfCtxT[any]
}
