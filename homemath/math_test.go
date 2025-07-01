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

	tests := []testCase[int]{ //nolint:wsl
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

func TestSum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []int
		want int
	}{
		{"empty", []int{}, 0},
		{"single", []int{5}, 5},
		{"multiple positive", []int{1, 2, 3, 4, 5}, 15},
		{"mixed", []int{-1, 2, -3, 4}, 2},
		{"all negative", []int{-1, -2, -3}, -6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := Sum(tt.args...); got != tt.want {
				t.Errorf("Sum(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}

	t.Run("float64", func(t *testing.T) {
		t.Parallel()

		if got := Sum(1.1, 2.2, 3.3); got != 6.6 {
			t.Errorf("Sum(1.1, 2.2, 3.3) = %v, want 6.6", got)
		}
	})
}

func TestSumSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []int
		want int
	}{
		{"empty slice", []int{}, 0},
		{"single element", []int{5}, 5},
		{"multiple positive", []int{1, 2, 3, 4, 5}, 15},
		{"mixed", []int{-1, 2, -3, 4}, 2},
		{"all negative", []int{-1, -2, -3}, -6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := SumSlice(tt.args); got != tt.want {
				t.Errorf("SumSlice(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}

	t.Run("float64", func(t *testing.T) {
		t.Parallel()

		slice := []float64{1.1, 2.2, 3.3}

		if got := SumSlice(slice); got != 6.6 {
			t.Errorf("SumSlice(%v) = %v, want 6.6", slice, got)
		}
	})
}

func TestRandInt(t *testing.T) {
	t.Parallel()

	t.Run("valid range", func(t *testing.T) {
		t.Parallel()

		for i := 0; i < 100; i++ {
			got := RandInt(10)
			if got < 0 || got >= 10 {
				t.Errorf("RandInt(10) = %v, want [0, 10)", got)
			}
		}
	})

	t.Run("zero input", func(t *testing.T) {
		t.Parallel()

		if got := RandInt(0); got != 0 {
			t.Errorf("RandInt(0) = %v, want 0", got)
		}
	})

	t.Run("negative input", func(t *testing.T) {
		t.Parallel()

		if got := RandInt(-5); got != 0 {
			t.Errorf("RandInt(-5) = %v, want 0", got)
		}
	})
}

func TestRandIntRange(t *testing.T) {
	t.Parallel()

	t.Run("valid range", func(t *testing.T) {
		t.Parallel()

		for i := 0; i < 100; i++ {
			got := RandIntRange(5, 15)
			if got < 5 || got > 15 {
				t.Errorf("RandIntRange(5, 15) = %v, want [5, 15]", got)
			}
		}
	})

	t.Run("equal min max", func(t *testing.T) {
		t.Parallel()

		if got := RandIntRange(5, 5); got != 5 {
			t.Errorf("RandIntRange(5, 5) = %v, want 5", got)
		}
	})

	t.Run("min greater than max", func(t *testing.T) {
		t.Parallel()

		if got := RandIntRange(10, 5); got != 10 {
			t.Errorf("RandIntRange(10, 5) = %v, want 10", got)
		}
	})
}
