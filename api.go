package uni

import (
	"github.com/jison/uni/core"
	"github.com/jison/uni/core/model"
)

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

//goland:noinspection GoUnusedFunction
func suppressUnusedWarningDsl() {
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
	var _ = ByTags
	var _ = Tags
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
}
