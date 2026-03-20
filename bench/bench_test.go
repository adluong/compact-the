package bench

import (
	"math/rand"
	"sync"
	"testing"

	"compact-the/lsss"
	"compact-the/params"
	"compact-the/scheme"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

var benchPP *params.PublicParams
var benchSK *rlwe.SecretKey
var benchPK *rlwe.PublicKey
var benchCT *rlwe.Ciphertext
var benchShares []ring.Poly
var benchPartialDecs map[int]ring.Poly

func setupBench(b *testing.B, numParties, threshold, bw int) {
	b.Helper()
	bgvParams, _ := params.DefaultRegimeB()
	pp, err := scheme.Setup(bgvParams, numParties, threshold, bw, 40)
	if err != nil {
		b.Fatal(err)
	}
	benchPP = pp

	kgen := rlwe.NewKeyGenerator(bgvParams)
	benchSK, benchPK = kgen.GenKeyPairNew()

	encoder := bgv.NewEncoder(bgvParams)
	encryptor := rlwe.NewEncryptor(bgvParams, benchPK)
	T := bgvParams.PlaintextModulus()
	slots := bgvParams.MaxSlots()

	plaintext := make([]uint64, slots)
	for i := range plaintext {
		plaintext[i] = uint64(rand.Int63n(int64(T)))
	}
	pt := bgv.NewPlaintext(bgvParams, bgvParams.MaxLevel())
	encoder.Encode(plaintext, pt)
	benchCT, _ = encryptor.EncryptNew(pt)

	benchShares, _ = scheme.Share(pp, benchSK)

	benchPartialDecs = make(map[int]ring.Poly)
	for j := 0; j < numParties; j++ {
		dj, _ := scheme.PartDec(pp, benchShares[j], benchCT)
		benchPartialDecs[j] = dj
	}
}

// =====================================================================
// Per-algorithm benchmarks: t=2, N=3, B_W=1
// =====================================================================

func BenchmarkShare_T2_N3(b *testing.B) {
	setupBench(b, 3, 2, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Share(benchPP, benchSK)
	}
}

func BenchmarkPartDec_T2_N3(b *testing.B) {
	setupBench(b, 3, 2, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.PartDec(benchPP, benchShares[0], benchCT)
	}
}

func BenchmarkCombine_T2_N3(b *testing.B) {
	setupBench(b, 3, 2, 1)
	S := []int{0, 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Combine(benchPP, S, benchCT, benchPartialDecs)
	}
}

func BenchmarkFinDec_T2_N3(b *testing.B) {
	setupBench(b, 3, 2, 1)
	S := []int{0, 1}
	bPrime, det, _ := scheme.Combine(benchPP, S, benchCT, benchPartialDecs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.FinDec(benchPP, det, bPrime)
	}
}

// =====================================================================
// Per-algorithm benchmarks: t=3, N=10, B_W=3 (search matrix)
// =====================================================================

func BenchmarkShare_T3_N10_BW3(b *testing.B) {
	setupBench(b, 10, 3, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Share(benchPP, benchSK)
	}
}

func BenchmarkPartDec_T3_N10_BW3(b *testing.B) {
	setupBench(b, 10, 3, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.PartDec(benchPP, benchShares[0], benchCT)
	}
}

func BenchmarkCombine_T3_N10_BW3(b *testing.B) {
	setupBench(b, 10, 3, 1)
	S := lsss.AllQualifyingSets(10, 3)[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Combine(benchPP, S, benchCT, benchPartialDecs)
	}
}

func BenchmarkFinDec_T3_N10_BW3(b *testing.B) {
	setupBench(b, 10, 3, 1)
	S := lsss.AllQualifyingSets(10, 3)[0]
	bPrime, det, _ := scheme.Combine(benchPP, S, benchCT, benchPartialDecs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.FinDec(benchPP, det, bPrime)
	}
}

// =====================================================================
// Per-algorithm benchmarks: t=3, N=10, Vandermonde (B_W=64)
// =====================================================================

func BenchmarkShare_T3_N10_Vand(b *testing.B) {
	setupBench(b, 10, 3, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Share(benchPP, benchSK)
	}
}

func BenchmarkPartDec_T3_N10_Vand(b *testing.B) {
	setupBench(b, 10, 3, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.PartDec(benchPP, benchShares[0], benchCT)
	}
}

func BenchmarkCombine_T3_N10_Vand(b *testing.B) {
	setupBench(b, 10, 3, 0)
	S := lsss.AllQualifyingSets(10, 3)[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Combine(benchPP, S, benchCT, benchPartialDecs)
	}
}

func BenchmarkFinDec_T3_N10_Vand(b *testing.B) {
	setupBench(b, 10, 3, 0)
	S := lsss.AllQualifyingSets(10, 3)[0]
	bPrime, det, _ := scheme.Combine(benchPP, S, benchCT, benchPartialDecs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.FinDec(benchPP, det, bPrime)
	}
}

// =====================================================================
// Per-algorithm benchmarks: t=4, N=8, Vandermonde
// =====================================================================

func BenchmarkShare_T4_N8(b *testing.B) {
	setupBench(b, 8, 4, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Share(benchPP, benchSK)
	}
}

func BenchmarkPartDec_T4_N8(b *testing.B) {
	setupBench(b, 8, 4, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.PartDec(benchPP, benchShares[0], benchCT)
	}
}

func BenchmarkCombine_T4_N8(b *testing.B) {
	setupBench(b, 8, 4, 0)
	S := lsss.AllQualifyingSets(8, 4)[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Combine(benchPP, S, benchCT, benchPartialDecs)
	}
}

func BenchmarkFinDec_T4_N8(b *testing.B) {
	setupBench(b, 8, 4, 0)
	S := lsss.AllQualifyingSets(8, 4)[0]
	bPrime, det, _ := scheme.Combine(benchPP, S, benchCT, benchPartialDecs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.FinDec(benchPP, det, bPrime)
	}
}

// =====================================================================
// Per-algorithm benchmarks: t=5, N=10, Vandermonde (B3 fix)
// =====================================================================

func BenchmarkShare_T5_N10(b *testing.B) {
	setupBench(b, 10, 5, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Share(benchPP, benchSK)
	}
}

func BenchmarkPartDec_T5_N10(b *testing.B) {
	setupBench(b, 10, 5, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.PartDec(benchPP, benchShares[0], benchCT)
	}
}

func BenchmarkCombine_T5_N10(b *testing.B) {
	setupBench(b, 10, 5, 0)
	S := lsss.AllQualifyingSets(10, 5)[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Combine(benchPP, S, benchCT, benchPartialDecs)
	}
}

func BenchmarkFinDec_T5_N10(b *testing.B) {
	setupBench(b, 10, 5, 0)
	S := lsss.AllQualifyingSets(10, 5)[0]
	bPrime, det, _ := scheme.Combine(benchPP, S, benchCT, benchPartialDecs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.FinDec(benchPP, det, bPrime)
	}
}

// =====================================================================
// PartDec decomposition benchmarks
// =====================================================================

func BenchmarkPartDec_RingMul(b *testing.B) {
	setupBench(b, 10, 3, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.PartDecRingMul(benchPP, benchShares[0], benchCT)
	}
}

func BenchmarkPartDec_SmudgingNoise(b *testing.B) {
	setupBench(b, 10, 3, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.PartDecSmudgingNoise(benchPP, benchCT)
	}
}

func BenchmarkPartDec_SmudgingNoise_CSPRNG(b *testing.B) {
	setupBench(b, 10, 3, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.PartDecSmudgingNoiseFast(benchPP, benchCT)
	}
}

// =====================================================================
// Encrypt benchmark (explains E2E overhead vs sum-of-parts)
// =====================================================================

func BenchmarkEncrypt(b *testing.B) {
	bgvParams, _ := params.DefaultRegimeB()
	kgen := rlwe.NewKeyGenerator(bgvParams)
	_, pk := kgen.GenKeyPairNew()
	encoder := bgv.NewEncoder(bgvParams)
	encryptor := rlwe.NewEncryptor(bgvParams, pk)

	T := bgvParams.PlaintextModulus()
	slots := bgvParams.MaxSlots()
	plaintext := make([]uint64, slots)
	for i := range plaintext {
		plaintext[i] = uint64(rand.Int63n(int64(T)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pt := bgv.NewPlaintext(bgvParams, bgvParams.MaxLevel())
		encoder.Encode(plaintext, pt)
		encryptor.EncryptNew(pt)
	}
}

// =====================================================================
// Setup/KeyGen benchmarks (G3 fix)
// =====================================================================

func BenchmarkSetup_T3_N10_BW3(b *testing.B) {
	bgvParams, _ := params.DefaultRegimeB()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Setup(bgvParams, 10, 3, 1, 40)
	}
}

func BenchmarkSetup_T3_N10_Vand(b *testing.B) {
	bgvParams, _ := params.DefaultRegimeB()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Setup(bgvParams, 10, 3, 0, 40)
	}
}

func BenchmarkKeyGen(b *testing.B) {
	bgvParams, _ := params.DefaultRegimeB()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kgen := rlwe.NewKeyGenerator(bgvParams)
		kgen.GenKeyPairNew()
	}
}

