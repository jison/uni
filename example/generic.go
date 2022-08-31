package example

import "github.com/jison/uni/generic/uni"

func generic() {
	uni.NewModule(uni.Func(func() int {
		return 0
	}, uni.Return(0, uni.AsT[any]())))
}
