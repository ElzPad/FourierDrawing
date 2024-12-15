package fourier

import (
	"math"
)

func DiscreteFourierTransform(x []float64) ([]complex128) {
	N := len(x)
	X := make([]complex128, N)

	for k:=0; k<N; k++ {
		for n:=0; n<N; n++ {
			arg := 2 * math.Pi * float64(n) * float64(k) / float64(N);
			X[k] += complex(x[n], 0) * complex(math.Cos(arg), -math.Sin(arg))
		}
	}

	return X
}