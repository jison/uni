package graph

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func attrsFromMustDistinct(view AttrsView) Attrs {
	attrs := Attrs{}

	if view == nil {
		return attrs
	}

	view.Iterate(func(key interface{}, value interface{}) bool {
		if _, ok := attrs[key]; ok {
			panic(fmt.Sprintf("key %v has already met", key))
		}

		attrs[key] = value
		return true
	})
	return attrs
}

type _duplicateAttrs struct {
	attrs AttrsView
}

func (d *_duplicateAttrs) Get(key interface{}) (interface{}, bool) {
	return d.attrs.Get(key)
}

func (d *_duplicateAttrs) Set(key interface{}, value interface{}) {
	d.attrs.Set(key, value)
}

func (d *_duplicateAttrs) Del(key interface{}) {
	d.attrs.Del(key)
}

func (d *_duplicateAttrs) Has(key interface{}) bool {
	return d.attrs.Has(key)
}

func (d *_duplicateAttrs) Iterate(f func(key interface{}, value interface{}) bool) bool {
	d.attrs.Iterate(f)
	return d.attrs.Iterate(f)
}

func Test_attrsFromMustDistinct(t *testing.T) {
	t.Run("all distinct attrs", func(t *testing.T) {
		attrs := Attrs{"a": "b", "c": "d"}
		assert.Equal(t, attrs, attrsFromMustDistinct(attrs))
	})

	t.Run("with duplicate attrs", func(t *testing.T) {
		assert.Panics(t, func() {
			attrs := &_duplicateAttrs{attrs: Attrs{"a": "b", "c": "d"}}
			assert.Equal(t, attrs, attrsFromMustDistinct(attrs))
		})
	})
}

func TestAttrs_Attrs(t *testing.T) {
	tests := []struct {
		name string
		a    Attrs
		want Attrs
	}{
		{"attrs is nil", nil, nil},
		{"attrs is not nil", Attrs{1: 2, 2: 3}, Attrs{1: 2, 2: 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.a, "Attrs()")
		})
	}
}

func TestAttrs_Get(t *testing.T) {
	tests := []struct {
		name  string
		attrs Attrs
		key   interface{}
		want  interface{}
		want1 bool
	}{
		{"attrs is nil", nil, 1, nil, false},
		{"key exists", Attrs{1: 2, "a": "b"}, 1, 2, true},
		{"key does not exist", Attrs{1: 2, "a": "b"}, "aa", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.attrs.Get(tt.key)
			assert.Equalf(t, tt.want, got, "Get(%v)", tt.key)
			assert.Equalf(t, tt.want1, got1, "Get(%v)", tt.key)
		})
	}
}

func TestAttrs_Has(t *testing.T) {
	tests := []struct {
		name  string
		attrs Attrs
		key   interface{}
		want  bool
	}{
		{"attrs is nil", nil, 1, false},
		{"key exists", Attrs{1: 2, "a": "b"}, 1, true},
		{"key does not exist", Attrs{1: 2, "a": "b"}, "aa", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.attrs.Has(tt.key), "Has(%v)", tt.key)
		})
	}
}

func TestAttrs_Iterate(t *testing.T) {
	tests := []struct {
		name  string
		attrs Attrs
		want  Attrs
	}{
		{"attrs is nil", nil, Attrs{}},
		{"attrs has single value", Attrs{1: 2}, Attrs{1: 2}},
		{"attrs has multiple values", Attrs{1: 2, "a": "b"}, Attrs{1: 2, "a": "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allAttr := make(Attrs)
			tt.attrs.Iterate(func(key interface{}, value interface{}) bool {
				assert.False(t, allAttr.Has(key))
				allAttr[key] = value
				assert.True(t, allAttr.Has(key))
				return true
			})
			assert.EqualValues(t, tt.want, allAttr)
		})
	}

	t.Run("interrupt iteration", func(t *testing.T) {
		attrs := Attrs{1: 2, "a": "b", 3: 4, "c": "d"}
		allAttr := make(Attrs)
		attrs.Iterate(func(key interface{}, value interface{}) bool {
			assert.False(t, allAttr.Has(key))
			allAttr[key] = value
			assert.True(t, allAttr.Has(key))
			return allAttr.Len() < 2
		})
		assert.Equal(t, 2, allAttr.Len())
	})
}

func TestAttrs_Set(t *testing.T) {
	type args struct {
		key   interface{}
		value interface{}
	}
	tests := []struct {
		name  string
		attrs Attrs
		args  args
		want  Attrs
	}{
		{"attrs is nil", nil, args{1, 2}, nil},
		{"attrs is empty", Attrs{}, args{1, 2}, Attrs{1: 2}},
		{"attrs has values", Attrs{1: 2}, args{"a", "b"}, Attrs{1: 2, "a": "b"}},
		{"same key", Attrs{1: 2}, args{1, 3}, Attrs{1: 3}},
		{"set nil key", Attrs{1: 2}, args{nil, 3}, Attrs{1: 2}},
		{"set nil value", Attrs{1: 2}, args{2, nil}, Attrs{1: 2, 2: nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.attrs.Set(tt.args.key, tt.args.value)
			assert.Equal(t, tt.want, tt.attrs)
		})
	}
}

