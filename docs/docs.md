# Uni

## Usage

### add provider in modules

```go
package main

import "github.com/jison/uni"

type DBConfig struct {
	User     string
	Pass     string
	Database string
}

type DB interface {
	Query()
}

type db struct {
	config DBConfig
}

func (d *db) Query() {}

var dbModule = uni.NewModule(
    uni.Struct(&db{}, uni.As((*DB)(nil))),
)

var cfgModule = uni.NewModule(
	uni.Value(&DBConfig{
		User:     "admin",
		Pass:     "pass",
		Database: "db",
    }),
)
```

### gather all modules you need

```go
var mainModule = uni.NewModule(
	uni.Module(dbModule),
	uni.Module(cfgModule),
)
```

### init container

```go
container, err = uni.NewContainer(mainModule)
if err != nil {
	// ...
}
```

### get instances from container

```go
db, err := uni.ValueOf(container, (*DB)(nil))
if err != nil {
	// ...
}
```

### generic apis

```go
import "github.com/jison/uni/generic/uni"

m := uni.NewModule(
    uni.StructT[*db](uni.AsT[DB]()),
)

container, err = uni.NewContainer(m)
db, err := uni.ValueOfT[DB](container)
if err != nil {
	// ...
}
```

## Concepts

### Type value

There are several places where uni needs to specify the type,
here are a few ways to get the type.

#### native type

```go
uni.TypeOf(0) // int
uni.TypeOf("") // string
uni.TypeOf('a') // rune

// generic apis
uni.TypeOfT[int]() // int
uni.TypeOfT[string]() // string
uni.TypeOfT[rune]() // rune
```

#### struct

```go

type testStruct struct {}

uni.TypeOf(testStruct{}) // testStruct
uni.TypeOf(&testStruct) // *testStruct

// generic apis
uni.TypeOfT[testStruct]() // testStruct
uni.TypeOfT[*testStruct]() // *testStruct
```

#### interface

```go
type testInterface interface {}

uni.TypeOf((*testInterface)(nil)) // testInterface

// generic apis
uni.TypeOfT[testInterface]() // testInterface
```

In most cases, uni.TypeOf can be omitted, for example

```go
// The following two lines of code are equivalent
uni.As((*testInterface)(nil))
uni.As(uni.TypeOf((*testInterface)(nil)))

// generic apis
// The following two lines of code are equivalent
uni.As(uni.TypeOfT[testInterface]())
uni.AsT[testInterface]()
```

### Tag

`Tag` can be used to represent a class of components, it can be created
by `uni.NewTag`, The names are only for easy differentiation,
each tag is not equal to each other, even if their names are the same.

### Component

`Component` is the basic unit of dependency injection, a component
can be depended on by other components. They can be a concrete value,
the result of a struct constructor, or the return value of a function.

#### properties
A `Component` may contain the following properties

- **Type** This is the most important property of a component. A component
must have an explicit type and cannot be an error.

- **Name** A string that identifies the name of the component, it should
be noted that components of the same type cannot have the same name.

- **Tags** A special tag used to represent a class of components. A
component can have multiple Tags.

- **As** indicates which interfaces are implemented by the component. The
interfaces in As, like Type, collectively describe the component's type.
And noted that components can not as an error interface.

- **Hidden** Indicates whether the component is hidden or not. Hidden
components will not be matched directly by type. You need to add name
or tag along with the specified type to be matched.

- **Ignored** Indicates whether the component is ignored, and ignored
  components will no longer be injected into other components.

#### match rules

components can be matched by type in `Type` or `As`, or if you want to
match more precise components, you can specify the `Name` or `Tags` of
the component.

When a component is `Hidden`, the Name or Tag must be specified to match.

When a component is `Ignored`, no conditions will be matched, which means
that component will not be matched.

Some examples

