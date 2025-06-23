package hometests

import (
	"fmt"
	"net"
	"testing"

	"github.com/vmyroslav/home-lib/homemath"
)

// RandomPort generates a random port, trying up to maxRetries times.
func RandomPort(t *testing.T) int {
	t.Helper()

	maxRetries := 5

	port, err := randomPort(&homemathRandomizer{}, defaultListener{}, maxRetries)
	if err != nil {
		t.Fatal(err)
	}

	return port
}

// randomPort is the testable version of RandomPort.
func randomPort(rd randomizer, listener networkListener, maxRetries int) (int, error) {
	for retries := 0; retries < maxRetries; retries++ {
		port := rd.Intn(65535-1024) + 1024

		conn, err := listener.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			_ = conn.Close()
			return port, nil
		}
	}

	return 0, fmt.Errorf("failed to generate a valid port number after %d retries", maxRetries)
}

type randomizer interface {
	Intn(n int) int
}

type networkListener interface {
	Listen(network, address string) (net.Listener, error)
}

type defaultListener struct{}

func (d defaultListener) Listen(network, address string) (net.Listener, error) {
	return net.Listen(network, address)
}

// homemathRandomizer implements the randomizer interface using homemath functions.
type homemathRandomizer struct{}

func (h *homemathRandomizer) Intn(n int) int {
	return homemath.RandInt(n)
}
