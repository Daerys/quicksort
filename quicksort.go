package quicksort

import (
	"math/rand/v2"
	"sync"
)

func Seq(a []int) {
	if len(a) < 2 {
		return
	}

	i, j := 0, len(a)-1
	pivot := a[rand.IntN(len(a))]

	for i <= j {
		for a[i] < pivot {
			i++
		}
		for a[j] > pivot {
			j--
		}

		if i <= j {
			a[i], a[j] = a[j], a[i]
			i++
			j--
		}
	}

	if j > 0 {
		Seq(a[:j+1])
	}
	if i < len(a) {
		Seq(a[i:])
	}
}

func Par(a []int) {
	par(a, 4)
}

func par(a []int, p int) {
	if len(a) < 2 {
		return
	}
	if p <= 1 {
		Seq(a)
		return
	}

	i, j := 0, len(a)-1
	pivot := a[rand.IntN(len(a))]

	for i <= j {
		for a[i] < pivot {
			i++
		}
		for a[j] > pivot {
			j--
		}

		if i <= j {
			a[i], a[j] = a[j], a[i]
			i++
			j--
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		par(a[:j+1], p/2)
	}()

	par(a[i:], p-p/2)
	wg.Wait()
}
