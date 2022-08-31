package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func comOf(v interface{}, opts ...ValueProviderOption) Component {
	p := Value(v, opts...).Provider()
	return p.Components().ToArray()[0]
}

func criOf(t TypeVal, opts ...CriteriaOption) Criteria {
	return NewCriteria(t, opts...).Criteria()
}

func Test_componentRepository_AllComponents(t *testing.T) {
	tag1 := NewSymbol("tag1")
	tag2 := NewSymbol("tag2")

	tests := []struct {
		name string
		coms componentSet
	}{
		{"0", newComponentSet()},
		{"1", newComponentSet(
			comOf(0, Name("abc")),
		)},
		{"n", newComponentSet(
			comOf(0, Name("abc")), comOf(""),
			comOf("", Tags(tag1)),
			comOf("", Tags(tag2)),
			comOf("", Tags(tag2)),
		)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewRepository(tt.coms)
			got := m.AllComponents().ToSet()
			assert.Equal(t, tt.coms, got)
		})
	}
}

func Test_componentRepository_ComponentsMatch(t *testing.T) {
	scope1 := NewScope("scope1")
	scope2 := NewScope("scope2", scope1)
	scope3 := NewScope("scope3")

	tag1 := NewSymbol("tag1")
	tag2 := NewSymbol("tag2")
	tag3 := NewSymbol("tag3")

	type testStruct struct {
		a int
		b string
	}

	type testInterface interface{}

	coms := []Component{
		comOf(1),                                 // 0
		comOf(""),                                // 1
		comOf(testStruct{}),                      // 2
		comOf(&testStruct{}),                     // 3
		comOf(1, Name("name1"), InScope(scope1)), // 4
		comOf(1, Ignore()),                       // 5
		comOf(1, Ignore(), InScope(scope2)),      // 6
		comOf(1, Name("name2"), Hide()),          // 7
		comOf(1, Name("name2"), Tags(tag3), InScope(scope3)),                       // 8
		comOf(1, Name("name2"), Tags(tag3), InScope(scope3), Hide()),               // 9
		comOf("", Name("name2"), InScope(scope2)),                                  // 10
		comOf("", Tags(tag1), Hide()),                                              // 11
		comOf("", Name("name3"), Ignore()),                                         // 12
		comOf("", Hide()),                                                          // 13
		comOf("", Name("name3"), Tags(tag2, tag3)),                                 // 14
		comOf(1, As((*testInterface)(nil)), Name("name3"), Tags(tag1), Hide()),     // 15
		comOf("", As((*testInterface)(nil)), Name("name3"), Tags(tag2)),            // 16
		comOf(testStruct{}, As((*testInterface)(nil)), Name("name3"), Tags(tag3)),  // 17
		comOf(&testStruct{}, As((*testInterface)(nil)), Name("name3"), Tags(tag3)), // 18
	}

	type args struct {
		coms []Component
		cri  Criteria
	}
	tests := []struct {
		name        string
		args        args
		wantIndexes []int
	}{
		{"0", args{coms, criOf(1)}, []int{0, 4, 8}},
		{"1", args{coms, criOf(1, ByName("name1"))}, []int{4}},
		{"2", args{coms, criOf(1, ByName("name2"))}, []int{7, 8, 9}},
		{"3", args{coms, criOf(1, ByName("name3"))}, []int{15}},
		{"4", args{coms, criOf(1, ByTags(tag2))}, []int{}},
		{"5", args{coms, criOf(1, ByTags(tag3))}, []int{8, 9}},
		{"6", args{coms, criOf(1, ByName("name3"), ByTags(tag1))}, []int{15}},
		{"7", args{coms, criOf(1, ByName("name3"), ByTags(tag3))}, []int{}},
		{"8", args{coms, criOf("", ByName("name3"), ByTags(tag3))}, []int{14}},
		{"9", args{coms, criOf("", ByName("name3"), ByTags(tag2))}, []int{14, 16}},
		{"10", args{coms, criOf("", ByTags(tag2, tag3))}, []int{14}},
		{"11", args{coms, criOf("")}, []int{1, 10, 14, 16}},
		{"12", args{coms, criOf(testStruct{})}, []int{2, 17}},
		{"13", args{coms, criOf(testStruct{}, ByName("name3"))}, []int{17}},
		{"14", args{coms, criOf(testStruct{}, ByName("name4"))}, []int{}},
		{"15", args{coms, criOf(&testStruct{})}, []int{3, 18}},
		{"16", args{coms, criOf(&testStruct{}, ByName("name3"))}, []int{18}},
		{"17", args{coms, criOf((*testInterface)(nil))}, []int{16, 17, 18}},
		{"18", args{coms, criOf((*testInterface)(nil), ByTags(tag1))}, []int{15}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := newComponentSet(tt.args.coms...)
			m := NewRepository(set)

			var wantComs []Component
			for _, idx := range tt.wantIndexes {
				wantComs = append(wantComs, tt.args.coms[idx])
			}
			got := m.ComponentsMatch(tt.args.cri)
			assert.Equal(t, newComponentSet(wantComs...), newComponentSet(got.ToArray()...))
		})
	}
}