// =====================================================================
// Single-party BFV decrypt baseline (G4 fix)
// =====================================================================

func BenchmarkBFVDecrypt(b *testing.B) {
	bgvParams, _ := params.DefaultRegimeB()
	kgen := rlwe.NewKeyGenerator(bgvParams)
	sk, pk := kgen.GenKeyPairNew()
	encoder := bgv.NewEncoder(bgvParams)
	encryptor := rlwe.NewEncryptor(bgvParams, pk)
	decryptor := rlwe.NewDecryptor(bgvParams, sk)

	T := bgvParams.PlaintextModulus()
	slots := bgvParams.MaxSlots()
	plaintext := make([]uint64, slots)
	for i := range plaintext {
		plaintext[i] = uint64(rand.Int63n(int64(T)))
	}

	pt := bgv.NewPlaintext(bgvParams, bgvParams.MaxLevel())
	encoder.Encode(plaintext, pt)
	ct, _ := encryptor.EncryptNew(pt)

	result := make([]uint64, slots)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ptDec := decryptor.DecryptNew(ct)
		encoder.Decode(ptDec, result)
	}
}

// =====================================================================
// Threshold decryption benchmarks (primary metric — no Encrypt in loop)
// Times: Share + N×PartDec + Combine + FinDec
// =====================================================================

