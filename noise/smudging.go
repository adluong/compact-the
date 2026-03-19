package noise

import (
	"crypto/rand"
	"math/big"

	"github.com/tuneinsight/lattigo/v6/ring"
)

// SampleSmudgingNoise samples a polynomial with each coefficient uniformly
// random in [-B_sm, B_sm] and stores it in RNS representation.
// bsmLog2 is log₂(B_sm). The noise MUST be uniform, not Gaussian.
func SampleSmudgingNoise(ringQ *ring.Ring, bsmLog2 int) ring.Poly {
	n := ringQ.N()
	level := ringQ.Level()
	result := ringQ.NewPoly()

	// B_sm = 2^bsmLog2
	bsm := new(big.Int).Lsh(big.NewInt(1), uint(bsmLog2))
	// range = 2*B_sm + 1
	rangeSize := new(big.Int).Lsh(bsm, 1)
	rangeSize.Add(rangeSize, big.NewInt(1))

	for l := 0; l < n; l++ {
		// Sample uniform in [0, 2*B_sm]
		val, _ := rand.Int(rand.Reader, rangeSize)
		// Subtract B_sm to center: val ∈ [-B_sm, B_sm]
		val.Sub(val, bsm)

		// Reduce mod each RNS prime and store
		for i := 0; i <= level; i++ {
			qi := ringQ.SubRings[i].Modulus
			qiBig := new(big.Int).SetUint64(qi)
			// Compute val mod qi (positive representative)
			coeff := new(big.Int).Mod(val, qiBig)
			result.Coeffs[i][l] = coeff.Uint64()
		}
	}

	return result
}
