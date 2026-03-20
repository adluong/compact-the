// Package noise implements smudging noise sampling and noise bound
// computation for the threshold HE scheme.
package noise

import "math/big"

// ComputeBsmRegimeB computes the smudging noise bound for Regime B (all-F RLWE).
// B_sm = B_W * B_ct * 2^kappa = B_W * 2^(bctLog2 + kappa), computed exactly.
func ComputeBsmRegimeB(bw, bctLog2, kappa int) *big.Int {
	bsm := new(big.Int).SetInt64(int64(bw))
	bsm.Lsh(bsm, uint(bctLog2+kappa))
	return bsm
}

// VerifyCorrectness checks the noise margin: |δ|*B_ct + Λ_S*B_sm < Δ/2
// where Δ = Q/T for BFV-style encoding.
func VerifyCorrectness(absDelta int64, lambdaS int64, bsm *big.Int, bctLog2 int, logQ int, logT int) bool {
	// |δ|*B_ct = absDelta * 2^bctLog2
	deltaBct := new(big.Int).Mul(big.NewInt(absDelta), new(big.Int).Lsh(big.NewInt(1), uint(bctLog2)))
	// Λ_S * B_sm (exact)
	lambdaBsm := new(big.Int).Mul(big.NewInt(lambdaS), bsm)
	noise := new(big.Int).Add(deltaBct, lambdaBsm)

	// Δ/2 = Q/(2T) ≈ 2^(logQ - logT - 1)
	halfDelta := new(big.Int).Lsh(big.NewInt(1), uint(logQ-logT-1))

	return noise.Cmp(halfDelta) < 0
}

// LambdaS computes Λ_S = Σ |λ̂_j| for a set of cofactors.
func LambdaS(cofactors []int64) int64 {
	sum := int64(0)
	for _, c := range cofactors {
		if c < 0 {
			sum += -c
		} else {
			sum += c
		}
	}
	return sum
}
