package sub_module2

import (
	"strconv"

	"github.com/jison/uni"
)

var Module1 = uni.NewModule(
	uni.Value(123),
)

var ModuleBuilder1 = uni.NewModuleBuilder()

func init() {
	ModuleBuilder1.AddProvider(uni.Func(func(i int) string {
		return strconv.Itoa(i)
	}))
}
