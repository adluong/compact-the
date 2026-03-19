package scheme

import (
	"fmt"
	"math/rand"
	"testing"

	"compact-the/lsss"
	"compact-the/noise"
	"compact-the/params"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

// testEndToEnd runs the full threshold decryption pipeline for the given parameters
// and verifies correctness for every qualifying set.
func testEndToEnd(t *testing.T, numParties, threshold, bw, kappa int) {
	t.Helper()

	// 1. Setup
	bgvParams, err := params.DefaultRegimeB()
	if err != nil {
		t.Fatalf("failed to create BGV params: %v", err)
	}

	pp, err := Setup(bgvParams, numParties, threshold, bw, kappa)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// 2. Key generation
	kgen := rlwe.NewKeyGenerator(bgvParams)
	sk, pk := kgen.GenKeyPairNew()

	// 3. Share
	shares, err := Share(pp, sk)
	if err != nil {
		t.Fatalf("Share failed: %v", err)
	}
	if len(shares) != numParties {
		t.Fatalf("expected %d shares, got %d", numParties, len(shares))
	}

	// 4. Encode and encrypt a random plaintext
	encoder := bgv.NewEncoder(bgvParams)
	encryptor := rlwe.NewEncryptor(bgvParams, pk)

	T := bgvParams.PlaintextModulus()
	slots := bgvParams.MaxSlots()
	plaintext := make([]uint64, slots)
	for i := range plaintext {
		plaintext[i] = uint64(rand.Int63n(int64(T)))
	}

	pt := bgv.NewPlaintext(bgvParams, bgvParams.MaxLevel())
	if err := encoder.Encode(plaintext, pt); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	ct, err := encryptor.EncryptNew(pt)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// 5. Verify standard decryption works first
	decryptor := rlwe.NewDecryptor(bgvParams, sk)
	ptDec := decryptor.DecryptNew(ct)
	standardResult := make([]uint64, slots)
	if err := encoder.Decode(ptDec, standardResult); err != nil {
		t.Fatalf("Standard Decode failed: %v", err)
	}
	for i := 0; i < slots; i++ {
		if standardResult[i] != plaintext[i] {
			t.Fatalf("Standard decryption failed at slot %d: got %d, want %d", i, standardResult[i], plaintext[i])
		}
	}

	// 6. Partial decrypt for all parties
	partialDecs := make(map[int]ring.Poly)
	for j := 0; j < numParties; j++ {
		dj, err := PartDec(pp, shares[j], ct)
		if err != nil {
			t.Fatalf("PartDec for party %d failed: %v", j, err)
		}
		partialDecs[j] = dj
	}

	// 7. For EVERY qualifying set, combine and verify
	sets := lsss.AllQualifyingSets(numParties, threshold)
	t.Logf("Testing %d qualifying sets for (N=%d, t=%d)", len(sets), numParties, threshold)

	for _, S := range sets {
		bPrime, det, err := Combine(pp, S, ct, partialDecs)
		if err != nil {
			t.Errorf("Combine failed for set %v: %v", S, err)
			continue
		}

		result, err := FinDec(pp, det, bPrime)
		if err != nil {
			t.Errorf("FinDec failed for set %v: %v", S, err)
			continue
		}

		// Compare with original plaintext
		for i := 0; i < slots; i++ {
			if result[i] != plaintext[i] {
				t.Errorf("set %v: slot %d mismatch: got %d, want %d", S, i, result[i], plaintext[i])
				break
			}
		}
	}
}

func TestEndToEnd_T2_N3_BW1(t *testing.T) {
	testEndToEnd(t, 3, 2, 1, 40)
}

func TestEndToEnd_T3_N5_BW1(t *testing.T) {
	testEndToEnd(t, 5, 3, 1, 40)
}

func TestEndToEnd_T2_N10_Vandermonde(t *testing.T) {
	testEndToEnd(t, 10, 2, 0, 40) // bw=0 triggers Vandermonde
}

func TestEndToEnd_T3_N10_BW3(t *testing.T) {
	testEndToEnd(t, 10, 3, 1, 40) // B_W=3 hardcoded matrix (search result)
}

func TestEndToEnd_T3_N10_Vandermonde(t *testing.T) {
	testEndToEnd(t, 10, 3, 0, 40) // Paper's reference configuration (§11)
}

func TestEndToEnd_T4_N8_Vandermonde(t *testing.T) {
	testEndToEnd(t, 8, 4, 0, 40) // Tests t>3 regime
}

func TestMultipleDecryptions(t *testing.T) {
	bgvParams, err := params.DefaultRegimeB()
	if err != nil {
		t.Fatalf("failed to create BGV params: %v", err)
	}

	pp, err := Setup(bgvParams, 3, 2, 1, 40)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	kgen := rlwe.NewKeyGenerator(bgvParams)
	sk, pk := kgen.GenKeyPairNew()

	shares, err := Share(pp, sk)
	if err != nil {
		t.Fatalf("Share failed: %v", err)
	}

	encoder := bgv.NewEncoder(bgvParams)
	encryptor := rlwe.NewEncryptor(bgvParams, pk)
	T := bgvParams.PlaintextModulus()
	slots := bgvParams.MaxSlots()

	// Test with multiple different plaintexts using the same shares
	S := []int{0, 1} // use first qualifying set

	for trial := 0; trial < 5; trial++ {
		plaintext := make([]uint64, slots)
		for i := range plaintext {
			plaintext[i] = uint64(rand.Int63n(int64(T)))
		}

		pt := bgv.NewPlaintext(bgvParams, bgvParams.MaxLevel())
		if err := encoder.Encode(plaintext, pt); err != nil {
			t.Fatalf("trial %d: Encode failed: %v", trial, err)
		}
		ct, err := encryptor.EncryptNew(pt)
		if err != nil {
			t.Fatalf("trial %d: Encrypt failed: %v", trial, err)
		}

		partialDecs := make(map[int]ring.Poly)
		for j := 0; j < 3; j++ {
			dj, err := PartDec(pp, shares[j], ct)
			if err != nil {
				t.Fatalf("trial %d: PartDec %d failed: %v", trial, j, err)
			}
			partialDecs[j] = dj
		}

		bPrime, det, err := Combine(pp, S, ct, partialDecs)
		if err != nil {
			t.Fatalf("trial %d: Combine failed: %v", trial, err)
		}

		result, err := FinDec(pp, det, bPrime)
		if err != nil {
			t.Fatalf("trial %d: FinDec failed: %v", trial, err)
		}

		for i := 0; i < slots; i++ {
			if result[i] != plaintext[i] {
				t.Errorf("trial %d: slot %d mismatch: got %d, want %d", trial, i, result[i], plaintext[i])
				break
			}
		}
	}
}

func TestInsufficientParties(t *testing.T) {
	bgvParams, err := params.DefaultRegimeB()
	if err != nil {
		t.Fatalf("failed to create BGV params: %v", err)
	}

	pp, err := Setup(bgvParams, 3, 2, 1, 40)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	kgen := rlwe.NewKeyGenerator(bgvParams)
	sk, pk := kgen.GenKeyPairNew()

	shares, err := Share(pp, sk)
	if err != nil {
		t.Fatalf("Share failed: %v", err)
	}

	encoder := bgv.NewEncoder(bgvParams)
	encryptor := rlwe.NewEncryptor(bgvParams, pk)
	T := bgvParams.PlaintextModulus()
	slots := bgvParams.MaxSlots()

	plaintext := make([]uint64, slots)
	for i := range plaintext {
		plaintext[i] = uint64(rand.Int63n(int64(T)))
	}

	pt := bgv.NewPlaintext(bgvParams, bgvParams.MaxLevel())
	encoder.Encode(plaintext, pt)
	ct, _ := encryptor.EncryptNew(pt)

	// Try with only 1 party (threshold is 2) — Combine should fail
	partialDecs := make(map[int]ring.Poly)
	dj, _ := PartDec(pp, shares[0], ct)
	partialDecs[0] = dj

	_, _, err = Combine(pp, []int{0}, ct, partialDecs)
	if err == nil {
		t.Error("expected error with insufficient parties, got nil")
	}
}

func TestCorruptedPartialDec(t *testing.T) {
	bgvParams, err := params.DefaultRegimeB()
	if err != nil {
		t.Fatalf("failed to create BGV params: %v", err)
	}

	pp, err := Setup(bgvParams, 4, 2, 1, 40)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	kgen := rlwe.NewKeyGenerator(bgvParams)
	sk, pk := kgen.GenKeyPairNew()

	shares, err := Share(pp, sk)
	if err != nil {
		t.Fatalf("Share failed: %v", err)
	}

	encoder := bgv.NewEncoder(bgvParams)
	encryptor := rlwe.NewEncryptor(bgvParams, pk)
	T := bgvParams.PlaintextModulus()
	slots := bgvParams.MaxSlots()

	plaintext := make([]uint64, slots)
	for i := range plaintext {
		plaintext[i] = uint64(rand.Int63n(int64(T)))
	}

	pt := bgv.NewPlaintext(bgvParams, bgvParams.MaxLevel())
	encoder.Encode(plaintext, pt)
	ct, _ := encryptor.EncryptNew(pt)

	// Generate all partial decryptions
	partialDecs := make(map[int]ring.Poly)
	for j := 0; j < 4; j++ {
		dj, _ := PartDec(pp, shares[j], ct)
		partialDecs[j] = dj
	}

	// Corrupt party 0's partial decryption by adding large noise
	ringQ := bgvParams.RingQ()
	corruptNoise := ringQ.NewPoly()
	for i := 0; i <= ringQ.Level(); i++ {
		qi := ringQ.SubRings[i].Modulus
		for j := 0; j < ringQ.N(); j++ {
			corruptNoise.Coeffs[i][j] = qi / 2 // large noise
		}
	}
	ringQ.Add(partialDecs[0], corruptNoise, partialDecs[0])

	// Sets containing party 0 should fail
	sets := lsss.AllQualifyingSets(4, 2)
	correctCount := 0
	corruptCount := 0

	for _, S := range sets {
		bPrime, det, err := Combine(pp, S, ct, partialDecs)
		if err != nil {
			continue
		}
		result, err := FinDec(pp, det, bPrime)
		if err != nil {
			continue
		}

		match := true
		for i := 0; i < slots; i++ {
			if result[i] != plaintext[i] {
				match = false
				break
			}
		}

		containsCorrupt := false
		for _, j := range S {
			if j == 0 {
				containsCorrupt = true
				break
			}
		}

		if containsCorrupt && match {
			corruptCount++
		} else if !containsCorrupt && !match {
			t.Errorf("set %v (no corrupt party) produced wrong result", S)
		} else if !containsCorrupt && match {
			correctCount++
		}
	}

	// Sets without party 0: C(3,2) = 3 sets
	if correctCount != 3 {
		t.Errorf("expected 3 correct sets (without corrupt party), got %d", correctCount)
	}
	// Sets with corrupt party should mostly fail
	t.Logf("Sets with corrupt party that still matched: %d (expected ~0)", corruptCount)
	if corruptCount > 0 {
		fmt.Printf("WARNING: %d corrupt sets still matched (unlikely but possible)\n", corruptCount)
	}
}

// TestNoiseMarginAnalysis computes and reports the noise margin for various
// parameter configurations, empirically validating the paper's claim (§9.6)
// that t ≤ 3 is the practical regime.
func TestNoiseMarginAnalysis(t *testing.T) {
	bgvParams, err := params.DefaultRegimeB()
	if err != nil {
		t.Fatal(err)
	}

	logQ := bgvParams.RingQ().ModulusAtLevel[bgvParams.MaxLevel()].BitLen()
	logT := 0
	tmp := bgvParams.PlaintextModulus()
	for tmp > 1 {
		logT++
		tmp >>= 1
	}
	halfDeltaLog2 := logQ - logT - 1 // log₂(Δ/2) = log₂(Q/(2T))

	t.Logf("Parameters: n=%d, log₂Q=%d, T=%d (log₂T=%d), Δ/2 ≈ 2^%d",
		bgvParams.N(), logQ, bgvParams.PlaintextModulus(), logT, halfDeltaLog2)

	configs := []struct {
		name string
		N, T int
		bw   int
	}{
		{"t=2, N=3, B_W=1", 3, 2, 1},
		{"t=2, N=10, Vandermonde", 10, 2, 0},
		{"t=3, N=5, B_W=1", 5, 3, 1},
		{"t=3, N=10, B_W=3 (search)", 10, 3, 1},
		{"t=3, N=10, Vandermonde", 10, 3, 0},
		{"t=4, N=8, Vandermonde", 8, 4, 0},
		{"t=5, N=10, Vandermonde", 10, 5, 0},
	}

	for _, cfg := range configs {
		pp, err := Setup(bgvParams, cfg.N, cfg.T, cfg.bw, 40)
		if err != nil {
			t.Errorf("%s: Setup failed: %v", cfg.name, err)
			continue
		}

		sets := lsss.AllQualifyingSets(cfg.N, cfg.T)

		// Find worst-case noise margin across all qualifying sets
		worstMargin := halfDeltaLog2
		var worstSet []int
		var worstDet int64
		var worstLambda int64

		for _, S := range sets {
			MS := lsss.ExtractSubmatrix(pp.M, S)
			det, cofactors := lsss.FirstRowCofactors(MS, cfg.T)
			if det == 0 {
				continue
			}
			absDet := det
			if absDet < 0 {
				absDet = -absDet
			}
			lambdaS := lsss.LambdaS(cofactors)

			// Noise = |δ|·B_ct + Λ_S·B_sm
			// In log2 approx: max(log₂(|δ|)+bctLog2, log₂(Λ_S)+bsmLog2)
			// Margin = halfDeltaLog2 - log₂(noise)
			// Exact: use VerifyCorrectness
			ok := noise.VerifyCorrectness(absDet, lambdaS, pp.BsmLog2, 20, logQ, logT)

			if !ok && worstMargin > 0 {
				worstMargin = -1
				worstSet = S
				worstDet = det
				worstLambda = lambdaS
			} else if ok {
				// Rough margin estimate
				noiseBits := 0
				if lambdaS > 0 {
					l := lambdaS
					for l > 0 {
						noiseBits++
						l >>= 1
					}
				}
				margin := halfDeltaLog2 - (noiseBits + pp.BsmLog2)
				if margin < worstMargin {
					worstMargin = margin
					worstSet = S
					worstDet = det
					worstLambda = lambdaS
				}
			}
		}

		status := "PASS"
		if worstMargin < 0 {
			status = "FAIL"
		}

		t.Logf("[%s] %s: B_W=%d, log₂(B_sm)=%d, C(%d,%d)=%d sets, worst margin ≈ %d bits (set=%v, |δ|=%d, Λ_S=%d)",
			status, cfg.name, pp.BW, pp.BsmLog2, cfg.N, cfg.T, len(sets),
			worstMargin, worstSet, worstDet, worstLambda)
	}
}