func Test_componentRepository_ComponentsWithScope(t *testing.T) {
	scope1 := NewScope("scope1")
	scope2 := NewScope("scope2", scope1)
	scope3 := NewScope("scope3")

	tag1 := NewSymbol("tag1")
	tag2 := NewSymbol("tag2")
	tag3 := NewSymbol("tag3")

	type testStruct struct {
		a int
		b string
	}

	type testInterface interface{}

	coms := []Component{
		comOf(1),                                 // 0
		comOf(""),                                // 1
		comOf(testStruct{}),                      // 2
		comOf(&testStruct{}),                     // 3
		comOf(1, Name("name1"), InScope(scope1)), // 4
		comOf(1, Ignore()),                       // 5
		comOf(1, Ignore(), InScope(scope2)),      // 6
		comOf(1, Name("name2"), Hide()),          // 7
		comOf(1, Name("name2"), Tags(tag3), InScope(scope3)),                             // 8
		comOf(1, Name("name2"), Tags(tag3), InScope(scope3), Hide()),                     // 9
		comOf("", Name("name2"), InScope(scope2)),                                        // 10
		comOf("", Tags(tag1), Hide()),                                                    // 11
		comOf("", Name("name3"), Ignore()),                                               // 12
		comOf("", Hide()),                                                                // 13
		comOf("", Name("name3"), Tags(tag2, tag3)),                                       // 14
		comOf(1, As((*testInterface)(nil)), Name("name3"), Tags(tag1), Hide()),           // 15
		comOf("", As((*testInterface)(nil)), Name("name3"), Tags(tag2), InScope(scope1)), // 16
		comOf(testStruct{}, As((*testInterface)(nil)), Name("name3"), Tags(tag3)),        // 17
		comOf(&testStruct{}, As((*testInterface)(nil)), Name("name3"), Tags(tag3)),       // 18
	}

	type args struct {
		coms  []Component
		scope Scope
	}
	tests := []struct {
		name        string
		args        args
		wantIndexes []int
	}{
		{"nil", args{coms, nil}, []int{0, 1, 2, 3, 7, 11, 13, 14, 15, 17, 18}},
		{"scope1", args{coms, scope1}, []int{4, 16}},
		{"scope2", args{coms, scope2}, []int{10}},
		{"scope3", args{coms, scope3}, []int{8, 9}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := newComponentSet(tt.args.coms...)
			m := NewRepository(set)

			var wantComs []Component
			for _, idx := range tt.wantIndexes {
				wantComs = append(wantComs, tt.args.coms[idx])
			}
			got := m.ComponentsWithScope(tt.args.scope)
			assert.Equal(t, newComponentSet(wantComs...), newComponentSet(got.ToArray()...))
		})
	}
}

