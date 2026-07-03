package controller

import (
	"math"

	"github.com/nobonobo/q16"
)

type LPF struct {
	value q16.Fixed
	alpha q16.Fixed // 0 < alpha <= 1
}

func NewLPF(alpha q16.Fixed) *LPF {
	return &LPF{
		alpha: alpha,
	}
}

func (f *LPF) Reset(v q16.Fixed) {
	f.value = v
}

func (f *LPF) Update(v q16.Fixed) q16.Fixed {
	f.value += q16.Mul(f.alpha, (v - f.value))
	return f.value
}

func CalcAlpha(fc float64, dt float64) q16.Fixed {
	tau := 1.0 / (2.0 * math.Pi * fc)
	alpha := dt / (tau + dt)
	return q16.FromFloat64(alpha)
}
