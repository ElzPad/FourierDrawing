package fourier

import (
	"math"
	"math/cmplx"
	"sort"
)

type FourierElement struct{
	Freq int
	Val complex128
}

func DiscreteFourierTransform(x []float64, sortByModule bool) ([]FourierElement) {
	N := len(x)
	X := make([]FourierElement, N)

	for k:=0; k<N; k++ {
		X[k].Freq = k
		for n:=0; n<N; n++ {
			arg := 2 * math.Pi * float64(n) * float64(k) / float64(N);
			X[k].Val += complex(x[n], 0) * complex(math.Cos(arg), -math.Sin(arg))
		}
	}

	if (sortByModule) {
		sort.Slice(X, func (i, j int) (bool) {
			return cmplx.Abs(X[i].Val) > cmplx.Abs(X[j].Val) 
		})
	}

	return X
}

func InverseDFT(X []FourierElement) ([]float64) {
	N := len(X)
	x := make([]float64, N)

	for n:=0; n<N; n++ {
		res := 0.0+0.0i

		for k:=0; k<N; k++ {
			arg := 2 * math.Pi * float64(n) * float64(X[k].Freq) / float64(N);
			res += X[k].Val * complex(math.Cos(arg), math.Sin(arg))
		}
		x[n] = real(res)/float64(N)
	}

	return x
}