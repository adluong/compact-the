package scheme

import (
	"compact-the/noise"
	"compact-the/params"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
)

// PartDec implements Algorithm 3: PartDec(pp, j, s_j, ct) → d_j.
// Computes the partial decryption d_j = c_1 · s_j + e_j^sm.
// The share s_j must be in coefficient form.
// The result d_j is returned in coefficient form.
func PartDec(pp *params.PublicParams, share ring.Poly, ct *rlwe.Ciphertext) (ring.Poly, error) {
	ringQ := pp.BGVParams.RingQ()
	level := ct.Level()
	ringQLvl := ringQ.AtLevel(level)

	// Extract c_1 from ciphertext
	c1 := ct.Value[1]

	// We need both c1 and share in NTT form for ring multiplication.
	c1NTT := ringQ.NewPoly()
	c1NTT.CopyLvl(level, c1)
	if !ct.IsNTT {
		ringQLvl.NTT(c1NTT, c1NTT)
	}

	shareNTT := ringQ.NewPoly()
	shareNTT.CopyLvl(level, share)
	ringQLvl.NTT(shareNTT, shareNTT)

	// Convert to Montgomery form for MulCoeffsMontgomery
	ringQLvl.MForm(c1NTT, c1NTT)

	// Compute signal = c_1 · s_j in NTT domain
	signal := ringQ.NewPoly()
	ringQLvl.MulCoeffsMontgomery(c1NTT, shareNTT, signal)

	// Convert back to coefficient form
	ringQLvl.INTT(signal, signal)

	// Sample smudging noise (uniform in [-B_sm, B_sm]) using CSPRNG
	smudgingNoise := noise.SampleSmudgingNoiseFast(ringQLvl, pp.Bsm)

	// d_j = signal + smudging noise
	dj := ringQ.NewPoly()
	ringQLvl.Add(signal, smudgingNoise, dj)

	return dj, nil
}

// PartDecRingMul computes only the ring multiplication c_1 · s_j (no noise).
// Exported for benchmark decomposition.
func PartDecRingMul(pp *params.PublicParams, share ring.Poly, ct *rlwe.Ciphertext) ring.Poly {
	ringQ := pp.BGVParams.RingQ()
	level := ct.Level()
	ringQLvl := ringQ.AtLevel(level)

	c1 := ct.Value[1]
	c1NTT := ringQ.NewPoly()
	c1NTT.CopyLvl(level, c1)
	if !ct.IsNTT {
		ringQLvl.NTT(c1NTT, c1NTT)
	}

	shareNTT := ringQ.NewPoly()
	shareNTT.CopyLvl(level, share)
	ringQLvl.NTT(shareNTT, shareNTT)
	ringQLvl.MForm(c1NTT, c1NTT)

	signal := ringQ.NewPoly()
	ringQLvl.MulCoeffsMontgomery(c1NTT, shareNTT, signal)
	ringQLvl.INTT(signal, signal)
	return signal
}

// PartDecSmudgingNoise samples only the smudging noise using crypto/rand (conservative).
// Exported for benchmark decomposition.
func PartDecSmudgingNoise(pp *params.PublicParams, ct *rlwe.Ciphertext) ring.Poly {
	ringQ := pp.BGVParams.RingQ()
	ringQLvl := ringQ.AtLevel(ct.Level())
	return noise.SampleSmudgingNoise(ringQLvl, pp.Bsm)
}

// PartDecSmudgingNoiseFast samples only the smudging noise using ChaCha20 CSPRNG.
// Exported for benchmark decomposition.
func PartDecSmudgingNoiseFast(pp *params.PublicParams, ct *rlwe.Ciphertext) ring.Poly {
	ringQ := pp.BGVParams.RingQ()
	ringQLvl := ringQ.AtLevel(ct.Level())
	return noise.SampleSmudgingNoiseFast(ringQLvl, pp.Bsm)
}
