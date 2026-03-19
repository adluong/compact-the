package scheme

import (
	"fmt"
	"math/big"

	"compact-the/params"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

// FinDec implements Algorithm 5: FinDec(pp, S, b') → m' or ⊥.
//
// In Lattigo's unified BGV, the plaintext is encoded as T⁻¹·NTT_T(perm(m)) mod Q,
// where NTT_T and perm handle slot packing. After threshold combining:
//
//	b' = δ · (c0 + c1·sk) + noise = δ · T⁻¹·encoded(m) + noise
//
// We use the BGV encoder's Decode to properly undo the encoding, then multiply
// by δ⁻¹ mod T to remove the determinant factor.
func FinDec(pp *params.PublicParams, det int64, bPrime ring.Poly) ([]uint64, error) {
	T := pp.BGVParams.PlaintextModulus()

	// Compute δ⁻¹ mod T
	detBig := big.NewInt(det)
	TBig := new(big.Int).SetUint64(T)
	detInv := new(big.Int).ModInverse(detBig, TBig)
	if detInv == nil {
		return nil, fmt.Errorf("δ=%d has no inverse mod T=%d", det, T)
	}
	detInvU64 := detInv.Uint64()

	// Create a fake plaintext from b' for the BGV decoder.
	// The decoder expects the polynomial in the same form as RLWE decrypt output.
	// Lattigo's decryptor outputs in NTT form (when ct.IsNTT=true).
	// Our b' is in coefficient form, so we set IsNTT=false.
	level := bPrime.Level()
	pt := rlwe.NewPlaintext(pp.BGVParams, level)
	pt.Value.CopyLvl(level, bPrime)
	pt.IsNTT = false
	pt.IsBatched = true
	pt.Scale = pp.BGVParams.DefaultScale()

	// Use the BGV encoder to decode (handles slot packing, T*x mod Q mod T)
	encoder := bgv.NewEncoder(pp.BGVParams)
	slots := pp.BGVParams.MaxSlots()
	decoded := make([]uint64, slots)
	if err := encoder.Decode(pt, decoded); err != nil {
		return nil, fmt.Errorf("BGV Decode failed: %w", err)
	}

	// Each decoded[i] = δ·m[i] mod T. Multiply by δ⁻¹ mod T.
	result := make([]uint64, slots)
	for i := 0; i < slots; i++ {
		result[i] = (decoded[i] * detInvU64) % T
	}

	return result, nil
}
