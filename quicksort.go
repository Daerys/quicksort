package quicksort

func Seq(a []int) {
	if len(a) < 2 {
		return
	}

	i, j := 0, len(a)-1
	pivot := a[len(a)/2]

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
