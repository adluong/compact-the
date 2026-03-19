package scheme

import (
	"fmt"

	"compact-the/lsss"
	"compact-the/params"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
)

// Combine implements Algorithm 4: Combine(pp, S, c_0, {d_j}_{j∈S}) → b'.
// Computes b' = δ · c_0 + Σ_{j∈S} λ̂_j · d_j.
// The qualifying set S uses 0-based party indices.
// c_0 and d_j must be in coefficient form.
func Combine(pp *params.PublicParams, qualifyingSet []int, ct *rlwe.Ciphertext, partialDecs map[int]ring.Poly) (ring.Poly, int64, error) {
	if len(qualifyingSet) != pp.T {
		return ring.Poly{}, 0, fmt.Errorf("qualifying set size %d ≠ threshold %d", len(qualifyingSet), pp.T)
	}

	ringQ := pp.BGVParams.RingQ()
	level := ct.Level()
	ringQLvl := ringQ.AtLevel(level)

	// Extract and convert c_0 to coefficient form
	c0 := ringQ.NewPoly()
	c0.CopyLvl(level, ct.Value[0])
	if ct.IsNTT {
		ringQLvl.INTT(c0, c0)
	}

	// Extract submatrix M_S and compute determinant + cofactors
	MS := lsss.ExtractSubmatrix(pp.M, qualifyingSet)
	det, cofactors := lsss.FirstRowCofactors(MS, pp.T)
	if det == 0 {
		return ring.Poly{}, 0, fmt.Errorf("singular submatrix for set %v", qualifyingSet)
	}

	// Verify reconstruction identity
	if !lsss.VerifyReconstructionIdentity(MS, pp.T, det, cofactors) {
		return ring.Poly{}, 0, fmt.Errorf("reconstruction identity check failed for set %v", qualifyingSet)
	}

	// Compute b' = δ · c_0 + Σ λ̂_j · d_j
	bPrime := ringQ.NewPoly()
	tmp := ringQ.NewPoly()

	// Term 1: δ · c_0
	scalarMulSigned(ringQLvl, c0, det, bPrime)

	// Terms 2..t+1: Σ λ̂_j · d_j
	for idx, j := range qualifyingSet {
		dj, ok := partialDecs[j]
		if !ok {
			return ring.Poly{}, 0, fmt.Errorf("missing partial decryption for party %d", j)
		}
		scalarMulSigned(ringQLvl, dj, cofactors[idx], tmp)
		ringQLvl.Add(bPrime, tmp, bPrime)
	}

	return bPrime, det, nil
}
