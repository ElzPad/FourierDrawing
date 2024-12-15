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

func InverseDFT(X []complex128) ([]float64	) {
	N := len(X)
	x := make([]float64, N)

	for n:=0; n<N; n++ {
		res := 0.0+0.0i

		for k:=0; k<N; k++ {
			arg := 2 * math.Pi * float64(n) * float64(k) / float64(N);
			res += X[k] * complex(math.Cos(arg), math.Sin(arg))
		}
		x[n] = real(res)/float64(N)
	}

	return x
}