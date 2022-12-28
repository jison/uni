package model

import (
	"reflect"
	"testing"

	"github.com/jison/uni/internal/location"
	"github.com/stretchr/testify/assert"
)

func TestAs(t *testing.T) {
	type args struct {
		ifs []TypeVal
	}
	tests := []struct {
		name string
		args args
		want AsOption
	}{
		{"0", args{[]TypeVal{}}, AsOption{[]TypeVal{}}},
		{"1", args{[]TypeVal{0}}, AsOption{[]TypeVal{0}}},
		{"2", args{[]TypeVal{0, ""}}, AsOption{[]TypeVal{0, ""}}},
		{"with same type", args{[]TypeVal{0, "", 0}}, AsOption{[]TypeVal{0, "", 0}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, As(tt.args.ifs...))
		})
	}
}

func TestAsCollector(t *testing.T) {
	type args struct {
		as bool
	}
	tests := []struct {
		name string
		args args
		want AsCollectorOption
	}{
		{"true", args{true}, AsCollectorOption(true)},
		{"false", args{false}, AsCollectorOption(false)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, AsCollector(tt.args.as), "AsCollector(%v)", tt.args.as)
		})
	}
}

func TestField(t *testing.T) {
	type args struct {
		name string
		opts []DependencyOption
	}
	tests := []struct {
		name string
		args args
	}{
		{"nil", args{"abc", nil}},
		{"0", args{"abc", []DependencyOption{}}},
		{"1", args{"abc", []DependencyOption{ByName("")}}},
		{"2", args{"abc", []DependencyOption{ByName(""), Optional(true)}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := Field(tt.args.name, tt.args.opts...)
			assert.Equal(t, tt.args.name, opt.name)
			assert.Equal(t, tt.args.opts, opt.opts)
		})
	}
}

func TestHide(t *testing.T) {
	tests := []struct {
		name string
		want HiddenOption
	}{
		{"true", HiddenOption{true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Hide(), "Hide()")
		})
	}
}

func TestIgnore(t *testing.T) {
	tests := []struct {
		name string
		want IgnoreOption
	}{
		{"true", IgnoreOption{true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Ignore(), "Ignore()")
		})
	}
}

func TestIgnoreFields(t *testing.T) {

	type args struct {
		p func(field reflect.StructField) bool
	}
	tests := []struct {
		name string
		args args
	}{
		{"nil", args{nil}},
		{"nil", args{func(field reflect.StructField) bool {
			return true
		}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := IgnoreFields(tt.args.p)
			assert.NotNil(t, o)
		})
	}
}

func TestLocation(t *testing.T) {
	loc1 := location.GetCallLocation(0)

	type args struct {
		loc location.Location
	}
	tests := []struct {
		name string
		args args
		want LocationOption
	}{
		{"nil", args{nil}, LocationOption{nil}},
		{"nil", args{loc1}, LocationOption{loc1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Location(tt.args.loc), "Location(%v)", tt.args.loc)
		})
	}
}

func TestUpdateCallLocation(t *testing.T) {
	baseLoc := location.GetCallLocation(0)
	var opt UpdateCallLocationOption
	func() {
		opt = UpdateCallLocation()
	}()

	assert.Equal(t, baseLoc.FileName(), opt.FileName())
	assert.Equal(t, baseLoc.FileLine()+4, opt.FileLine())
}

func TestName(t *testing.T) {
	type args struct {
		n string
	}
	tests := []struct {
		name string
		args args
		want NameOption
	}{
		{"empty", args{""}, NameOption("")},
		{"abc", args{"abc"}, NameOption("abc")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Name(tt.args.n), "Name(%v)", tt.args.n)
		})
	}
}

func TestByName(t *testing.T) {
	type args struct {
		n string
	}
	tests := []struct {
		name string
		args args
		want ByNameOption
	}{
		{"empty", args{""}, ByNameOption("")},
		{"abc", args{"abc"}, ByNameOption("abc")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ByName(tt.args.n), "Name(%v)", tt.args.n)
		})
	}
}

