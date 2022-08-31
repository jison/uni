package sub_module1

import (
	"github.com/jison/uni"
	"github.com/jison/uni/example/sub_module1/sub_module2"
)

var Module1 = uni.NewModule(
	uni.Module(sub_module2.Module1),
)