func TestAttrs_Del(t *testing.T) {
	type args struct {
		key interface{}
	}
	tests := []struct {
		name  string
		attrs Attrs
		args  args
		want  Attrs
	}{
		{"attrs is nil", nil, args{1}, nil},
		{"attrs is empty", Attrs{}, args{1}, Attrs{}},
		{"attrs has values", Attrs{1: 2}, args{"a"}, Attrs{1: 2}},
		{"same key", Attrs{1: 2}, args{1}, Attrs{}},
		{"del nil key", Attrs{1: 2}, args{nil}, Attrs{1: 2}},
		{"del nonexistent key", Attrs{1: 2}, args{3}, Attrs{1: 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.attrs.Del(tt.args.key)
			assert.Equal(t, tt.want, tt.attrs)
		})
	}
}

func TestAttrsFrom(t *testing.T) {
	tests := []struct {
		name string
		a    AttrsView
		want Attrs
	}{
		{"attrs is nil", nil, Attrs{}},
		{"attrs is not nil", Attrs{1: 2, 2: 3}, Attrs{1: 2, 2: 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, AttrsFrom(tt.a), "Attrs()")
		})
	}
}

func Test_attrsProxy_Get(t *testing.T) {
	type want struct {
		val interface{}
		b   bool
	}

	tests := []struct {
		name  string
		attrs Attrs
		key   interface{}
		want  want
	}{
		{"get value from exist key", Attrs{"a": "b"}, "a", want{"b", true}},
		{"get value from nonexistent key", Attrs{"a": "b"}, "c", want{nil, false}},
		{"proxy nil", nil, "a", want{nil, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := attrsProxy(func(createIfNeed bool) (AttrsView, bool) {
				if tt.attrs == nil {
					return nil, false
				} else {
					return tt.attrs, true
				}
			})

			got, got1 := proxy.Get(tt.key)
			assert.Equalf(t, tt.want.val, got, "Get(%v)", tt.key)
			assert.Equalf(t, tt.want.b, got1, "Get(%v)", tt.key)
		})
	}
}

func Test_attrsProxy_Set(t *testing.T) {
	type args struct {
		key interface{}
		val interface{}
	}

	type want struct {
		val interface{}
		b   bool
	}

	tests := []struct {
		name  string
		attrs Attrs
		args  args
		want  want
	}{
		{"set value of a nonexistent key", Attrs{}, args{"a", "b"}, want{"b", true}},
		{"set value of a exist key", Attrs{"a": "c"}, args{"a", "b"}, want{"b", true}},
		{"set value from nil attrs", nil, args{"a", "b"}, want{nil, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := attrsProxy(func(createIfNeed bool) (AttrsView, bool) {
				if tt.attrs == nil {
					return nil, false
				} else {
					return tt.attrs, true
				}
			})

			proxy.Set(tt.args.key, tt.args.val)
			got, got1 := proxy.Get(tt.args.key)
			assert.Equal(t, tt.want.val, got)
			assert.Equal(t, tt.want.b, got1)
		})
	}
}

func Test_attrsProxy_Del(t *testing.T) {
	type args struct {
		key interface{}
	}

	type want struct {
		val interface{}
		b   bool
	}

	tests := []struct {
		name  string
		attrs Attrs
		args  args
		want  want
	}{
		{"del value of a nonexistent key", Attrs{}, args{"a"}, want{nil, false}},
		{"del value of a exist key", Attrs{"a": "c"}, args{"a"}, want{nil, false}},
		{"del value from nil attrs", nil, args{"a"}, want{nil, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := attrsProxy(func(createIfNeed bool) (AttrsView, bool) {
				if tt.attrs == nil {
					return nil, false
				} else {
					return tt.attrs, true
				}
			})

			proxy.Del(tt.args.key)
			got, got1 := proxy.Get(tt.args.key)
			assert.Equal(t, tt.want.val, got)
			assert.Equal(t, tt.want.b, got1)
		})
	}
}

func Test_attrsProxy_Has(t *testing.T) {
	type args struct {
		key interface{}
	}

	type want struct {
		b bool
	}

	tests := []struct {
		name  string
		attrs Attrs
		args  args
		want  want
	}{
		{"a nonexistent key", Attrs{}, args{"a"}, want{false}},
		{"a exist key", Attrs{"a": "c"}, args{"a"}, want{true}},
		{"nil attrs", nil, args{"a"}, want{false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := attrsProxy(func(createIfNeed bool) (AttrsView, bool) {
				if tt.attrs == nil {
					return nil, false
				} else {
					return tt.attrs, true
				}
			})

			got := proxy.Has(tt.args.key)
			assert.Equal(t, tt.want.b, got)
		})
	}
}

func Test_attrsProxy_Iterate(t *testing.T) {
	tests := []struct {
		name  string
		attrs Attrs
		want  Attrs
	}{
		{"empty attrs", Attrs{}, Attrs{}},
		{"no empty attrs", Attrs{"a": "b"}, Attrs{"a": "b"}},
		{"nil attrs", nil, Attrs{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := attrsProxy(func(createIfNeed bool) (AttrsView, bool) {
				if tt.attrs == nil {
					return nil, false
				} else {
					return tt.attrs, true
				}
			})

			got := attrsFromMustDistinct(proxy)
			assert.Equal(t, tt.want, got)
		})
	}
}