func TestOptional(t *testing.T) {
	type args struct {
		optional bool
	}
	tests := []struct {
		name string
		args args
		want OptionalOption
	}{
		{"true", args{true}, OptionalOption(true)},
		{"false", args{false}, OptionalOption(false)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Optional(tt.args.optional), "Optional(%v)", tt.args.optional)
		})
	}
}

func TestParam(t *testing.T) {
	type args struct {
		index int
		opts  []DependencyOption
	}
	tests := []struct {
		name string
		args args
	}{
		{"nil", args{0, nil}},
		{"0", args{1, []DependencyOption{}}},
		{"1", args{1, []DependencyOption{ByName("abc")}}},
		{"2", args{1, []DependencyOption{ByName("abc"), ByName("abc")}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := Param(tt.args.index, tt.args.opts...)
			assert.Equal(t, tt.args.index, o.index)
			assert.Equal(t, tt.args.opts, o.opts)
		})
	}
}

func TestReturn(t *testing.T) {
	type args struct {
		index int
		opts  []ComponentOption
	}
	tests := []struct {
		name string
		args args
	}{
		{"nil", args{0, nil}},
		{"0", args{1, []ComponentOption{}}},
		{"1", args{2, []ComponentOption{Name("abc")}}},
		{"2", args{3, []ComponentOption{Name("abc"), Name("abc")}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := Return(tt.args.index, tt.args.opts...)
			assert.Equal(t, tt.args.index, o.index)
			assert.Equal(t, tt.args.opts, o.opts)
		})
	}
}

func TestTags(t *testing.T) {
	tag1 := NewSymbol("tag1")
	tag2 := NewSymbol("tag2")

	type args struct {
		tags []Symbol
	}
	tests := []struct {
		name string
		args args
		want TagsOption
	}{
		{"nil", args{nil}, TagsOption{}},
		{"0", args{[]Symbol{}}, TagsOption{[]Symbol{}}},
		{"1", args{[]Symbol{tag1}}, TagsOption{[]Symbol{tag1}}},
		{"2", args{[]Symbol{tag1, tag2}}, TagsOption{[]Symbol{tag1, tag2}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Tags(tt.args.tags...))
		})
	}
}

func TestByTags(t *testing.T) {
	tag1 := NewSymbol("tag1")
	tag2 := NewSymbol("tag2")

	type args struct {
		tags []Symbol
	}
	tests := []struct {
		name string
		args args
		want ByTagsOption
	}{
		{"nil", args{nil}, ByTagsOption{}},
		{"0", args{[]Symbol{}}, ByTagsOption{[]Symbol{}}},
		{"1", args{[]Symbol{tag1}}, ByTagsOption{[]Symbol{tag1}}},
		{"2", args{[]Symbol{tag1, tag2}}, ByTagsOption{[]Symbol{tag1, tag2}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ByTags(tt.args.tags...))
		})
	}
}

func TestInScope(t *testing.T) {
	scope1 := NewScope("scope1")

	type args struct {
		scope Scope
	}
	tests := []struct {
		name string
		args args
		want ScopeOption
	}{
		{"nil", args{nil}, ScopeOption{nil}},
		{"scope", args{scope1}, ScopeOption{scope1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, InScope(tt.args.scope), "InScope(%v)", tt.args.scope)
		})
	}
}

func TestWithScope(t *testing.T) {
	//lint:ignore U1000 we need the field name to locate the field
	type testStruct struct {
		a int
	}
	scope1 := NewScope("scope1")

	m := NewModule(
		WithScope(scope1)(
			Func(func() {}),
			Struct(testStruct{}),
			Value(123),
		),
	)

	runCount := 0
	m.AllProviders().Iterate(func(p Provider) bool {
		runCount += 1
		assert.Equal(t, scope1, p.Scope())
		return true
	})
	assert.Equal(t, 3, runCount)
}
