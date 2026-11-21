package quicksort_test

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"slices"
	"sort"
	"sync"
	"testing"

	"quicksort"
)

var algorithms = []struct {
	name string
	sort func([]int)
}{
	{"Seq", quicksort.Seq[int]},
	{"Par", quicksort.Par[int]},
}

func checkSort(t *testing.T, input []int) {
	t.Helper()

	want := slices.Clone(input)
	sort.Ints(want)

	for _, alg := range algorithms {
		got := slices.Clone(input)
		alg.sort(got)
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("%s: n=%d, index=%d, got=%d, want=%d, input[:16]=%v",
					alg.name, len(input), i, got[i], want[i], input[:min(len(input), 16)])
			}
		}
	}
}

func TestBasicCases(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1

	cases := []struct {
		name string
		data []int
	}{
		{"nil", nil},
		{"empty", []int{}},
		{"single", []int{5}},
		{"two sorted", []int{1, 2}},
		{"two reversed", []int{2, 1}},
		{"two equal", []int{2, 2}},
		{"all equal", []int{7, 7, 7, 7, 7}},
		{"sorted", []int{-3, -2, -1, 0, 1, 2}},
		{"reversed", []int{3, 2, 1, 0, -1, -2}},
		{"duplicates", []int{3, -1, 3, 0, -1, 0, 3}},
		{"alternating", []int{1, -1, 1, -1, 1, -1}},
		{"min and max int", []int{maxInt, 0, minInt, maxInt, -1, minInt}},
		{"single outlier", []int{0, 0, 0, maxInt, 0, 0}},
		{"almost sorted", []int{-5, -4, 20, -2, -1, 0}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkSort(t, tc.data)
		})
	}
}

func TestAllShortArrays(t *testing.T) {
	count := 1
	for n := 0; n <= 8; n++ {
		for code := 0; code < count; code++ {
			a := make([]int, n)
			x := code
			for i := range a {
				a[i] = x%3 - 1
				x /= 3
			}
			checkSort(t, a)
		}
		count *= 3
	}
}

func TestPatterns(t *testing.T) {
	sizes := []int{
		0, 1, 2, 3, 15, 16, 17, 23, 24, 25,
		31, 32, 33, 63, 64, 65, 127, 128, 129,
		1023, 1024, 1025, 4095, 4096, 4097,
	}

	for _, n := range sizes {
		ascending := make([]int, n)
		descending := make([]int, n)
		identical := make([]int, n)
		fewValues := make([]int, n)
		alternating := make([]int, n)
		almostSorted := make([]int, n)

		for i := range n {
			ascending[i] = i
			descending[i] = n - i
			identical[i] = 7
			fewValues[i] = i % 5
			if i%2 == 0 {
				alternating[i] = 1_000_000
			} else {
				alternating[i] = -1_000_000
			}
			almostSorted[i] = i
		}
		if n > 1 {
			almostSorted[0], almostSorted[n-1] = almostSorted[n-1], almostSorted[0]
		}

		cases := []struct {
			name string
			data []int
		}{
			{"ascending", ascending},
			{"descending", descending},
			{"all equal", identical},
			{"few values", fewValues},
			{"alternating", alternating},
			{"almost sorted", almostSorted},
		}
		for _, tc := range cases {
			t.Run(fmt.Sprintf("%s/n=%d", tc.name, n), func(t *testing.T) {
				checkSort(t, tc.data)
			})
		}
	}
}

func TestRandom(t *testing.T) {
	for _, seed := range []int64{1, 42, 2026} {
		rng := rand.New(rand.NewSource(seed))
		for _, n := range []int{0, 1, 2, 3, 8, 31, 100, 1024, 4096} {
			for _, width := range []int{2, 7, 1_000_000} {
				t.Run(fmt.Sprintf("seed=%d/n=%d/width=%d", seed, n, width), func(t *testing.T) {
					for range 10 {
						a := make([]int, n)
						for i := range a {
							a[i] = rng.Intn(width) - width/2
						}
						checkSort(t, a)
					}
				})
			}
		}
	}
}

func TestLargeArrays(t *testing.T) {
	const n = 200_000
	rng := rand.New(rand.NewSource(123))

	wide := make([]int, n)
	fewValues := make([]int, n)
	reversed := make([]int, n)
	for i := range n {
		wide[i] = rng.Intn(2*n) - n
		fewValues[i] = rng.Intn(3) - 1
		reversed[i] = n - i
	}

	for _, tc := range []struct {
		name string
		data []int
	}{
		{"random", wide},
		{"many duplicates", fewValues},
		{"reversed", reversed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			checkSort(t, tc.data)
		})
	}
}

