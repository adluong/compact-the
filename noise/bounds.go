// Package noise implements smudging noise sampling and noise bound
// computation for the threshold HE scheme.
package noise

import "math/big"

// ComputeBsmRegimeB computes the smudging noise bound for Regime B (all-F RLWE).
// B_sm = B_W * B_ct * 2^kappa
func ComputeBsmRegimeB(bw, bctLog2, kappa int) int {
	// log2(B_sm) = log2(B_W) + log2(B_ct) + kappa
	// For B_W=1: log2(B_sm) = 0 + bctLog2 + kappa
	bwLog2 := 0
	bwTmp := bw
	for bwTmp > 1 {
		bwLog2++
		bwTmp >>= 1
	}
	return bwLog2 + bctLog2 + kappa
}

// VerifyCorrectness checks the noise margin: |δ|*B_ct + Λ_S*B_sm < Δ/2
// where Δ = Q/T for BFV-style encoding.
// All values are in log2 scale for comparison.
func VerifyCorrectness(absDelta int64, lambdaS int64, bsmLog2 int, bctLog2 int, logQ int, logT int) bool {
	// |δ|*B_ct ≈ absDelta * 2^bctLog2
	// Λ_S * B_sm ≈ lambdaS * 2^bsmLog2
	// Δ/2 = Q/(2T) ≈ 2^(logQ - logT - 1)

	// Use big.Int for exact computation
	deltaBct := new(big.Int).Mul(big.NewInt(absDelta), new(big.Int).Lsh(big.NewInt(1), uint(bctLog2)))
	lambdaBsm := new(big.Int).Mul(big.NewInt(lambdaS), new(big.Int).Lsh(big.NewInt(1), uint(bsmLog2)))
	noise := new(big.Int).Add(deltaBct, lambdaBsm)

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