func Test_componentRepository_ComponentsMatchDependency(t *testing.T) {
	scope1 := NewScope("scope1")
	scope2 := NewScope("scope2", scope1)
	scope3 := NewScope("scope3")

	type testStruct struct {
		a int
		b string
	}

	type testInterface interface{}

	m := NewModule(
		Func(func(ts testStruct) testStruct {
			return ts
		}, InScope(scope2), Return(0, Name("name1"))),
		Struct(testStruct{}, InScope(scope1), Name("name2"), As((*testInterface)(nil))),
		Struct(testStruct{}, InScope(scope2), Name("name3"), As((*testInterface)(nil))),
		Value(123, Name("name4")),
		Value(456, InScope(scope1), Name("name5")),
		Value(789, InScope(scope3), Name("name6")),
		Value("abc", Name("name7")),
		Value("def", Name("name8")),
	)
	rep := NewRepository(m.AllComponents())

	t.Run("scope", func(t *testing.T) {
		com := m.AllComponents().Filter(func(com Component) bool {
			return com.Name() == "name3"
		}).ToArray()[0]

		var dep Dependency
		com.Provider().Dependencies().Iterate(func(d Dependency) bool {
			if d.Type() == TypeOf(1) {
				dep = d
				return false
			}
			return true
		})
		coms := rep.ComponentsMatchDependency(dep).ToArray()
		assert.Equal(t, 2, len(coms))
		for _, com := range coms {
			assert.Contains(t, []string{"name4", "name5"}, com.Name())
		}
	})

	t.Run("match all within scope", func(t *testing.T) {
		dep := dependencyIteratorToArray(LoadAllConsumer(scope2).Consumer().Dependencies())[0]
		coms := rep.ComponentsMatchDependency(dep).ToArray()
		assert.Equal(t, 2, len(coms))
		for _, com := range coms {
			assert.Contains(t, []string{"name1", "name3"}, com.Name())
		}
	})

	t.Run("match all within global", func(t *testing.T) {
		dep := dependencyIteratorToArray(LoadAllConsumer(GlobalScope).Consumer().Dependencies())[0]
		coms := rep.ComponentsMatchDependency(dep).ToArray()
		assert.Equal(t, 3, len(coms))
		for _, com := range coms {
			assert.Contains(t, []string{"name4", "name7", "name8"}, com.Name())
		}
	})

	t.Run("match all can enter scope", func(t *testing.T) {
		com := m.AllComponents().Filter(func(com Component) bool {
			return com.Name() == "name3"
		}).ToArray()[0]

		var dep Dependency
		com.Provider().Dependencies().Iterate(func(d Dependency) bool {
			if d.Type() == TypeOf(1) {
				dep = d
				return false
			}
			return true
		})

		coms := rep.ComponentsMatchDependency(dep).ToArray()
		assert.Equal(t, 2, len(coms))
		for _, com := range coms {
			assert.Contains(t, []string{"name4", "name5"}, com.Name())
		}
	})

	t.Run("match collector", func(t *testing.T) {
		consumer := ValueConsumer(TypeOf([]testInterface{}), AsCollector(true), InScope(scope2)).Consumer()
		var dep Dependency
		consumer.Dependencies().Iterate(func(d Dependency) bool {
			dep = d
			return false
		})

		coms := rep.ComponentsMatchDependency(dep).ToArray()
		assert.Equal(t, 2, len(coms))
		for _, com := range coms {
			assert.Contains(t, []string{"name2", "name3"}, com.Name())
		}
	})

	t.Run("can not match components with provider same with dependency's consumer", func(t *testing.T) {
		com := m.AllComponents().Filter(func(com Component) bool {
			return com.Name() == "name1"
		}).ToArray()[0]

		dep := dependencyIteratorToArray(com.Provider().Dependencies())[0]
		coms := rep.ComponentsMatchDependency(dep).ToArray()
		assert.Equal(t, 2, len(coms))
		for _, com := range coms {
			assert.Contains(t, []string{"name2", "name3"}, com.Name())
		}

	})
}

