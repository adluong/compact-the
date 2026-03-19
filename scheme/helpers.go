// Package scheme implements the five core algorithms of the compact
// threshold homomorphic encryption scheme: Setup, Share, PartDec, Combine, FinDec.
package scheme

import (
	"math/big"

	"github.com/tuneinsight/lattigo/v6/ring"
)

// scalarMulSigned computes result = scalar * poly in R_q, where scalar is a
// signed integer. Handles negative scalars via Neg + MulScalar.
func scalarMulSigned(ringQ *ring.Ring, poly ring.Poly, scalar int64, result ring.Poly) {
	if scalar == 0 {
		for i := 0; i <= ringQ.Level(); i++ {
			for j := range result.Coeffs[i] {
				result.Coeffs[i][j] = 0
			}
		}
		return
	}
	absScalar := uint64(scalar)
	if scalar < 0 {
		absScalar = uint64(-scalar)
	}
	ringQ.MulScalar(poly, absScalar, result)
	if scalar < 0 {
		ringQ.Neg(result, result)
	}
}

// roundDivBigInt computes round(a/b) for big integers where b > 0.
// Uses the standard rounding convention: round to nearest, ties away from zero.
func roundDivBigInt(a, b *big.Int) *big.Int {
	half := new(big.Int).Rsh(b, 1)
	adjusted := new(big.Int).Set(a)
	if a.Sign() >= 0 {
		adjusted.Add(adjusted, half)
	} else {
		adjusted.Sub(adjusted, half)
	}
	return new(big.Int).Quo(adjusted, b) // truncate toward zero
}

// copyPoly creates a deep copy of a polynomial.
func copyPoly(ringQ *ring.Ring, src ring.Poly) ring.Poly {
	dst := ringQ.NewPoly()
	dst.CopyLvl(ringQ.Level(), src)
	return dst
}
