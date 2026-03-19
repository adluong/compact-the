# Compact Threshold Homomorphic Encryption (THE)

A Go implementation of the CDM00 compact threshold homomorphic encryption scheme using LSSS-based secret sharing with adjugate-based reconstruction and determinant absorption during BFV/BGV decoding.

Built on [Lattigo v6.2.0](https://github.com/tuneinsight/lattigo) (Tuneinsight).

## Prerequisites

- Go 1.24.0+
- Lattigo v6.2.0 (fetched automatically via `go mod`)

## Quick Start

```bash
# Run the demo (N=5, t=3, B_W=1)
cd compact-the
go run ./cmd/demo

# Run all tests
go test ./...

# Run benchmarks
go test -bench=. -benchtime=5s ./bench/
```

## API Reference

The scheme consists of five core algorithms in the `scheme` package.

### Algorithm 1: Setup

```go
func Setup(bgvParams bgv.Parameters, numParties, threshold, bw, kappa int) (*params.PublicParams, error)
```

Initializes the scheme. Constructs the LSSS share matrix M, validates all C(N,t) qualifying sets, and computes the smudging noise bound B_sm.

- `bgvParams` -- Lattigo BGV parameters (use `params.DefaultRegimeB()` for defaults)
- `numParties` -- Number of parties N
- `threshold` -- Reconstruction threshold t
- `bw` -- Entry bound on the W matrix (1 for optimal, 0 for Vandermonde)
- `kappa` -- Statistical security parameter (typically 40)

### Algorithm 2: Share

```go
func Share(pp *params.PublicParams, sk *rlwe.SecretKey) ([]ring.Poly, error)
```

Splits a secret key into t shares using the LSSS matrix M. Returns one share polynomial per party in coefficient form.

### Algorithm 3: PartDec

```go
func PartDec(pp *params.PublicParams, share ring.Poly, ct *rlwe.Ciphertext) (ring.Poly, error)
```

Computes a partial decryption d_j = c_1 * s_j + e_j^sm where e_j^sm is uniform smudging noise in [-B_sm, B_sm].

### Algorithm 4: Combine

```go
func Combine(pp *params.PublicParams, qualifyingSet []int, ct *rlwe.Ciphertext, partialDecs map[int]ring.Poly) (ring.Poly, int64, error)
```

Combines partial decryptions from a qualifying set S using adjugate-based reconstruction. Returns the combined polynomial b' and the determinant delta (needed by FinDec).

### Algorithm 5: FinDec

```go
func FinDec(pp *params.PublicParams, det int64, bPrime ring.Poly) ([]uint64, error)
```

Final decoding step. Absorbs the determinant delta via modular inverse and uses the BGV encoder to recover plaintext slots.

## Parameter Suites

All configurations use n=8192, logQ~218, T=65537, kappa=40.

| Configuration | t | N | B_W | W Matrix | Qualifying Sets |
|---|---|---|---|---|---|
| Minimal baseline | 2 | 3 | 1 | Hardcoded | 3 |
| Optimal t=3 | 3 | 5 | 1 | Hardcoded | 10 |
| Large N, low t | 2 | 10 | 9 | Vandermonde | 45 |
| Paper reference (S11) | 3 | 10 | 64 | Vandermonde | 120 |
| Higher threshold | 4 | 8 | 125 | Vandermonde | 70 |
| Stress test | 5 | 10 | 1296 | Vandermonde | 252 |

## Benchmark Results

Hardware: Intel Core i7-8700K @ 3.70GHz, Linux (WSL2), Go 1.24.0.

### Per-Algorithm Timing

| Operation | t=2, N=3 (B_W=1) | t=3, N=5 (B_W=1) | t=4, N=8 (Vand) |
|---|---|---|---|
| Share | 2.97 ms | 6.73 ms | 9.46 ms |
| PartDec | 19.7 ms | 19.7 ms | 23.9 ms |
| Combine | 0.82 ms | 0.84 ms | 0.99 ms |
| FinDec | 0.55 ms | 0.55 ms | 0.58 ms |

Note: PartDec is dominated by smudging noise sampling (~17 ms of 19.7 ms); the ring multiply is ~0.5 ms.

### End-to-End Timing

| Configuration | Time | Spec Target |
|---|---|---|
| t=2, N=3, B_W=1 | 74.7 ms | < 100 ms |
| t=3, N=5, B_W=1 | 112.1 ms | -- |
| t=3, N=10, Vandermonde | 248.4 ms | -- |
| t=4, N=8, Vandermonde | 197.3 ms | -- |
| t=5, N=10, Vandermonde | 262.9 ms | -- |

## Noise Margin Analysis

Correctness condition: \|delta\| * B_ct + Lambda_S * B_sm < Delta/2, where Delta/2 = Q/(2T) ~ 2^202.

| Configuration | B_W | log2(B_sm) | Worst \|delta\| | Worst Lambda_S | Noise (bits) | Margin (bits) |
|---|---|---|---|---|---|---|
| t=2, N=3, B_W=1 | 1 | 60 | 1 | 2 | ~61 | **141** |
| t=2, N=10, Vand | 9 | 63 | 2 | 16 | ~68 | **134** |
| t=3, N=5, B_W=1 | 1 | 60 | 1 | 2 | ~61 | **141** |
| t=3, N=10, Vand | 64 | 66 | 20 | 260 | ~75 | **127** |
| t=4, N=8, Vand | 125 | 66 | 228 | 1068 | ~77 | **125** |
| t=5, N=10, Vand | 1296 | 70 | 16254 | 133866 | ~87 | **115** |

Margin drops ~12-15 bits per unit increase in t with Vandermonde matrices. The scheme remains practical through t=5 with Regime B parameters.

## Project Structure

```
compact-the/
  go.mod, go.sum
  implementation-report.md
  params/
    params.go            # PublicParams, DefaultRegimeB(), VerifyParams()
    params_test.go
  lsss/
    matrix.go            # VandermondeW, BuildM, SearchBW1, HardcodedBW1
    adjugate.go          # Determinant, Adjugate, FirstRowCofactors
    adjugate_test.go
    qualify.go           # AllQualifyingSets, ExtractSubmatrix, ValidateAllSets
  scheme/
    setup.go             # Algorithm 1: Setup
    share.go             # Algorithm 2: Share
    partdec.go           # Algorithm 3: PartDec
    combine.go           # Algorithm 4: Combine
    findec.go            # Algorithm 5: FinDec
    helpers.go           # scalarMulSigned, roundDivBigInt, copyPoly
    scheme_test.go       # End-to-end + edge case tests
  noise/
    bounds.go            # ComputeBsmRegimeB, VerifyCorrectness
    smudging.go          # SampleSmudgingNoise (uniform)
  bench/
    bench_test.go        # 16 benchmark configurations
  cmd/demo/
    main.go              # Full pipeline demo
```

## Testing

```bash
# Full test suite (248 qualifying sets verified across 6 configurations)
go test ./...

# Verbose output
go test -v ./scheme/ ./lsss/ ./params/

# Run specific test
go test -v -run TestEndToEnd_T3_N10_Vandermonde ./scheme/

# Benchmarks
go test -bench=. -benchtime=5s ./bench/

# Single benchmark
go test -bench=BenchmarkEndToEnd_T2_N3 -benchtime=5s ./bench/
```

### Test Coverage

| Test | Config | Sets Verified |
|---|---|---|
| TestEndToEnd_T2_N3_BW1 | t=2, N=3, B_W=1 | 3 |
| TestEndToEnd_T3_N5_BW1 | t=3, N=5, B_W=1 | 10 |
| TestEndToEnd_T2_N10_Vandermonde | t=2, N=10, B_W=9 | 45 |
| TestEndToEnd_T3_N10_Vandermonde | t=3, N=10, B_W=64 | 120 |
| TestEndToEnd_T4_N8_Vandermonde | t=4, N=8, B_W=125 | 70 |
| TestMultipleDecryptions | t=2, N=3 (5 trials) | 5 |
| TestInsufficientParties | t=2, N=3 (1 party) | -- |
| TestCorruptedPartialDec | t=2, N=4 (corrupted) | 6 |
| TestNoiseMarginAnalysis | 6 configurations | -- |

## Known Deviations from Spec

1. **Plaintext modulus p=65537** (vs p=131072) -- Lattigo requires gcd(p,Q)=1 and p = 1 mod 2N
2. **No cryptographic erasure in Share** -- Go's GC prevents reliable memory wiping
3. **FinDec uses BGV Encoder.Decode** instead of manual CRT -- required for Lattigo's slot-packed encoding
4. **No hardcoded B_W=1 matrix for (t=3, N=10)** -- search space ~282 billion, falls back to Vandermonde (B_W=64, margin still 127 bits)
5. **Smudging noise is uniform** (not Gaussian) -- matches the spec's [-B_sm, B_sm] requirement

See `implementation-report.md` for full details.

## References

- CDM00 construction: LSSS-based compact threshold FHE with adjugate reconstruction
- [Lattigo v6.2.0](https://github.com/tuneinsight/lattigo) -- Lattice-based cryptographic library