```go
type OrderService interface {
	MakeAnOrder()
}

type orderService struct {
	
}

func (s *orderService) MakeAnOrder() {
	// to be implemented
}

tag1 := uni.NewTag("tag1")

module := uni.NewModule(
	uni.Func(
		func() *orderService, *orderService, *orderService, *orderService error {
			var s1, s2, s3, s4 *orderService
			// initialize
			return s1, s2, s3, s4, nil
        },
		uni.Return(0, uni.As((*OrderService)(nil))),
		uni.Return(1, uni.As((*OrderService)(nil)), uni.Name("OrderService")),
		uni.Return(2,
			uni.As((*OrderService)(nil)),
			uni.Name("OrderService2"),
			uni.Tags(tag1),
			uni.Hide(), // s3 is hidden
		),
		uni.Return(3, uni.Ignore()), // s4 will be ignored
	)
)

container := uni.NewContainer(module, uni.IgnoreMissing())
```

The following code matches one of s1, s2, uni will return a random one

```go
uni.ValueOf(container, (*orderService)(nil))
```

The following code will match s2

```go
uni.ValueOf(container, (*orderService)(nil), uni.ByName("OrderService"))
```

The following code will match s3

```go
uni.ValueOf(container, &orderService{},
	uni.ByTags(tag1),
	uni.ByName("OrderService2"),
)
```

### Provider

provider is used to construct Component, currently there are three
types of providers in uni

#### Value

```go

type Something interface {
	// ...
}

func somethingFactory() Something {
	// ...
	return nil
}

uni.NewModule(
	// provide component with type int and tagged with tag1
	uni.Value(10000, uni.Tags(tag1)), 

	// provide component with type func() Something
	uni.Value(somethingFactory),
	
	// provide component with type Something
	uni.Value(somethingFactory()), 
)
```

can use these options

> `Name`, `Tags`, `Scope`, `Ignore`, `Hide`, `As`

#### Struct

```go
type testStruct struct {
	a int
	b string
}

uni.NewModule(
	// provide component with type testStruct and name "abc"
	uni.Struct(testStruct{}, uni.Name("abc")),
	// provide component with type *testStruct
	uni.Struct(&testStruct{}), 
)
```

can use these options

> `Name`, `Tags`, `Scope`, `Ignore`, `Hide`, `As`, `Field`, `IgnoreFields`

#### Func

```go
uni.NewModule(
	// return value at 0, provide a component with type *something and Something.
	// return value at 1, provide a component with type Something.
	uni.Func(
		func (a int, b string) (*something, Something, error) {
			return &something{}, &something{}, nil
        },
	    uni.Return(0, uni.As((*Something)(nil))),
	),
)
```

can use these options

> `Scope`, `Param`, `Return`

### Dependency

`Dependency` is used to describe the conditions for matching `Component`s,
and the matching Components are used as input to the Provider. So far uni
has two types of Dependency, field of struct and parameter of function. 

#### field of struct

```go
type testStruct struct {
	a int // it is ok to be a Dependency even if the field is unexported.
	B string
}

uni.NewModule(
	// this struct provider have two dependencies
	// one is a component with type int
	// one is a component with type string
	uni.Struct(&testStruct{}),
)
```

We can use `uni.Field` to assign component options to specific fields

```go
type testStruct struct {
	a int
	B string
}

uni.NewModule(
	// this struct provider have two dependencies
	// one is a component with type int
	// one is a component with type string and with name "abc"
	uni.Struct(&testStruct{},
		uni.Field("B", uni.ByName("abc")),
	),
)
```

We can use `uni.IgnoreFields` to ignore some fields, the ignored fields will
not be injected, they will be filled with default values.

```go
type testStruct struct {
	a int
	B string
}

uni.NewModule(
	// this struct provider have one dependency
	// one is a component with type string
	uni.Struct(&testStruct{},
		uni.IgnoreFields(func(field reflect.StructField) bool {
			return field.Name == "a"
		}),
	),
)
```

