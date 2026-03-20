package noise

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"math/big"

	"github.com/tuneinsight/lattigo/v6/ring"
	"golang.org/x/crypto/chacha20"
)

// SampleSmudgingNoise samples a polynomial with each coefficient uniformly
// random in [-B_sm, B_sm] and stores it in RNS representation.
// Uses crypto/rand for conservative security. The noise MUST be uniform, not Gaussian.
func SampleSmudgingNoise(ringQ *ring.Ring, bsm *big.Int) ring.Poly {
	n := ringQ.N()
	level := ringQ.Level()
	result := ringQ.NewPoly()

	// range = 2*B_sm + 1
	rangeSize := new(big.Int).Lsh(bsm, 1)
	rangeSize.Add(rangeSize, big.NewInt(1))

	for l := 0; l < n; l++ {
		// Sample uniform in [0, 2*B_sm]
		val, _ := crypto_rand.Int(crypto_rand.Reader, rangeSize)
		// Subtract B_sm to center: val ∈ [-B_sm, B_sm]
		val.Sub(val, bsm)

		// Reduce mod each RNS prime and store
		for i := 0; i <= level; i++ {
			qi := ringQ.SubRings[i].Modulus
			qiBig := new(big.Int).SetUint64(qi)
			// Compute val mod qi (positive representative)
			coeff := new(big.Int).Mod(val, qiBig)
			result.Coeffs[i][l] = coeff.Uint64()
		}
	}

	return result
}

// SampleSmudgingNoiseFast samples a polynomial with each coefficient uniformly
// random in [-B_sm, B_sm] using a ChaCha20 stream seeded from crypto/rand.
// Equivalent security to SampleSmudgingNoise under the random-oracle model.
// The noise MUST be uniform, not Gaussian.
func SampleSmudgingNoiseFast(ringQ *ring.Ring, bsm *big.Int) ring.Poly {
	n := ringQ.N()
	level := ringQ.Level()
	result := ringQ.NewPoly()

	// Seed ChaCha20 from crypto/rand
	var seed [32]byte
	var nonce [12]byte
	crypto_rand.Read(seed[:])
	crypto_rand.Read(nonce[:])
	stream, _ := chacha20.NewUnauthenticatedCipher(seed[:], nonce[:])

	// range = 2*B_sm + 1
	rangeSize := new(big.Int).Lsh(bsm, 1)
	rangeSize.Add(rangeSize, big.NewInt(1))
	rangeBits := rangeSize.BitLen()
	byteLen := (rangeBits + 7) / 8
	topMask := byte((1 << uint(rangeBits%8)) - 1)
	if topMask == 0 {
		topMask = 0xFF
	}

	// Collect moduli once
	moduli := make([]uint64, level+1)
	for i := 0; i <= level; i++ {
		moduli[i] = ringQ.SubRings[i].Modulus
	}

	// Fast path: if rangeSize fits in uint64
	if rangeBits <= 64 {
		bsmU64 := bsm.Uint64()
		rangeSizeU64 := 2*bsmU64 + 1
		// Bit mask for rejection: mask to rangeBits
		var mask uint64
		if rangeBits == 64 {
			mask = ^uint64(0)
		} else {
			mask = (1 << uint(rangeBits)) - 1
		}

		// Pre-allocate buffer for ChaCha20 output (8 bytes per sample, 2x for rejection)
		buf := make([]byte, 8*n*2)
		stream.XORKeyStream(buf, buf)
		bufPos := 0

		for l := 0; l < n; l++ {
			for {
				if bufPos+8 > len(buf) {
					buf = make([]byte, 8*n)
					stream.XORKeyStream(buf, buf)
					bufPos = 0
				}
				raw := binary.LittleEndian.Uint64(buf[bufPos:]) & mask
				bufPos += 8
				if raw < rangeSizeU64 {
					for i := 0; i <= level; i++ {
						qi := moduli[i]
						if raw >= bsmU64 {
							result.Coeffs[i][l] = (raw - bsmU64) % qi
						} else {
							diff := (bsmU64 - raw) % qi
							if diff == 0 {
								result.Coeffs[i][l] = 0
							} else {
								result.Coeffs[i][l] = qi - diff
							}
						}
					}
					break
				}
			}
		}
		return result
	}

	// Slow path: rangeSize > 64 bits, use big.Int with top-byte masking
	buf := make([]byte, byteLen*n*2)
	stream.XORKeyStream(buf, buf)
	bufPos := 0

	// Pre-allocate big.Int scratch space for mod reduction
	qiBigs := make([]*big.Int, level+1)
	for i := 0; i <= level; i++ {
		qiBigs[i] = new(big.Int).SetUint64(moduli[i])
	}
	val := new(big.Int)
	coeff := new(big.Int)

	for l := 0; l < n; l++ {
		for {
			if bufPos+byteLen > len(buf) {
				buf = make([]byte, byteLen*n)
				stream.XORKeyStream(buf, buf)
				bufPos = 0
			}
			chunk := buf[bufPos : bufPos+byteLen]
			// Mask top byte to reduce rejection rate
			chunk[0] &= topMask
			val.SetBytes(chunk)
			bufPos += byteLen
			if val.Cmp(rangeSize) < 0 {
				val.Sub(val, bsm)
				for i := 0; i <= level; i++ {
					coeff.Mod(val, qiBigs[i])
					result.Coeffs[i][l] = coeff.Uint64()
				}
				break
			}
		}
	}

	return result
}
