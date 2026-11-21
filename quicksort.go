package quicksort

import (
	"cmp"
	"math/rand/v2"
	"runtime"
	"sync"
)

const (
	insertionCutoff    = 24      // Порог перехода на сортировку вставками
	seqCutoff          = 1 << 13 // Порог перехода на последовательную сортировку
	parPartitionCutoff = 1 << 18 // Порог включения параллельного разбиения
	minChunk           = 1 << 14 // Минимальный размер блока при параллельном разбиении
)

func Seq[T cmp.Ordered](a []T) {
	seqSort(a)
}

func Par[T cmp.Ordered](a []T) {
	if len(a) <= seqCutoff {
		seqSort(a)
		return
	}
	p := runtime.GOMAXPROCS(0)
	var buf []T
	if p > 1 && len(a) > parPartitionCutoff {
		buf = make([]T, len(a))
	}
	parSort(a, buf, p)
}

func seqSort[T cmp.Ordered](a []T) {
	for len(a) > insertionCutoff {
		m := partition(a)
		if m < len(a)-m {
			seqSort[T](a[:m])
			a = a[m:]
		} else {
			seqSort[T](a[m:])
			a = a[:m]
		}
	}
	insertionSort(a)
}

func parSort[T cmp.Ordered](a, buf []T, p int) {
	if len(a) <= seqCutoff {
		seqSort[T](a)
		return
	}

	var left, right, bufLeft, bufRight []T
	if p > 1 && len(a) > parPartitionCutoff {
		l, r := parPartition(a, buf, p)
		left, right = a[:l], a[r:]
		bufLeft, bufRight = buf[:l], buf[r:]
	} else {
		m := partition(a)
		left, right = a[:m], a[m:]
	}

	var wg sync.WaitGroup
	wg.Go(func() {
		parSort[T](left, bufLeft, (p+1)/2)
	})
	wg.Go(func() {
		parSort[T](right, bufRight, max(1, p/2))
	})

	wg.Wait()
}

func partition[T cmp.Ordered](a []T) int {
	n := len(a)
	p := a[rand.IntN(len(a))]
	i, j := -1, n
	for {
		for {
			i++
			if !less(a[i], p) {
				break
			}
		}
		for {
			j--
			if !less(p, a[j]) {
				break
			}
		}
		if i >= j {
			return j + 1
		}
		a[i], a[j] = a[j], a[i]
	}
}

func parPartition[T cmp.Ordered](a, buf []T, procs int) (l, r int) {
	n := len(a)
	p := a[rand.IntN(len(a))]

	k := max(1, min(4*procs, n/minChunk))
	chunk := func(c int) (lo, hi int) { return c * n / k, (c + 1) * n / k }

	type counts struct{ lt, eq, gt int }
	cnt := make([]counts, k)
	parallelFor(k, func(c int) {
		lo, hi := chunk(c)
		var lt, eq int
		for _, x := range a[lo:hi] {
			if less(x, p) {
				lt++
			} else if !less(p, x) {
				eq++
			}
		}
		cnt[c] = counts{lt, eq, hi - lo - lt - eq}
	})

	var totalLt, totalEq int
	for _, c := range cnt {
		totalLt += c.lt
		totalEq += c.eq
	}
	off := make([]counts, k)
	next := counts{0, totalLt, totalLt + totalEq}
	for c := range k {
		off[c] = next
		next.lt += cnt[c].lt
		next.eq += cnt[c].eq
		next.gt += cnt[c].gt
	}

	parallelFor(k, func(c int) {
		lo, hi := chunk(c)
		o := off[c]
		for _, x := range a[lo:hi] {
			if less(x, p) {
				buf[o.lt] = x
				o.lt++
			} else if less(p, x) {
				buf[o.gt] = x
				o.gt++
			} else {
				buf[o.eq] = x
				o.eq++
			}
		}
	})

	parallelFor(k, func(c int) {
		lo, hi := chunk(c)
		copy(a[lo:hi], buf[lo:hi])
	})

	return totalLt, totalLt + totalEq
}

func parallelFor(k int, f func(int)) {
	var wg sync.WaitGroup
	for c := range k {
		wg.Go(func() {
			f(c)
		})
	}
	wg.Wait()
}

func insertionSort[T cmp.Ordered](a []T) {
	for i := 1; i < len(a); i++ {
		x := a[i]
		j := i
		for j > 0 && less(x, a[j-1]) {
			a[j] = a[j-1]
			j--
		}
		a[j] = x
	}
}

func less[T cmp.Ordered](a, b T) bool {
	return a < b || (a != a && b == b)
}
