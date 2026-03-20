package bench

import (
	"math/rand"
	"testing"

	"compact-the/params"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/multiparty"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

// BenchmarkLattigoMultipartyDecrypt_N10 benchmarks Lattigo's native N-of-N
// collective key-switching (threshold decryption) at the same ring parameters
// as the compact THE scheme. This provides a calibration baseline.
//
// Lattigo's multiparty uses additive N-of-N secret sharing + KeySwitchProtocol.
// It does NOT support t-of-N directly (Thresholdizer/Combiner add overhead).
// The noise model uses Gaussian smudging, not uniform.
func BenchmarkLattigoMultipartyDecrypt_N10(b *testing.B) {
	lattigoMultipartyDecrypt(b, 10)
}

func BenchmarkLattigoMultipartyDecrypt_N3(b *testing.B) {
	lattigoMultipartyDecrypt(b, 3)
}

func lattigoMultipartyDecrypt(b *testing.B, numParties int) {
	b.Helper()

	bgvParams, err := params.DefaultRegimeB()
	if err != nil {
		b.Fatal(err)
	}

	// Generate N secret key shares; sum = ideal secret key
	kgen := rlwe.NewKeyGenerator(bgvParams)
	skShares := make([]*rlwe.SecretKey, numParties)
	skIdeal := rlwe.NewSecretKey(bgvParams)
	ringQP := bgvParams.RingQP()
	for i := 0; i < numParties; i++ {
		skShares[i] = kgen.GenSecretKeyNew()
		ringQP.Add(skIdeal.Value, skShares[i].Value, skIdeal.Value)
	}

	// Generate public key and encrypt
	pk := kgen.GenPublicKeyNew(skIdeal)
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
	ct, err := encryptor.EncryptNew(pt)
	if err != nil {
		b.Fatal(err)
	}

	// Create KeySwitch protocol (collective key-switching to decrypt)
	// Use Gaussian noise flooding similar to Lattigo's default
	sigmaSmudging := 8 * rlwe.DefaultNoise
	cks, err := multiparty.NewKeySwitchProtocol(bgvParams, ring.DiscreteGaussian{
		Sigma: sigmaSmudging,
		Bound: 6 * sigmaSmudging,
	})
	if err != nil {
		b.Fatal(err)
	}

	// Target key: zero (decrypt to plaintext)
	skZero := rlwe.NewSecretKey(bgvParams)

	// Pre-allocate shares
	shares := make([]multiparty.KeySwitchShare, numParties)
	for j := 0; j < numParties; j++ {
		shares[j] = cks.AllocateShare(ct.Level())
	}

	ctOut := rlwe.NewCiphertext(bgvParams, 1, ct.Level())
	decryptor := rlwe.NewDecryptor(bgvParams, skZero)
	result := make([]uint64, slots)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Each party generates its key-switch share
		for j := 0; j < numParties; j++ {
			cks.GenShare(skShares[j], skZero, ct, &shares[j])
		}

		// Aggregate all shares
		for j := 1; j < numParties; j++ {
			cks.AggregateShares(shares[0], shares[j], &shares[0])
		}

		// Finalize: key-switch ct to skZero
		cks.KeySwitch(ct, shares[0], ctOut)

		// Decrypt with zero key (just decoding)
		ptDec := decryptor.DecryptNew(ctOut)
		encoder.Decode(ptDec, result)
	}
}
