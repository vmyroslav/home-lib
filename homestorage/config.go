package homestorage

const defaultLimit = 1024

type config struct {
	capacity uint64
}

func newDefaultConfig() *config {
	return &config{capacity: defaultLimit}
}
