// Package lsss implements the linear secret sharing scheme (LSSS)
// matrix construction, adjugate computation, and qualifying set management
// for the CDM00-based threshold homomorphic encryption scheme.
package lsss

import "math/big"

// Minor returns the (row, col) minor of mat — the matrix with the given
// row and column deleted. mat must be size×size.
func Minor(mat [][]int64, row, col, size int) [][]int64 {
	m := make([][]int64, size-1)
	ri := 0
	for i := 0; i < size; i++ {
		if i == row {
			continue
		}
		m[ri] = make([]int64, size-1)
		ci := 0
		for j := 0; j < size; j++ {
			if j == col {
				continue
			}
			m[ri][ci] = mat[i][j]
			ci++
		}
		ri++
	}
	return m
}

// Determinant computes the determinant of a size×size integer matrix
// using cofactor expansion. For t ≤ 4 this is efficient.
func Determinant(mat [][]int64, size int) int64 {
	switch size {
	case 1:
		return mat[0][0]
	case 2:
		return mat[0][0]*mat[1][1] - mat[0][1]*mat[1][0]
	case 3:
		a, b, c := mat[0][0], mat[0][1], mat[0][2]
		d, e, f := mat[1][0], mat[1][1], mat[1][2]
		g, h, i := mat[2][0], mat[2][1], mat[2][2]
		return a*(e*i-f*h) - b*(d*i-f*g) + c*(d*h-e*g)
	default:
		det := int64(0)
		sign := int64(1)
		for j := 0; j < size; j++ {
			minor := Minor(mat, 0, j, size)
			det += sign * mat[0][j] * Determinant(minor, size-1)
			sign = -sign
		}
		return det
	}
}

// DeterminantBig computes the determinant using arbitrary precision.
// Use this when t >= 5 or B_W is large.
func DeterminantBig(mat [][]int64, size int) *big.Int {
	if size == 1 {
		return big.NewInt(mat[0][0])
	}
	det := new(big.Int)
	sign := int64(1)
	for j := 0; j < size; j++ {
		minor := Minor(mat, 0, j, size)
		cofactor := DeterminantBig(minor, size-1)
		cofactor.Mul(cofactor, big.NewInt(sign*mat[0][j]))
		det.Add(det, cofactor)
		sign = -sign
	}
	return det
}

// Adjugate computes the classical adjugate (transpose of cofactor matrix)
// of a size×size integer matrix.
func Adjugate(mat [][]int64, size int) [][]int64 {
	adj := make([][]int64, size)
	for i := 0; i < size; i++ {
		adj[i] = make([]int64, size)
		for j := 0; j < size; j++ {
			minor := Minor(mat, j, i, size) // note: transposed indices
			sign := int64(1)
			if (i+j)%2 != 0 {
				sign = -1
			}
			adj[i][j] = sign * Determinant(minor, size-1)
		}
	}
	return adj
}

// FirstRowCofactors computes the determinant and the first-row cofactors
// of the adjugate matrix. These are the reconstruction coefficients λ̂_j.
// Returns (det, cofactors) where cofactors[j] = adj(mat)[0][j].
func FirstRowCofactors(mat [][]int64, size int) (int64, []int64) {
	cofactors := make([]int64, size)
	for j := 0; j < size; j++ {
		minor := Minor(mat, j, 0, size) // adj[0][j] = (-1)^(0+j) * det(minor(j, 0))
		sign := int64(1)
		if j%2 != 0 {
			sign = -1
		}
		cofactors[j] = sign * Determinant(minor, size-1)
	}

	// Compute det via expansion: det = Σ_k mat[k][0] * cofactors[k]
	// Actually, adj(M)*M = det*I, so first row of adj times first column of M = det.
	// But cofactors here are adj[0][j], and det = Σ_j mat[j][0] * adj[0][j]
	// Wait: adj(M)[0][j] * M[j][k] summed over j gives det*δ_{0,k}.
	// So det = Σ_j adj[0][j] * M[j][0] = Σ_j cofactors[j] * mat[j][0].
	det := int64(0)
	for j := 0; j < size; j++ {
		det += cofactors[j] * mat[j][0]
	}

	return det, cofactors
}

// VerifyReconstructionIdentity checks that Σ_j cofactors[j] * mat[j][k] = det if k==0, 0 otherwise.
// Returns true if the identity holds.
func VerifyReconstructionIdentity(mat [][]int64, size int, det int64, cofactors []int64) bool {
	for k := 0; k < size; k++ {
		sum := int64(0)
		for j := 0; j < size; j++ {
			sum += cofactors[j] * mat[j][k]
		}
		if k == 0 && sum != det {
			return false
		}
		if k != 0 && sum != 0 {
			return false
		}
	}
	return true
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
