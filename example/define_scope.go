package example

import "github.com/jison/uni"

var Scope1 = uni.NewScope("Scope1")
var Scope2 = uni.NewScope("Scope2", Scope1)
var Scope3 = uni.NewScope("Scope3", Scope2)