func TestSubslice(t *testing.T) {
	want := []int{100, 200, -1, 0, 4, 5, 5, 300, 400}
	for _, alg := range algorithms {
		t.Run(alg.name, func(t *testing.T) {
			a := []int{100, 200, 5, -1, 4, 0, 5, 300, 400}
			alg.sort(a[2:7])
			if !slices.Equal(a, want) {
				t.Fatalf("got %v, want %v", a, want)
			}
		})
	}
}

func TestParWithDifferentCPUCounts(t *testing.T) {
	old := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(old)

	for _, cpus := range []int{1, 2, 4} {
		t.Run(fmt.Sprintf("cpus=%d", cpus), func(t *testing.T) {
			runtime.GOMAXPROCS(cpus)
			a := make([]int, 20_000)
			for i := range a {
				a[i] = (i * 97) % 1009
			}
			want := slices.Clone(a)
			sort.Ints(want)
			quicksort.Par(a)
			if !slices.Equal(a, want) {
				t.Fatal("Par returned incorrect result")
			}
		})
	}
}

func TestParConcurrentCalls(t *testing.T) {
	const workers = 12
	rng := rand.New(rand.NewSource(2026))

	var wg sync.WaitGroup
	for worker := range workers {
		a := make([]int, 10_000)
		for i := range a {
			a[i] = rng.Intn(101) - 50
		}
		want := slices.Clone(a)
		sort.Ints(want)

		wg.Add(1)
		go func(id int, data, expected []int) {
			defer wg.Done()
			quicksort.Par(data)
			if !slices.Equal(data, expected) {
				t.Errorf("worker %d returned incorrect result", id)
			}
		}(worker, a, want)
	}
	wg.Wait()
}

func FuzzSeqPar(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{2, 1, 0, 255, 128, 127})
	f.Add([]byte{7, 7, 7, 7, 7})
	f.Add([]byte{255, 0, 255, 0, 255})

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 4096 {
			t.Skip()
		}
		a := make([]int, len(data))
		for i, v := range data {
			a[i] = int(int8(v))
		}
		checkSort(t, a)
	})
}

func TestParallelBoundaries(t *testing.T) {
	sizes := []int{
		8191, 8192, 8193,
		262143, 262144, 262145,
	}

	rng := rand.New(rand.NewSource(42))

	for _, n := range sizes {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			a := make([]int, n)
			for i := range a {
				a[i] = rng.Intn(1000)
			}
			checkSort(t, a)
		})
	}
}

func TestFloatValues(t *testing.T) {
	for _, n := range []int{12, 300_000} {
		input := make([]float64, n)
		for i := range input {
			input[i] = float64(i%101 - 50)
		}
		input[0] = math.NaN()
		input[n/2] = math.Inf(-1)
		input[n-2] = math.Inf(1)
		input[n-1] = math.NaN()

		want := slices.Clone(input)
		sort.Float64s(want)

		for _, tc := range []struct {
			name string
			sort func([]float64)
		}{
			{"Seq", quicksort.Seq[float64]},
			{"Par", quicksort.Par[float64]},
		} {
			t.Run(tc.name, func(t *testing.T) {
				got := slices.Clone(input)
				tc.sort(got)
				for i := range want {
					if math.IsNaN(got[i]) && math.IsNaN(want[i]) {
						continue
					}
					if got[i] != want[i] {
						t.Fatalf("n=%d, index=%d: got %v, want %v", n, i, got[i], want[i])
					}
				}
			})
		}
	}
}

func TestStringValues(t *testing.T) {
	input := []string{"z", "", "aa", "a", "z", "ä", "A", "aa"}
	want := slices.Clone(input)
	sort.Strings(want)
	for _, tc := range []struct {
		name string
		sort func([]string)
	}{
		{"Seq", quicksort.Seq[string]},
		{"Par", quicksort.Par[string]},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := slices.Clone(input)
			tc.sort(got)
			if !slices.Equal(got, want) {
				t.Fatalf("got %v, want %v", got, want)
			}
		})
	}
}

// benchmarks
func BenchmarkSeq(b *testing.B) {
	benchmarkSort(b, quicksort.Seq[int])
}

func BenchmarkPar(b *testing.B) {
	benchmarkSort(b, quicksort.Par[int])
}

func benchmarkSort(b *testing.B, sortFn func([]int)) {
	rng := rand.New(rand.NewSource(42))
	original := make([]int, 100_000)

	for i := range original {
		original[i] = rng.Intn(1_000_000)
	}

	a := make([]int, len(original))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		copy(a, original)
		b.StartTimer()

		sortFn(a)
	}
}
