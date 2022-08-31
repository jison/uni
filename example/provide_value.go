package example

import (
	"fmt"
	"reflect"

	"github.com/jison/uni"
)

type structForValueExample struct {
	a string
	B int
}

func (s *structForValueExample) String() string {
	return fmt.Sprintf("structForValueExample{a: %q, B: %d}", s.a, s.B)
}

func ProvideValue() {
	uni.NewModule(
		uni.Value(123),
		uni.Value("abc"),
		uni.Value([2]int{123, 456}),
		uni.Value(map[int]string{123: "abc", 456: "def"}),
		uni.Value(structForValueExample{a: "abc", B: 123}),
		uni.Value(&structForValueExample{a: "abc", B: 123}),
		uni.Value(func() string { return "abc" }),
	)
}

func ProvideValueWithName() {
	uni.NewModule(
		uni.Value(123, uni.Name("max_connections")),
	)
}

func ProvideValueWithTags() {
	uni.NewModule(
		uni.Value(123, uni.Tags(Tag1)),
		uni.Value("abc", uni.Tags(Tag1, Tag2)),
		uni.Value(func() string { return "yes!" }, uni.Tags(Tag1, Tag2), uni.Tags(Tag3)),
	)
}

func ProvideValueAsInterface() {
	uni.NewModule(
		uni.Value(
			&structForValueExample{a: "abc", B: 123},
			uni.As((fmt.Stringer)(nil)),
		),
		uni.Value(
			&structForValueExample{a: "abc", B: 123},
			uni.As(reflect.TypeOf((fmt.Stringer)(nil)).Elem()),
		),
		uni.Value(
			&structForValueExample{a: "abc", B: 123},
			uni.As(uni.Type((fmt.Stringer)(nil))),
		),
	)
}

func IgnoreValueProvider() {
	uni.NewModule(
		uni.Value(&structForValueExample{a: "abc", B: 123}, uni.Ignore()),
	)
}

func HideValueProvider() {
	uni.NewModule(
		uni.Value(&structForValueExample{a: "abc", B: 123}, uni.Hide()),
	)
}

func ProvideValueWithBuilderApi() {
	uni.NewModule(
		uni.Value(&structForValueExample{a: "abc", B: 123}).
			SetName("test_struct").
			AddTags(Tag1).
			AddAs((*fmt.Stringer)(nil)).
			SetIgnore(true).
			SetScope(Scope3),
	)
}

func ProvideValueWithMixApi() {
	uni.NewModule(
		uni.Value(123, uni.Name("name")).AddTags(Tag1),
		uni.Value("abc", uni.Tags(Tag2)).SetScope(Scope3).AddTags(Tag3),
	)
}
