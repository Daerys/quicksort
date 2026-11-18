package quicksort

import "testing"

// Набор простых случаев
func TestSeq(t *testing.T) {
	tests := [][]int{
		{},
		{1},
		{2, 1},
		{3, 1, 2},
		{1, 2, 3, 4, 5},
		{5, 4, 3, 2, 1},
		{2, 2, 2, 2},
		{3, -1, 0, -5, 3},
	}

	for _, a := range tests {
		Seq(a)
		for i := 1; i < len(a); i++ {
			if a[i-1] > a[i] {
				t.Fatalf("not sorted: %v", a)
			}
		}
	}
}