func Test_componentMatch(t *testing.T) {
	tag1 := NewSymbol("tag1")
	tag2 := NewSymbol("tag2")
	tag3 := NewSymbol("tag3")

	type args struct {
		com Component
		cri Criteria
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"component is nil", args{nil, NewCriteria(1).Criteria()}, false},
		{"criteria is nil", args{comOf(1), nil}, false},
		{"type mismatch", args{comOf(1), criOf("")}, false},
		{"type match", args{comOf(1, Name("abc")), criOf(1)}, true},
		{"type mismatch and name match", args{
			comOf(1, Name("abc"), Tags(tag1, tag2)),
			criOf("", ByName("abc")),
		}, false},
		{"type match and name match", args{
			comOf(0, Name("abc"), Tags(tag1, tag2)),
			criOf(0, ByName("abc")),
		}, true},
		{"type match and name mismatch", args{
			comOf(0, Name("abc"), Tags(tag1, tag2)),
			criOf(0, ByName("abc2")),
		}, false},
		{"type match and all tag match", args{
			comOf(0, Name("abc"), Tags(tag1, tag2, tag3)),
			criOf(0, ByTags(tag1, tag2)),
		}, true},
		{"type match and one tag mismatch", args{
			comOf(0, Name("abc"), Tags(tag1, tag2)),
			criOf(0, ByTags(tag1, tag3)),
		}, false},
		{"type match and name match and tag match", args{
			comOf(0, Name("abc"), Tags(tag1, tag2)),
			criOf(0, ByName("abc"), ByTags(tag1)),
		}, true},
		{"type match and name match and tag match", args{
			comOf(0, Name("abc"), Tags(tag1, tag2)),
			criOf(0, ByName("abc"), ByTags(tag1)),
		}, true},
		{"type mismatch and as type match", args{
			comOf("", Name("abc"), Tags(tag1, tag2), As(0, "")),
			criOf(0),
		}, true},
		{"type mismatch and as type mismatch", args{
			comOf("", Name("abc"), Tags(tag1, tag2), As('a', "")),
			criOf(0),
		}, false},
		{"type mismatch and as type match and name match", args{
			comOf("", Name("abc"), Tags(tag1, tag2), As(0, "")),
			criOf(0, ByName("abc")),
		}, true},
		{"type mismatch and as type match and name mismatch", args{
			comOf("", Name("abc"), Tags(tag1, tag2), As(0, "")),
			criOf(0, ByName("abc2")),
		}, false},
		{"type mismatch and as type match and tag mismatch", args{
			comOf("", Name("abc"), Tags(tag1, tag2), As(0, "")),
			criOf(0, ByTags(tag3)),
		}, false},
		{"type mismatch and as type match and tag match", args{
			comOf("", Name("abc"), Tags(tag1, tag2), As(0, "")),
			criOf(0, ByTags(tag1)),
		}, true},
		{"type mismatch and as type match and name and tag match", args{
			comOf("", Name("abc"), Tags(tag1, tag2), As(0, "")),
			criOf(0, ByName("abc"), ByTags(tag1)),
		}, true},
		{"type mismatch and as type match and name and tag mismatch", args{
			comOf("", Name("abc"), Tags(tag1, tag2), As(0, "")),
			criOf(0, ByName("abc"), ByTags(tag3)),
		}, false},
		{"type match and hidden", args{
			comOf("", Name("abc"), Tags(tag1, tag2), Hide()),
			criOf(0),
		}, false},
		{"type mismatch and as type match and hidden", args{
			comOf("", Name("abc"), Tags(tag1, tag2), Hide(), As(0, "")),
			criOf(0),
		}, false},
		{"type mismatch and as type match and name match hidden", args{
			comOf("", Name("abc"), Tags(tag1, tag2), Hide(), As(0, "")),
			criOf(0, ByName("abc")),
		}, true},
		{"type mismatch and as type match and tag match hidden", args{
			comOf("", Name("abc"), Tags(tag1, tag2), Hide(), As(0, "")),
			criOf(0, ByTags(tag1)),
		}, true},
		{"type mismatch and as type match and name mismatch and tag match hidden", args{
			comOf("", Name("abc"), Tags(tag1, tag2), Hide(), As(0, "")),
			criOf(0, ByName("abc2"), ByTags(tag1)),
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := componentMatch(tt.args.com, tt.args.cri); got != tt.want {
				t.Errorf("componentMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
