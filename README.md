# Uni

> a dependency injection library for golang

**WARNING: This project has not been fully validated in a production
environment, so please use carefully.**

## Key Features

### Non-intrusive

uni provide a clean api to describe dependence, 
you dont need to mix uni code in your implements.

### Modular

uni support describe dependence in a modular way.

### Life cycle management

uni use a scope base model to manage component life cycle,
component can only construct in the specify scope, and after
leaving the scope, all components in the scope will be released.

### Errors aware

there are many kinds of error in a dependence graph:

- some dependencies can not be fulfilled
- some dependencies can be fulfilled by more than one component
- there are cycles in dependence graph

uni can discover all these errors in the beginning stage,
and returns structured error messages to help identify problems

### Generic apis

uni provides generic apis,
and you still can use uni in go1.17

### Concurrency safe

uni guarantees that each provider under the same container
can only be executed in one goroutine at the same time

## Installation

```shell
go get github.com/jison/uni 
```

## Usage

you can find the sample codes in `example` folder

## Documentation

see [docs/docs.md](docs/docs.md)

## Todo

- [ ] add command tools to improve debug experience
- [ ] add support for recovering of function provider panic
- [ ] TBD: add decorator support for components
- [ ] TBD: add mock api for components 
