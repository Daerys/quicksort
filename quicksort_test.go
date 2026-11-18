package quicksort

import (
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"testing"
)

func checkBoth(t *testing.T, input []int) {
	t.Helper()

	want := append([]int(nil), input...)
	sort.Ints(want)

	for _, alg := range []struct {
		name string
		run  func([]int)
	}{
		{"Seq", Seq},
		{"Par", Par},
	} {
		got := append([]int(nil), input...)
		alg.run(got)

		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("%s: n=%d, индекс=%d, получили=%d, ожидали=%d, вход=%v",
					alg.name, len(input), i, got[i], want[i], preview(input))
			}
		}
	}
}

func preview(a []int) []int {
	if len(a) > 16 {
		return a[:16]
	}
	return a
}

func TestExamples(t *testing.T) {
	max := int(^uint(0) >> 1)
	min := -max - 1

	tests := []struct {
		name string
		a    []int
	}{
		{"nil", nil},
		{"empty", []int{}},
		{"single", []int{42}},
		{"pair_sorted", []int{1, 2}},
		{"pair_reverse", []int{2, 1}},
		{"pair_equal", []int{7, 7}},
		{"all_equal", []int{5, 5, 5, 5, 5, 5}},
		{"ascending", []int{-3, -2, -1, 0, 1, 2, 3}},
		{"descending", []int{3, 2, 1, 0, -1, -2, -3}},
		{"duplicates", []int{3, -1, 3, 0, -1, 0, 3}},
		{"alternating", []int{1, -1, 1, -1, 1, -1}},
		{"extreme_values", []int{max, 0, min, max, -1, min, 1}},
		{"one_outlier", []int{0, 0, 0, 0, max, 0, 0, 0}},
		{"almost_sorted", []int{-5, -4, -3, 20, -1, 0, 1}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			checkBoth(t, tc.a)
		})
	}
}

// Все массивы длины 0..8 из {-1, 0, 1}.
func TestExhaustiveSmall(t *testing.T) {
	for n, count := 0, 1; n <= 8; n, count = n+1, count*3 {
		for mask := 0; mask < count; mask++ {
			a := make([]int, n)
			x := mask
			for i := range a {
				a[i] = x%3 - 1
				x /= 3
			}
			checkBoth(t, a)
		}
	}
}

func TestPatterns(t *testing.T) {
	sizes := []int{
		0, 1, 2, 3, 7, 8, 15, 16, 17, 23, 24, 25,
		31, 32, 33, 63, 64, 65, 127, 128, 129, 255,
		256, 257, 1023, 1024, 1025, 4095, 4096, 4097,
	}

	for _, n := range sizes {
		for pattern := range 7 {
			a := make([]int, n)
			for i := range a {
				switch pattern {
				case 0: // возрастает
					a[i] = i
				case 1: // убывает
					a[i] = n - i
				case 2: // все одинаковые
					a[i] = 7
				case 3: // мало различных значений
					a[i] = i % 5
				case 4: // большие и маленькие по очереди
					if i%2 == 0 {
						a[i] = 1_000_000
					} else {
						a[i] = -1_000_000
					}
				case 5: // растёт до середины, затем убывает
					if i < n/2 {
						a[i] = i
					} else {
						a[i] = n - i
					}
				case 6: // почти упорядоченный
					a[i] = i
				}
			}
			if pattern == 6 && n > 1 {
				a[0], a[n-1] = a[n-1], a[0]
			}
			checkBoth(t, a)
		}
	}
}

func TestRandom(t *testing.T) {
	for _, seed := range []int64{1, 42, 2026} {
		rng := rand.New(rand.NewSource(seed))
		for k := range 300 {
			n := rng.Intn(2049)
			a := make([]int, n)
			for i := range a {
				switch k % 4 {
				case 0:
					a[i] = rng.Intn(7) - 3
				case 1:
					a[i] = rng.Intn(2_000_001) - 1_000_000
				case 2:
					a[i] = rng.Intn(2)
				case 3:
					a[i] = i + rng.Intn(5) - 2
				}
			}
			checkBoth(t, a)
		}
	}
}

func TestLarge(t *testing.T) {
	const n = 200_000
	rng := rand.New(rand.NewSource(123))

	for pattern := range 4 {
		a := make([]int, n)
		for i := range a {
			switch pattern {
			case 0:
				a[i] = rng.Intn(n)
			case 1:
				a[i] = n - i
			case 2:
				a[i] = 1
			case 3:
				a[i] = rng.Intn(3) - 1
			}
		}
		checkBoth(t, a)
	}
}

func TestSortOnlySubslice(t *testing.T) {
	for _, alg := range []struct {
		name string
		run  func([]int)
	}{
		{"Seq", Seq},
		{"Par", Par},
	} {
		t.Run(alg.name, func(t *testing.T) {
			a := []int{100, 200, 5, -1, 4, 0, 5, 300, 400}
			alg.run(a[2:7])
			want := []int{100, 200, -1, 0, 4, 5, 5, 300, 400}
			for i := range a {
				if a[i] != want[i] {
					t.Fatalf("изменился элемент %d: got %v, want %v", i, a, want)
				}
			}
		})
	}
}

func TestParBudgets(t *testing.T) {
	for _, p := range []int{0, 1, 2, 3, 4, 5, 8} {
		a := []int{8, -2, 7, 7, 1, 0, 5, -1, 4, 3, 9, 9, 2, 6}
		par(a, p)
		for i := 1; i < len(a); i++ {
			if a[i-1] > a[i] {
				t.Fatalf("p=%d: %v", p, a)
			}
		}
	}
}

func TestParDifferentGOMAXPROCS(t *testing.T) {
	old := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(old)

	for _, p := range []int{1, 2, 4} {
		runtime.GOMAXPROCS(p)
		a := make([]int, 20_000)
		for i := range a {
			a[i] = (i * 97) % 1009
		}
		checkBoth(t, a)
	}
}

func TestParConcurrentCalls(t *testing.T) {
	const calls = 12
	rng := rand.New(rand.NewSource(2026))
	data := make([][]int, calls)
	want := make([][]int, calls)
	bad := make([]bool, calls)
	for i := range data {
		data[i] = make([]int, 10_000)
		for j := range data[i] {
			data[i][j] = rng.Intn(101) - 50
		}
		want[i] = append([]int(nil), data[i]...)
		sort.Ints(want[i])
	}

	var wg sync.WaitGroup
	for i := range data {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			Par(data[i])
			for j := range data[i] {
				if data[i][j] != want[i][j] {
					bad[i] = true
					return
				}
			}
		}(i)
	}
	wg.Wait()

	for i, failed := range bad {
		if failed {
			t.Fatalf("одновременный вызов %d дал неправильный результат", i)
		}
	}
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
		for i, x := range data {
			a[i] = int(int8(x))
		}
		checkBoth(t, a)
	})
}
