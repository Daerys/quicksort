package quicksort

import (
	"math/rand"
	"sort"
	"testing"
)

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

func TestSeqRandom(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	sizes := []int{10, 100, 1000, 10_000, 100_000, 1_000_000}

	for _, n := range sizes {
		for k := range 10 {
			a := make([]int, n)
			for i := range a {
				a[i] = rng.Intn(201) - 100
			}

			expected := append([]int(nil), a...)
			sort.Ints(expected)

			Seq(a)

			for i := range a {
				if a[i] != expected[i] {
					t.Fatalf("n=%d, iteration=%d, index=%d", n, k, i)
				}
			}
		}
	}
}