func benchThresholdDecrypt(b *testing.B, numParties, threshold, bw int) {
	b.Helper()
	bgvParams, _ := params.DefaultRegimeB()
	pp, err := scheme.Setup(bgvParams, numParties, threshold, bw, 40)
	if err != nil {
		b.Fatal(err)
	}
	kgen := rlwe.NewKeyGenerator(bgvParams)
	sk, pk := kgen.GenKeyPairNew()
	encoder := bgv.NewEncoder(bgvParams)
	encryptor := rlwe.NewEncryptor(bgvParams, pk)

	T := bgvParams.PlaintextModulus()
	slots := bgvParams.MaxSlots()
	plaintext := make([]uint64, slots)
	for i := range plaintext {
		plaintext[i] = uint64(rand.Int63n(int64(T)))
	}

	// Encode+Encrypt ONCE, outside timing loop
	pt := bgv.NewPlaintext(bgvParams, bgvParams.MaxLevel())
	encoder.Encode(plaintext, pt)
	ct, _ := encryptor.EncryptNew(pt)

	S := lsss.AllQualifyingSets(numParties, threshold)[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shares, _ := scheme.Share(pp, sk)

		partialDecs := make(map[int]ring.Poly)
		for j := 0; j < numParties; j++ {
			dj, _ := scheme.PartDec(pp, shares[j], ct)
			partialDecs[j] = dj
		}

		bPrime, det, _ := scheme.Combine(pp, S, ct, partialDecs)
		scheme.FinDec(pp, det, bPrime)
	}
}

func BenchmarkThresholdDecrypt_T2_N3(b *testing.B)        { benchThresholdDecrypt(b, 3, 2, 1) }
func BenchmarkThresholdDecrypt_T2_N4(b *testing.B)        { benchThresholdDecrypt(b, 4, 2, 1) }
func BenchmarkThresholdDecrypt_T3_N5(b *testing.B)        { benchThresholdDecrypt(b, 5, 3, 1) }
func BenchmarkThresholdDecrypt_T3_N10_BW3(b *testing.B)   { benchThresholdDecrypt(b, 10, 3, 1) }
func BenchmarkThresholdDecrypt_T3_N10_Vand(b *testing.B)  { benchThresholdDecrypt(b, 10, 3, 0) }
func BenchmarkThresholdDecrypt_T4_N8(b *testing.B)        { benchThresholdDecrypt(b, 8, 4, 0) }
func BenchmarkThresholdDecrypt_T5_N10(b *testing.B)       { benchThresholdDecrypt(b, 10, 5, 0) }

// =====================================================================
// Full pipeline benchmarks (includes Encode+Encrypt — for completeness)
// =====================================================================

