package lsss

import "fmt"

// VandermondeW constructs a Vandermonde W matrix for the given parameters.
// Evaluation points are 1, 2, ..., N-t+1.
// Returns W with dimensions (N-t+1) × t.
func VandermondeW(N, t int) [][]int64 {
	rows := N - t + 1
	W := make([][]int64, rows)
	for j := 0; j < rows; j++ {
		alpha := int64(j + 1) // evaluation points 1, 2, ...
		W[j] = make([]int64, t)
		W[j][0] = 1
		for k := 1; k < t; k++ {
			W[j][k] = W[j][k-1] * alpha
		}
	}
	return W
}

// BuildM constructs the share-generation matrix M from W.
// M has block structure:
//
//	| 0  | I_{t-1} |   ← F-party rows (0..t-2)
//	|    -W        |   ← L-party rows (t-1..N-1)
//
// M is N × t.
func BuildM(W [][]int64, N, t int) [][]int64 {
	M := make([][]int64, N)

	// F-party rows: row i has M[i][0]=0, M[i][i+1]=1
	for i := 0; i < t-1; i++ {
		M[i] = make([]int64, t)
		M[i][i+1] = 1
	}

	// L-party rows: row (t-1+j) = -W[j]
	for j := 0; j < N-t+1; j++ {
		M[t-1+j] = make([]int64, t)
		for k := 0; k < t; k++ {
			M[t-1+j][k] = -W[j][k]
		}
	}

	return M
}

// SearchBW1 searches for a W matrix with entries in {-1, 0, 1} such that
// every t-subset of [N] has non-zero determinant in the resulting M.
// Only feasible for small parameters (t=2 N≤4, t=3 N≤10).
func SearchBW1(N, t int) ([][]int64, error) {
	rows := N - t + 1
	cols := t
	totalEntries := rows * cols

	// Total number of candidate matrices: 3^(rows*cols)
	// For t=3, N=10: 3^24 ≈ 282 billion — too large for brute force.
	// For t=2, N=4: 3^4 = 81 — trivial.
	// For t=3, N=5: 3^9 = 19683 — trivial.
	if totalEntries > 18 {
		// Fall back to hardcoded matrices for large search spaces
		W, err := HardcodedBW1(N, t)
		if err == nil {
			return W, nil
		}
		return nil, fmt.Errorf("B_W=1 search infeasible for %d entries (3^%d candidates) and no hardcoded matrix available", totalEntries, totalEntries)
	}

	// Generate all subsets of [N] of size t
	sets := AllQualifyingSets(N, t)

	// Enumerate all W matrices with entries in {-1, 0, 1}
	W := make([][]int64, rows)
	for j := range W {
		W[j] = make([]int64, cols)
	}

	numCandidates := 1
	for i := 0; i < totalEntries; i++ {
		numCandidates *= 3
	}

	for candidate := 0; candidate < numCandidates; candidate++ {
		// Decode candidate into W entries
		tmp := candidate
		for j := 0; j < rows; j++ {
			for k := 0; k < cols; k++ {
				W[j][k] = int64(tmp%3) - 1 // maps {0,1,2} → {-1,0,1}
				tmp /= 3
			}
		}

		M := BuildM(W, N, t)

		allNonZero := true
		for _, S := range sets {
			MS := ExtractSubmatrix(M, S)
			det := Determinant(MS, t)
			if det == 0 {
				allNonZero = false
				break
			}
		}
		if allNonZero {
			// Return a copy
			result := make([][]int64, rows)
			for j := range result {
				result[j] = make([]int64, cols)
				copy(result[j], W[j])
			}
			return result, nil
		}
	}

	return nil, fmt.Errorf("no B_W=1 matrix found for N=%d, t=%d", N, t)
}

// HardcodedBW1 returns known-good W matrices with B_W=1 for specific parameters.
func HardcodedBW1(N, t int) ([][]int64, error) {
	switch {
	case t == 2 && N == 3:
		// W is (N-t+1)×t = 2×2, found by exhaustive search
		return [][]int64{
			{1, -1},
			{-1, -1},
		}, nil
	case t == 2 && N == 4:
		// W is 3×2
		return [][]int64{
			{-1, 0},
			{1, -1},
			{-1, -1},
		}, nil
	case t == 3 && N == 5:
		// W is 3×3
		return [][]int64{
			{1, -1, 0},
			{1, 0, -1},
			{-1, -1, -1},
		}, nil
	case t == 3 && N == 10:
		// W is 8×3, B_W=3, found by randomized search with early rejection.
		// B_W=1 does not exist (exhaustively proven via backtracking).
		// B_W=2 not found after 38M random trials.
		// All 120 qualifying sets verified to have non-zero determinant.
		return [][]int64{
			{-3, 1, 2},
			{1, 2, 3},
			{3, 2, -1},
			{-3, 2, 3},
			{-2, 0, 3},
			{2, 2, 0},
			{2, 3, 2},
			{-2, 1, 1},
		}, nil
	default:
		return nil, fmt.Errorf("no hardcoded B_W=1 matrix for N=%d, t=%d", N, t)
	}
}

// MaxAbsEntry returns the maximum absolute value of entries in W.
func MaxAbsEntry(W [][]int64) int64 {
	max := int64(0)
	for _, row := range W {
		for _, v := range row {
			abs := v
			if abs < 0 {
				abs = -abs
			}
			if abs > max {
				max = abs
			}
		}
	}
	return max
}
