package homestorage

const defaultCapacity = 1024

type config struct {
	capacity uint64
}

func newDefaultConfig() *config {
	return &config{capacity: defaultCapacity}
}
