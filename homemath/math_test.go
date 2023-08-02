package homemath

import (
	"testing"

	"golang.org/x/exp/constraints"
)

func TestMin_Max_Ints(t *testing.T) {
	t.Parallel()

	type testCase[T constraints.Ordered] struct {
		name    string
		args    []T
		wantMax T
		wantMin T
	}
	tests := []testCase[int]{
		{
			name:    "MinMax of empty slice",
			args:    []int{},
			wantMax: 0,
			wantMin: 0,
		},
		{
			name:    "MinMax of positive ints",
			args:    []int{1, 2, 3, 4, 5},
			wantMax: 5,
			wantMin: 1,
		},
		{
			name:    "MinMax of negative ints",
			args:    []int{-1, -2, -3, -4, -5},
			wantMax: -1,
			wantMin: -5,
		},
		{
			name:    "MinMax of mixed ints",
			args:    []int{-1, 2, -3, 4, -5},
			wantMax: 4,
			wantMin: -5,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := Max(tt.args...); got != tt.wantMax {
				t.Errorf("Max() = %v, want %v", got, tt.wantMax)
			}

			if got := Min(tt.args...); got != tt.wantMin {
				t.Errorf("Min() = %v, want %v", got, tt.wantMin)
			}
		})
	}
}

func TestMin_Max_Floats(t *testing.T) {
	t.Parallel()

	type testCase[T constraints.Ordered] struct {
		name    string
		args    []T
		wantMax T
		wantMin T
	}
	tests := []testCase[float64]{
		{
			name:    "MinMax of empty slice",
			args:    []float64{},
			wantMax: 0,
			wantMin: 0,
		},
		{
			name:    "MinMax of positive floats",
			args:    []float64{1.1, 2.2, 3.3, 4.4, 5.5},
			wantMax: 5.5,
			wantMin: 1.1,
		},
		{
			name:    "MinMax of negative floats",
			args:    []float64{-1.1, -2.2, -3.3, -4.4, -5.5},
			wantMax: -1.1,
			wantMin: -5.5,
		},
		{
			name:    "MinMax of mixed floats",
			args:    []float64{-1.1, 2.2, -3.3, 4.4, -5.5},
			wantMax: 4.4,
			wantMin: -5.5,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := Max(tt.args...); got != tt.wantMax {
				t.Errorf("Max() = %v, want %v", got, tt.wantMax)
			}

			if got := Min(tt.args...); got != tt.wantMin {
				t.Errorf("Min() = %v, want %v", got, tt.wantMin)
			}
		})
	}
}

func Test_Min_Max_Strings(t *testing.T) {
	t.Parallel()

	type testCase[T constraints.Ordered] struct {
		name    string
		args    []T
		wantMax T
		wantMin T
	}

	tests := []testCase[string]{
		{
			name:    "MinMax of empty slice",
			args:    []string{},
			wantMax: "",
			wantMin: "",
		},
		{
			name:    "MinMax of chars",
			args:    []string{"a", "b", "c", "d", "e"},
			wantMax: "e",
			wantMin: "a",
		},
		{
			name:    "MinMax of simple strings",
			args:    []string{"aa", "bb", "cc", "dd", "ee"},
			wantMax: "ee",
			wantMin: "aa",
		},
		{
			name:    "MinMax of strings",
			args:    []string{"hello", "hello me", " ", "_", ""},
			wantMax: "hello me",
			wantMin: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := Max(tt.args...); got != tt.wantMax {
				t.Errorf("Max() = %v, want %v", got, tt.wantMax)
			}

			if got := Min(tt.args...); got != tt.wantMin {
				t.Errorf("Min() = %v, want %v", got, tt.wantMin)
			}
		})
	}
}
