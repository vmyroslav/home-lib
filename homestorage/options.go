package homestorage

type Option interface {
	apply(cfg *config)
}

type optionFn func(cfg *config)

func (fn optionFn) apply(cfg *config) {
	fn(cfg)
}

func WithCapacity(l uint64) Option {
	return optionFn(func(cfg *config) {
		cfg.capacity = l
	})
}