#### parameter of function

```go
type testStruct struct {
	a int
	B string
}

uni.NewModule(
	// this function provider have two dependencies
	// one is a component with type int
	// one is a component with type string
	uni.Func(
		func(a int, b string) *testStruct {
			return &testStruct{
				a: a, B: b,	
			}
		},
	),
)
```

We can use `uni.Param` to assign component options to specific parameters

```go
type testStruct struct {
	a int
	B string
}

uni.NewModule(
	// this function provider have two dependencies
	// one is a component with type int and with name "abc"
	// one is a component with type string
	uni.Func(
		func(a int, b string) *testStruct {
			return &testStruct{
				a: a, B: b,	
			}
		},
		uni.Param(0, uni.Name("abc")),
	),
)
```

#### optional

`Dependency` can be set as optional. An optional `Dependency` will be set
to 'zero' value if there is no component match. 

```go
type testStruct struct {
	a int
	B string
}

uni.NewModule(
	// if no component with type string matches, the parameter b will be ""
	uni.Func(
		func(a int, b string) *testStruct {
			return &testStruct{
				a: a, B: b,	
			}
		},
		uni.Param(1, uni.Optional(true)),
	),
)
```

#### collector

`Dependency` with slice type can be set as collector. A collector
`Dependency` will gather all the component match the element type of
the slice and other options of `Dependency` 

```go
uni.NewModule(
	// parameter `a` match all component with type int and with name "abc"
	uni.Func(
		func(a []int) []int {
			return a	
		},
		uni.Param(0, uni.AsCollector(true), uni.ByName("abc")),
	)
)
```

> if a function is variadic, then the last parameter of this function
> will be set as collector by default.

```go
uni.NewModule(
	// parameter `a` match all component with type int and with name "abc"
	uni.Func(
		func(a ...int) []int {
			return a	
		},
		uni.Param(0, uni.ByName("abc")),
	)
)
```

### Module

we can define providers in module

```go
m1 := uni.NewModule(
	uni.Value(123),
	uni.Struct(testStruct{}),
	uni.Func(func () TestInterface { return nil }),
)
```

and we can add other module as submodules

```go
m2 := uni.NewModule(
	uni.Module(m1),
	uni.Value(456),
)
```

`Module` have a builder api to build a module

```go
mb := uni.NewModuleBuilder()
mb.AddProvider(uni.Value(123))
mb.AddProvider(uni.Struct(testStruct{}))
m1 := mb.Module()
```

> in fact, `uni.Func`, `uni.Value`, `uni.Struct` all have builder apis.

### Scope

`Scope` indicates the "available scope" of the component, and the component
can be injected only if it enters the available scope of the component.

There is a `global scope` by default, if the `Provider` does not specify
its scope, it will default to the `global scope`.

When constructing a scope, you can specify which scopes can enter the
current scope directly. If no scope is specified, it is assumed that
the scope can be entered directly from the `global scope`.

