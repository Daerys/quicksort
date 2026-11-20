package main

import (
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"runtime"
	"slices"
	"time"

	"quicksort"
)

func main() {
	n := flag.Int("n", 100_000_000, "размер массива")
	runs := flag.Int("runs", 5, "число повторов")
	seed := flag.Uint64("seed", 42, "seed генератора")
	flag.Parse()

	if *n < 0 || *runs < 1 {
		fmt.Fprintln(os.Stderr, "n должен быть >= 0, runs >= 1")
		os.Exit(2)
	}

	a := make([]int, *n)
	b := make([]int, *n)
	var seqTotal, parTotal time.Duration

	fmt.Printf("Go %s, %s/%s, NumCPU=%d, n=%d, runs=%d, seed=%d\n",
		runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), *n, *runs, *seed)
	fmt.Println("| Run | Seq (1 CPU), s | Par (4 CPU), s | Speedup |")
	fmt.Println("|---:|---:|---:|---:|")

	for run := 1; run <= *runs; run++ {
		rng := rand.New(rand.NewPCG(*seed, uint64(run)))
		for i := range a {
			a[i] = rng.Int()
		}
		copy(b, a)

		var seqTime, parTime time.Duration
		// Меняем порядок, чтобы одна версия не была постоянно первой.
		if run%2 == 1 {
			seqTime = measure(1, func() { quicksort.Seq(a) })
			parTime = measure(4, func() { quicksort.Par(b) })
		} else {
			parTime = measure(4, func() { quicksort.Par(b) })
			seqTime = measure(1, func() { quicksort.Seq(a) })
		}

		if !slices.IsSorted(a) || !slices.IsSorted(b) || !slices.Equal(a, b) {
			fmt.Fprintf(os.Stderr, "неправильный результат на запуске %d\n", run)
			os.Exit(1)
		}

		seqTotal += seqTime
		parTotal += parTime
		fmt.Printf("| %d | %.3f | %.3f | %.2fx |\n",
			run, seqTime.Seconds(), parTime.Seconds(), seqTime.Seconds()/parTime.Seconds())
	}

	seqAvg := seqTotal.Seconds() / float64(*runs)
	parAvg := parTotal.Seconds() / float64(*runs)
	fmt.Printf("| **Average** | **%.3f** | **%.3f** | **%.2fx** |\n",
		seqAvg, parAvg, seqAvg/parAvg)
}

func measure(procs int, sortFn func()) time.Duration {
	runtime.GOMAXPROCS(procs)
	start := time.Now()
	sortFn()
	return time.Since(start)
}
