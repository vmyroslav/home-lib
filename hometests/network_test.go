package hometests

import (
	"errors"
	"net"
	"testing"
)

func TestRandomPort(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		randValue    int
		listenErr    error
		maxRetries   int
		wantErr      bool
		expectedPort int
		retriesCount int
	}{
		{
			name:         "Success on first try",
			randValue:    1000,
			listenErr:    nil,
			maxRetries:   5,
			wantErr:      false,
			expectedPort: 2024,
			retriesCount: 0,
		},
		{
			name:         "Success with retries",
			randValue:    1000,
			listenErr:    nil,
			maxRetries:   5,
			wantErr:      false,
			expectedPort: 2024,
			retriesCount: 3,
		},
		{
			name:         "Failure after max retries",
			randValue:    1000,
			listenErr:    errors.New("port not available"),
			maxRetries:   5,
			wantErr:      true,
			expectedPort: 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockRand := &mockRand{intnValue: tc.randValue}
			mockListener := &mockListener{listenErr: tc.listenErr, failedCallsNum: tc.retriesCount}

			port, err := randomPort(mockRand, mockListener, tc.maxRetries)

			if (err != nil) != tc.wantErr {
				t.Errorf("RandomPort() error = %v, wantErr %v", err, tc.wantErr)
			}

			if port != tc.expectedPort {
				t.Errorf("RandomPort() port = %v, expectedPort %v", port, tc.expectedPort)
			}
		})
	}
}

type mockRand struct {
	intnValue int
}

func (m *mockRand) Intn(_ int) int {
	return m.intnValue
}

type mockListener struct {
	listenErr      error
	failedCallsNum int

	callCount int
}

func (m *mockListener) Listen(_, _ string) (net.Listener, error) {
	if m.listenErr != nil {
		return nil, m.listenErr
	}

	m.callCount++
	if m.callCount < m.failedCallsNum {
		return nil, errors.New("port is busy")
	}

	return &net.TCPListener{}, nil
}