When a component matches, in addition to the rules [here](#match-rules),
the system also considers whether the scope of the component's provider
can enter (directly or indirectly) the current scope.

There are two ways to set a `Provider`'s scope, `uni.Scope` and
`uni.WithScope`.

```go
scope1 := uni.NewScope("scope1")
scope2 := uni.NewScope("scope2", scope1)

uni.NewModule(
	// provide a component with type int in scope1
	uni.Value(123, uni.Scope(scope1)),
	
	// provide a component with type string in scope2
	// and depend on a component with type int
	// because scope1 can enter scope2, so here will inject
	// the 123 int value above
	uni.Func(func (a int) string { return "" }, uni.Scope(scope2)),
	
	// provide a component with type string in global scope
	// and there is not component with type int in global scope,
	// so the a parameter can not match any components.
	uni.Func(func (a int) string { return "" }),
)
```

We can use `uni.WithScope` to set multiple `Provider`s's scope.

```go
scope1 := uni.NewScope("scope1")

uni.NewModule(
	// all these `Provider`s are in scope1
	uni.WithScope(scope1) (
		uni.Value(123),
		uni.Struct(testStruct{}),
		uni.Func(func() string { return  "" }),
	),
)
```

### Container

`Container` is a container for values of the components.

#### create

`Container`s
can be created from `Module`s, if there are errors in the module or
problems in the dependency graph, an error will be reported when 
creating it.

Multiple `Container`s can be created, although this is not necessary in
most cases, each `Container` is independent and does not affect each other.

```go
m1 := uni.NewModule(
	//...
)

c, err := uni.NewContainer(m1)
if err != nil {
	
}
```

there are many kinds of error in a dependence graph:

- some dependencies can not be fulfilled
- some dependencies can be fulfilled by more than one component
- there are cycles in dependence graph

we can choose to ignore some kind of they, with `uni.IgnoreMissing`,
`uni.IgnoreUncertain`, `uni.IgnoreCycle`.

```go
m1 := uni.NewModule(
	//...
)

// ignore the errors of 'some dependencies can not be fulfilled'
c, err := uni.NewContainer(m1, uni.IgnoreMissing())
if err != nil {
	
}
```

#### Scope

We can use `uni.EnterScope` and `uni.LeaveScope` to manage the scope of container.

```go
scope1 := uni.NewScope("scope1")

m1 := uni.NewModule(
	//...
)

c1, _ := uni.NewContainer(m1)

c2, err := uni.EnterScope(c, scope1)
// if scope1 can not enter directly from global, it is an error
if err != nil {
	// ...
}
// do something with c2

// c3 is same with c1, in global scope
c3 := uni.LeaveScope(c2)
```

#### load values

All value in container are 'lazy', they will only be instantiated when they
are needed. If you want to instantiate some values before actually using
they, you can use `Load` and `LoadAll` of `Container`.

`Load` load all the components match the criteria in current scope.

```go
m1 := uni.NewModule(
	//...
)

c, _ := uni.NewContainer(m1)
err := c.Load(
	uni.Type(0, uni.ByName("abc")),
	uni.Type(""),
	uni.Type((*TestInterface)(nil), uni.ByTags(tag1)),
)
```

`LoadAll` load all the components in current scope.

```go
m1 := uni.NewModule(
	//...
)

c, _ := uni.NewContainer(m1)
err := c.LoadAll()
```

#### consume value

We have several ways to consumer the value in the container

##### ValueOf

```go
m1 := uni.NewModule(
	//...
)

c, _ := uni.NewContainer(m1)

// val should be a instance of TestInterface, if there is a component
// with type TestInterface matched.
val, err := uni.ValueOf(c, (*TestInterface)(nil))
```

##### StructOf

```go
m1 := uni.NewModule(
	//...
)

c, _ := uni.NewContainer(m1)

// val should be a instance of testStruct, if there is a component
// with type testStruct and with name "abc" matched.
val, err := uni.StructOf(c, testStruct{}, uni.ByName("abc"))
```

##### FuncOf

```go
m1 := uni.NewModule(
	//...
)

c, _ := uni.NewContainer(m1)

// if all parameter are found, this function will be called, and
// ret will be the value of function returned
ret, err := uni.FuncOf(c, func (a int, b string) (string, error) {
	return "", nil
})
```

#### context

`Container` can be used in a golang style, which is being carried by
context. 

```go
m1 := uni.NewModule(
	//...
)

c1, _ := uni.NewContainer(m1)

ctx := uni.WithContainerCtx(context.TODO(), c1)

val, err := uni.ValueOfCtx(ctx, uni.TypeOf("")).Execute()
//...
```

## Options

- Name
- ByName
- Tags
- ByTags
- Scope
- WithScope
- Ignore
- Hide
- Optional
- As
- AsCollector
- Field
- IgnoreFields
- Param
- Return