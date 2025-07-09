package homeconfig

// Option defines a generic interface for applying a configuration to a struct of type T.
// The type parameter T can be any type, but it's intended for configuration structs.
type Option[T any] interface {
	Apply(*T)
}

// OptionFunc allows you to use ordinary functions as configuration options.
type OptionFunc[T any] func(*T)

func (o OptionFunc[T]) Apply(c *T) {
	o(c)
}

// ApplyOptions applies a slice of options to a configuration struct.
func ApplyOptions[T any](config *T, options ...Option[T]) {
	for _, opt := range options {
		opt.Apply(config)
	}
}
