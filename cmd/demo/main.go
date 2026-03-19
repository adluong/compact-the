// Demo program for the compact threshold homomorphic encryption scheme.
// Runs the full pipeline: Setup → KeyGen → Share → Encrypt → PartDec → Combine → FinDec.
package main

import (
	"fmt"
	"math/rand"
	"time"

	"compact-the/lsss"
	"compact-the/params"
	"compact-the/scheme"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

func main() {
	N := 5  // parties
	t := 3  // threshold
	bw := 1 // B_W bound
	kappa := 40

	fmt.Printf("=== Compact THE Demo ===\n")
	fmt.Printf("Parties: %d, Threshold: %d, B_W: %d, κ: %d\n\n", N, t, bw, kappa)

	// Setup
	start := time.Now()
	bgvParams, err := params.DefaultRegimeB()
	if err != nil {
		panic(err)
	}
	pp, err := scheme.Setup(bgvParams, N, t, bw, kappa)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Setup:    %v\n", time.Since(start))
	fmt.Printf("  Ring degree: n = %d\n", bgvParams.N())
	fmt.Printf("  log₂ Q ≈ %d\n", bgvParams.RingQ().ModulusAtLevel[bgvParams.MaxLevel()].BitLen())
	fmt.Printf("  Plaintext modulus: T = %d\n", bgvParams.PlaintextModulus())
	fmt.Printf("  Smudging noise: log₂ B_sm = %d\n", pp.BsmLog2)
	fmt.Printf("  Qualifying sets: C(%d,%d) = %d\n\n", N, t, len(lsss.AllQualifyingSets(N, t)))

	// KeyGen
	start = time.Now()
	kgen := rlwe.NewKeyGenerator(bgvParams)
	sk, pk := kgen.GenKeyPairNew()
	fmt.Printf("KeyGen:   %v\n", time.Since(start))

	// Share
	start = time.Now()
	shares, err := scheme.Share(pp, sk)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Share:    %v\n", time.Since(start))

	// Encrypt
	encoder := bgv.NewEncoder(bgvParams)
	encryptor := rlwe.NewEncryptor(bgvParams, pk)
	T := bgvParams.PlaintextModulus()
	slots := bgvParams.MaxSlots()

	plaintext := make([]uint64, slots)
	for i := range plaintext {
		plaintext[i] = uint64(rand.Int63n(int64(T)))
	}

	start = time.Now()
	pt := bgv.NewPlaintext(bgvParams, bgvParams.MaxLevel())
	encoder.Encode(plaintext, pt)
	ct, err := encryptor.EncryptNew(pt)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Encrypt:  %v\n", time.Since(start))

	// PartDec for all parties
	start = time.Now()
	partialDecs := make(map[int]ring.Poly)
	for j := 0; j < N; j++ {
		dj, err := scheme.PartDec(pp, shares[j], ct)
		if err != nil {
			panic(err)
		}
		partialDecs[j] = dj
	}
	fmt.Printf("PartDec:  %v (all %d parties)\n", time.Since(start), N)

	// Combine + FinDec for all qualifying sets
	sets := lsss.AllQualifyingSets(N, t)
	successes := 0
	var combineTotal, findecTotal time.Duration

	for _, S := range sets {
		start = time.Now()
		bPrime, det, err := scheme.Combine(pp, S, ct, partialDecs)
		combineTotal += time.Since(start)
		if err != nil {
			fmt.Printf("  Combine FAILED for set %v: %v\n", S, err)
			continue
		}

		start = time.Now()
		result, err := scheme.FinDec(pp, det, bPrime)
		findecTotal += time.Since(start)
		if err != nil {
			fmt.Printf("  FinDec FAILED for set %v: %v\n", S, err)
			continue
		}

		// Verify
		correct := true
		for i := 0; i < slots; i++ {
			if result[i] != plaintext[i] {
				correct = false
				break
			}
		}
		if correct {
			successes++
		} else {
			fmt.Printf("  MISMATCH for set %v\n", S)
		}
	}

	fmt.Printf("Combine:  %v total (%v avg per set)\n", combineTotal, combineTotal/time.Duration(len(sets)))
	fmt.Printf("FinDec:   %v total (%v avg per set)\n", findecTotal, findecTotal/time.Duration(len(sets)))
	fmt.Printf("\nResults: %d/%d qualifying sets decrypted correctly\n", successes, len(sets))

	if successes == len(sets) {
		fmt.Println("\nSUCCESS: All qualifying sets produce correct decryption!")
	} else {
		fmt.Printf("\nFAILURE: %d sets failed\n", len(sets)-successes)
	}
}
