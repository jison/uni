package internal

type Verifiable interface {
	Validate() error
}
