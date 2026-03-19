package lsss

import (
	"fmt"
	"math/big"
)

// AllQualifyingSets returns all C(N, t) subsets of {0, ..., N-1} of size t.
// Subsets are returned in lexicographic order with 0-based indices.
func AllQualifyingSets(N, t int) [][]int {
	var result [][]int
	var generate func(start int, current []int)
	generate = func(start int, current []int) {
		if len(current) == t {
			s := make([]int, t)
			copy(s, current)
			result = append(result, s)
			return
		}
		remaining := t - len(current)
		for i := start; i <= N-remaining; i++ {
			generate(i+1, append(current, i))
		}
	}
	generate(0, nil)
	return result
}

// ExtractSubmatrix extracts the rows of M indexed by the elements of rows.
// Returns a len(rows) × cols submatrix.
func ExtractSubmatrix(M [][]int64, rows []int) [][]int64 {
	sub := make([][]int64, len(rows))
	for i, r := range rows {
		sub[i] = make([]int64, len(M[r]))
		copy(sub[i], M[r])
	}
	return sub
}

// ValidateAllSets checks that every t-subset S of [N] yields a non-singular
// submatrix M_S, and that det(M_S) is coprime to each RNS prime.
// qModuli are the RNS primes q_1, ..., q_L.
func ValidateAllSets(M [][]int64, t int, sets [][]int, qModuli []uint64) error {
	for _, S := range sets {
		MS := ExtractSubmatrix(M, S)
		det := Determinant(MS, t)
		if det == 0 {
			return fmt.Errorf("singular submatrix for set %v: det = 0", S)
		}
		// Check gcd(det, q_i) = 1 for each RNS prime
		detBig := big.NewInt(det)
		if det < 0 {
			detBig.Neg(detBig)
		}
		for _, qi := range qModuli {
			g := new(big.Int).GCD(nil, nil, detBig, new(big.Int).SetUint64(qi))
			if g.Cmp(big.NewInt(1)) != 0 {
				return fmt.Errorf("det(M_S) = %d not coprime to prime %d for set %v", det, qi, S)
			}
		}
	}
	return nil
}

// PrecomputeCofactors computes and caches the determinant and first-row cofactors
// for every qualifying set. Returns maps from set index to (det, cofactors).
func PrecomputeCofactors(M [][]int64, t int, sets [][]int) ([]int64, [][]int64) {
	dets := make([]int64, len(sets))
	allCofactors := make([][]int64, len(sets))
	for i, S := range sets {
		MS := ExtractSubmatrix(M, S)
		det, cofactors := FirstRowCofactors(MS, t)
		dets[i] = det
		allCofactors[i] = cofactors
	}
	return dets, allCofactors
}
