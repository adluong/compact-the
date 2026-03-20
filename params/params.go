// Package params defines the public parameters for the compact threshold
// homomorphic encryption scheme, wrapping Lattigo v6 BGV parameters.
package params

import (
	"fmt"
	"math/big"

	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

// PublicParams contains all public parameters for the THE scheme.
type PublicParams struct {
	BGVParams bgv.Parameters // Lattigo BGV parameters (unified BGV/BFV)
	N         int            // Number of parties
	T         int            // Reconstruction threshold
	BW        int            // Entry bound on W matrix
	M         [][]int64      // N × T share-generation matrix
	W         [][]int64      // (N-T+1) × T L-party block
	Kappa     int            // Statistical security parameter κ
	Bsm       *big.Int       // B_sm smudging noise bound (exact)
}

// NewBGVParams creates Lattigo BGV parameters for the given configuration.
// Uses LogQ to let Lattigo auto-generate NTT-friendly primes.
// The PlaintextModulus must be coprime with Q (Lattigo requirement).
func NewBGVParams(logN int, logQ []int, logP []int, plaintextModulus uint64) (bgv.Parameters, error) {
	paramLit := bgv.ParametersLiteral{
		LogN:             logN,
		LogQ:             logQ,
		LogP:             logP,
		PlaintextModulus: plaintextModulus,
	}
	params, err := bgv.NewParametersFromLiteral(paramLit)
	if err != nil {
		return bgv.Parameters{}, fmt.Errorf("failed to create BGV parameters: %w", err)
	}
	return params, nil
}

// DefaultRegimeB returns BGV parameters for Regime B (all-F RLWE simulation).
// n=2^13, logQ≈218, p=65537, κ=40.
// Note: p=65537 is used instead of 131072 because Lattigo requires gcd(p,q)=1,
// and 131072=2^17 shares a factor of 2 with NTT-friendly primes.
func DefaultRegimeB() (bgv.Parameters, error) {
	return NewBGVParams(
		13,                       // LogN = 13 → n = 8192
		[]int{55, 55, 55, 53},   // LogQ: sum ≈ 218
		[]int{55},               // LogP: auxiliary modulus for key-switching
		65537,                   // PlaintextModulus: prime, 65537 = 2^16 + 1
	)
}

// VerifyParams checks essential parameter properties.
func VerifyParams(params bgv.Parameters, numParties int) error {
	// Check ring degree is power of 2
	n := params.N()
	if n&(n-1) != 0 {
		return fmt.Errorf("ring degree %d is not a power of 2", n)
	}

	// Check p > N (needed for δ⁻¹ mod p to exist for small determinants)
	p := params.PlaintextModulus()
	if p <= uint64(numParties) {
		return fmt.Errorf("plaintext modulus %d must be > number of parties %d", p, numParties)
	}

	// Check gcd(p, Q) = 1 (Lattigo enforces this, but verify)
	pBig := new(big.Int).SetUint64(p)
	qBig := params.RingQ().ModulusAtLevel[params.MaxLevel()]
	g := new(big.Int).GCD(nil, nil, pBig, qBig)
	if g.Cmp(big.NewInt(1)) != 0 {
		return fmt.Errorf("gcd(p=%d, Q) = %s ≠ 1", p, g.String())
	}

	return nil
}
