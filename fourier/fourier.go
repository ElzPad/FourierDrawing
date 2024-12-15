package fourier

import (
	"math"
)

type FourierElem struct {
	Re			float64
	Im			float64
	Freq		float64
	Ampl		float64
	Phase		float64
}

func DiscreteFourierTransform(x []int) ([]FourierElem) {
	N := len(x)
	X := make([]FourierElem, N)

	for k:=0; k<N; k++ {
		re := 0.0
		im := 0.0

		for n:=0; n<N; n++ {
			arg := 2 * math.Pi * float64(n) * float64(k) / float64(N);
			re += float64(x[n]) * math.Cos(arg)
			im -= float64(x[n]) * math.Sin(arg)
		}

		freq := float64(k)
		ampl := math.Sqrt(re*re + im*im)
		phase := math.Atan2(im, re)

		X[k] = FourierElem{re, im, freq, ampl, phase}
	}

	return X
}