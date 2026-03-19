package lsss

import "testing"

func TestDeterminant2x2(t *testing.T) {
	mat := [][]int64{
		{1, 2},
		{3, 4},
	}
	got := Determinant(mat, 2)
	want := int64(1*4 - 2*3) // -2
	if got != want {
		t.Errorf("Determinant 2x2: got %d, want %d", got, want)
	}
}

func TestDeterminant3x3(t *testing.T) {
	mat := [][]int64{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 0},
	}
	got := Determinant(mat, 3)
	want := int64(1*(5*0-6*8) - 2*(4*0-6*7) + 3*(4*8-5*7)) // 27
	if got != want {
		t.Errorf("Determinant 3x3: got %d, want %d", got, want)
	}
}

func TestDeterminantIdentity(t *testing.T) {
	mat := [][]int64{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	}
	got := Determinant(mat, 3)
	if got != 1 {
		t.Errorf("Determinant identity: got %d, want 1", got)
	}
}

func TestFirstRowCofactors2x2(t *testing.T) {
	mat := [][]int64{
		{3, 1},
		{2, 5},
	}
	det, cofactors := FirstRowCofactors(mat, 2)
	wantDet := int64(3*5 - 1*2) // 13
	if det != wantDet {
		t.Errorf("det: got %d, want %d", det, wantDet)
	}
	// adj[0] = [d, -b] = [5, -1]
	wantCofactors := []int64{5, -1}
	for i, c := range cofactors {
		if c != wantCofactors[i] {
			t.Errorf("cofactor[%d]: got %d, want %d", i, c, wantCofactors[i])
		}
	}
}

func TestFirstRowCofactors3x3(t *testing.T) {
	mat := [][]int64{
		{1, 2, 3},
		{0, 1, 4},
		{5, 6, 0},
	}
	det, cofactors := FirstRowCofactors(mat, 3)

	// Verify reconstruction identity
	if !VerifyReconstructionIdentity(mat, 3, det, cofactors) {
		t.Errorf("reconstruction identity failed for det=%d, cofactors=%v", det, cofactors)
	}
}

func TestReconstructionIdentityAllSets(t *testing.T) {
	// Build a small share matrix and verify identity for all qualifying sets
	N, threshold := 5, 3
	W := VandermondeW(N, threshold)
	M := BuildM(W, N, threshold)
	sets := AllQualifyingSets(N, threshold)

	for _, S := range sets {
		MS := ExtractSubmatrix(M, S)
		det, cofactors := FirstRowCofactors(MS, threshold)
		if det == 0 {
			t.Errorf("zero determinant for set %v", S)
			continue
		}
		if !VerifyReconstructionIdentity(MS, threshold, det, cofactors) {
			t.Errorf("reconstruction identity failed for set %v: det=%d, cofactors=%v", S, det, cofactors)
		}
	}
}

func TestReconstructionIdentityBW1(t *testing.T) {
	// Test with B_W=1 matrices
	testCases := []struct {
		N, T int
	}{
		{3, 2},
		{4, 2},
		{5, 3},
	}

	for _, tc := range testCases {
		W, err := HardcodedBW1(tc.N, tc.T)
		if err != nil {
			t.Fatalf("N=%d, T=%d: %v", tc.N, tc.T, err)
		}
		M := BuildM(W, tc.N, tc.T)
		sets := AllQualifyingSets(tc.N, tc.T)

		for _, S := range sets {
			MS := ExtractSubmatrix(M, S)
			det, cofactors := FirstRowCofactors(MS, tc.T)
			if det == 0 {
				t.Errorf("N=%d, T=%d, set %v: zero determinant", tc.N, tc.T, S)
				continue
			}
			if !VerifyReconstructionIdentity(MS, tc.T, det, cofactors) {
				t.Errorf("N=%d, T=%d, set %v: reconstruction identity failed", tc.N, tc.T, S)
			}
		}
	}
}

func TestLambdaS(t *testing.T) {
	cofactors := []int64{5, -3, 2}
	got := LambdaS(cofactors)
	want := int64(10)
	if got != want {
		t.Errorf("LambdaS: got %d, want %d", got, want)
	}
}

func TestAllQualifyingSets(t *testing.T) {
	sets := AllQualifyingSets(5, 3)
	// C(5,3) = 10
	if len(sets) != 10 {
		t.Errorf("expected 10 sets, got %d", len(sets))
	}

	sets = AllQualifyingSets(10, 2)
	// C(10,2) = 45
	if len(sets) != 45 {
		t.Errorf("expected 45 sets, got %d", len(sets))
	}
}

func TestSearchBW1(t *testing.T) {
	testCases := []struct {
		N, T int
	}{
		{3, 2},
		{4, 2},
		{5, 3},
	}

	for _, tc := range testCases {
		W, err := SearchBW1(tc.N, tc.T)
		if err != nil {
			t.Errorf("N=%d, T=%d: search failed: %v", tc.N, tc.T, err)
			continue
		}
		// Verify all entries are in {-1, 0, 1}
		for j, row := range W {
			for k, v := range row {
				if v < -1 || v > 1 {
					t.Errorf("N=%d, T=%d: W[%d][%d]=%d out of {-1,0,1}", tc.N, tc.T, j, k, v)
				}
			}
		}
		// Verify all subsets have nonzero det
		M := BuildM(W, tc.N, tc.T)
		sets := AllQualifyingSets(tc.N, tc.T)
		for _, S := range sets {
			MS := ExtractSubmatrix(M, S)
			det := Determinant(MS, tc.T)
			if det == 0 {
				t.Errorf("N=%d, T=%d: zero det for set %v", tc.N, tc.T, S)
			}
		}
	}
}
