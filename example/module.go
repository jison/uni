package example

import (
	"strconv"

	"github.com/jison/uni"
	"github.com/jison/uni/example/sub_module1"
	"github.com/jison/uni/example/sub_module1/sub_module2"
)

type structForModuleExample struct {
	a int
	B string
}

func funcForModuleExample(a int) string {
	return strconv.Itoa(a)
}

func AddSubModule() {
	uni.NewModule(
		uni.Module(sub_module1.Module1),
		uni.Module(sub_module2.Module1),
		uni.Module(sub_module2.ModuleBuilder1.Module()),
	)
}

func AddProviders() {
	uni.NewModule(
		uni.Value(123),
		uni.Struct(structForModuleExample{}),
		uni.Func(funcForModuleExample),
	)
}

func BuildModuleWithBuilderApi() {
	var mb = uni.NewModuleBuilder()

	mb.
		AddModule(sub_module1.Module1).
		AddProvider(uni.Value(123))
}