func benchFullPipeline(b *testing.B, numParties, threshold, bw int) {
	b.Helper()
	bgvParams, _ := params.DefaultRegimeB()
	pp, err := scheme.Setup(bgvParams, numParties, threshold, bw, 40)
	if err != nil {
		b.Fatal(err)
	}
	kgen := rlwe.NewKeyGenerator(bgvParams)
	sk, pk := kgen.GenKeyPairNew()
	encoder := bgv.NewEncoder(bgvParams)
	encryptor := rlwe.NewEncryptor(bgvParams, pk)

	T := bgvParams.PlaintextModulus()
	slots := bgvParams.MaxSlots()
	plaintext := make([]uint64, slots)
	for i := range plaintext {
		plaintext[i] = uint64(rand.Int63n(int64(T)))
	}

	S := lsss.AllQualifyingSets(numParties, threshold)[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shares, _ := scheme.Share(pp, sk)

		pt := bgv.NewPlaintext(bgvParams, bgvParams.MaxLevel())
		encoder.Encode(plaintext, pt)
		ct, _ := encryptor.EncryptNew(pt)

		partialDecs := make(map[int]ring.Poly)
		for j := 0; j < numParties; j++ {
			dj, _ := scheme.PartDec(pp, shares[j], ct)
			partialDecs[j] = dj
		}

		bPrime, det, _ := scheme.Combine(pp, S, ct, partialDecs)
		scheme.FinDec(pp, det, bPrime)
	}
}

func BenchmarkFullPipeline_T2_N3(b *testing.B)        { benchFullPipeline(b, 3, 2, 1) }
func BenchmarkFullPipeline_T3_N5(b *testing.B)        { benchFullPipeline(b, 5, 3, 1) }
func BenchmarkFullPipeline_T3_N10_BW3(b *testing.B)   { benchFullPipeline(b, 10, 3, 1) }
func BenchmarkFullPipeline_T3_N10_Vand(b *testing.B)  { benchFullPipeline(b, 10, 3, 0) }
func BenchmarkFullPipeline_T4_N8(b *testing.B)        { benchFullPipeline(b, 8, 4, 0) }
func BenchmarkFullPipeline_T5_N10(b *testing.B)       { benchFullPipeline(b, 10, 5, 0) }

// =====================================================================
// Threshold decryption parallel: PartDec runs concurrently
// =====================================================================

func BenchmarkThresholdDecrypt_T3_N10_BW3_Parallel(b *testing.B) {
	bgvParams, _ := params.DefaultRegimeB()
	pp, err := scheme.Setup(bgvParams, 10, 3, 1, 40)
	if err != nil {
		b.Fatal(err)
	}
	kgen := rlwe.NewKeyGenerator(bgvParams)
	sk, pk := kgen.GenKeyPairNew()
	encoder := bgv.NewEncoder(bgvParams)
	encryptor := rlwe.NewEncryptor(bgvParams, pk)

	T := bgvParams.PlaintextModulus()
	slots := bgvParams.MaxSlots()
	plaintext := make([]uint64, slots)
	for i := range plaintext {
		plaintext[i] = uint64(rand.Int63n(int64(T)))
	}

	// Encode+Encrypt ONCE, outside timing loop
	pt := bgv.NewPlaintext(bgvParams, bgvParams.MaxLevel())
	encoder.Encode(plaintext, pt)
	ct, _ := encryptor.EncryptNew(pt)

	S := lsss.AllQualifyingSets(10, 3)[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shares, _ := scheme.Share(pp, sk)

		partialDecs := make(map[int]ring.Poly)
		var mu sync.Mutex
		var wg sync.WaitGroup
		for j := 0; j < 10; j++ {
			wg.Add(1)
			go func(partyIdx int) {
				defer wg.Done()
				dj, _ := scheme.PartDec(pp, shares[partyIdx], ct)
				mu.Lock()
				partialDecs[partyIdx] = dj
				mu.Unlock()
			}(j)
		}
		wg.Wait()

		bPrime, det, _ := scheme.Combine(pp, S, ct, partialDecs)
		scheme.FinDec(pp, det, bPrime)
	}
}

// =====================================================================
// Setup benchmarks (one-time costs)
// =====================================================================

func BenchmarkSetup_T2_N3(b *testing.B) {
	bgvParams, _ := params.DefaultRegimeB()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Setup(bgvParams, 3, 2, 1, 40)
	}
}

func BenchmarkSetup_T4_N8(b *testing.B) {
	bgvParams, _ := params.DefaultRegimeB()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Setup(bgvParams, 8, 4, 0, 40)
	}
}

func BenchmarkSetup_T5_N10(b *testing.B) {
	bgvParams, _ := params.DefaultRegimeB()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Setup(bgvParams, 10, 5, 0, 40)
	}
}
