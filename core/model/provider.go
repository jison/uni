package model

type Provider interface {
	Consumer

	Components() ComponentCollection
	Scope() Scope
	Validate() error
}

type baseProvider struct {
	//baseConsumer
}

type ProviderBuilder interface {
	Provider() Provider
}
