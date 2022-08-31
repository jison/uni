package example

import "github.com/jison/uni"

func test() {
	uni.Func(func(a int) string { return "" },
		uni.Param(0, uni.ByName("a")),
		uni.Return(0, uni.Name("r")),
	)
}
