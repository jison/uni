package graph

type AttrsView interface {
	Get(key interface{}) (interface{}, bool)
	Set(key interface{}, value interface{})
	Del(key interface{})
	Has(key interface{}) bool
	Iterate(func(key interface{}, value interface{}) bool) bool
}

type Attrs map[interface{}]interface{}

var _ AttrsView = Attrs{}

func (a Attrs) Get(key interface{}) (interface{}, bool) {
	v, ok := a[key]
	return v, ok
}

func (a Attrs) Set(key interface{}, value interface{}) {
	if a == nil || key == nil {
		return
	}
	a[key] = value
}

func (a Attrs) Del(key interface{}) {
	delete(a, key)
}

func (a Attrs) Has(key interface{}) bool {
	_, ok := a[key]
	return ok
}

func (a Attrs) Len() int {
	return len(a)
}

func (a Attrs) Iterate(f func(key interface{}, value interface{}) bool) bool {
	for k, v := range a {
		if !f(k, v) {
			return false
		}
	}
	return true
}

func AttrsFrom(view AttrsView) Attrs {
	attrs := Attrs{}

	if view == nil {
		return attrs
	}

	view.Iterate(func(key interface{}, value interface{}) bool {
		attrs[key] = value
		return true
	})
	return attrs
}

type attrsProxy func(createIfNeed bool) (AttrsView, bool)

func (p attrsProxy) Get(key interface{}) (interface{}, bool) {
	a, ok := p(false)
	if !ok {
		return nil, false
	}
	return a.Get(key)
}

func (p attrsProxy) Set(key interface{}, value interface{}) {
	a, ok := p(true)
	if !ok {
		return
	}
	a.Set(key, value)
}

func (p attrsProxy) Del(key interface{}) {
	a, ok := p(true)
	if !ok {
		return
	}
	a.Del(key)
}

func (p attrsProxy) Has(key interface{}) bool {
	a, ok := p(false)
	if !ok {
		return false
	}
	return a.Has(key)
}

func (p attrsProxy) Iterate(f func(key interface{}, value interface{}) bool) bool {
	a, ok := p(false)
	if !ok {
		return true
	}
	return a.Iterate(f)
}
