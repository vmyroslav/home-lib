package homestorage

import "github.com/vmyroslav/home-lib/homeconfig"

type Option = homeconfig.Option[config]

func WithCapacity(l uint64) Option {
	return homeconfig.OptionFunc[config](func(cfg *config) {
		cfg.capacity = l
	})
}
