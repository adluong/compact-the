package bench

import (
	"math/rand"
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

func BenchmarkShare_T2_N3(b *testing.B) {
	setupBench(b, 3, 2, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Share(benchPP, benchSK)
	}
}

func BenchmarkShare_T3_N10(b *testing.B) {
	setupBench(b, 10, 3, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Share(benchPP, benchSK)
	}
}

func BenchmarkPartDec(b *testing.B) {
	setupBench(b, 3, 2, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.PartDec(benchPP, benchShares[0], benchCT)
	}
}

func BenchmarkCombine_T2(b *testing.B) {
	setupBench(b, 3, 2, 1)
	S := []int{0, 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Combine(benchPP, S, benchCT, benchPartialDecs)
	}
}

func BenchmarkCombine_T3(b *testing.B) {
	setupBench(b, 10, 3, 0)
	S := lsss.AllQualifyingSets(10, 3)[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Combine(benchPP, S, benchCT, benchPartialDecs)
	}
}

func BenchmarkFinDec(b *testing.B) {
	setupBench(b, 3, 2, 1)
	S := []int{0, 1}
	bPrime, det, _ := scheme.Combine(benchPP, S, benchCT, benchPartialDecs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.FinDec(benchPP, det, bPrime)
	}
}

// benchEndToEnd is a helper that benchmarks the full pipeline for a given config.
func benchEndToEnd(b *testing.B, numParties, threshold, bw int) {
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

// --- End-to-end benchmarks ---

func BenchmarkEndToEnd_T2_N3(b *testing.B) {
	benchEndToEnd(b, 3, 2, 1) // B_W=1
}

func BenchmarkEndToEnd_T3_N5(b *testing.B) {
	benchEndToEnd(b, 5, 3, 1) // B_W=1
}

func BenchmarkEndToEnd_T3_N10(b *testing.B) {
	benchEndToEnd(b, 10, 3, 0) // Vandermonde, paper's reference config (§11)
}

func BenchmarkEndToEnd_T4_N8(b *testing.B) {
	benchEndToEnd(b, 8, 4, 0) // Vandermonde, t>3 regime
}

func BenchmarkEndToEnd_T5_N10(b *testing.B) {
	benchEndToEnd(b, 10, 5, 0) // Vandermonde, stress test
}

// --- Per-algorithm benchmarks for t=4 ---

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

func BenchmarkCombine_T5_N10(b *testing.B) {
	setupBench(b, 10, 5, 0)
	S := lsss.AllQualifyingSets(10, 5)[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheme.Combine(benchPP, S, benchCT, benchPartialDecs)
	}
}
