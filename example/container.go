package example

import "github.com/jison/uni"

var module1 = uni.NewModule()

// init
func initContainer() {
	var container uni.Container
	var err error
	container, err = uni.NewContainer(module1)
	if err != nil {
		doSomethingWithError(err)
	} else {
		doSomethingWithContainer(container)
	}
}

func doSomethingWithError(_ error) {
	// ...
}

func doSomethingWithContainer(_ uni.Container) {
	// ...
}

func doSomethingWith(_ interface{}) {
	// ...
}
