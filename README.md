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
go test ./lsss/ ./params/ ./scheme/

# Run benchmarks
go test -bench=. -benchtime=5s ./bench/

# Run the B_W search utility (exhaustive proof B_W=1 impossible for t=3, N=10)
go run ./cmd/bw1search
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
- `bw` -- Entry bound request (1 for smallest available, 0 for Vandermonde)
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

All configurations use n=8192, logQ~219, T=65537, kappa=40.

| Configuration | t | N | B_W | W Matrix | Qualifying Sets |
|---|---|---|---|---|---|
| Minimal baseline | 2 | 3 | 1 | Hardcoded | 3 |
| Optimal t=2 | 2 | 4 | 1 | Hardcoded | 6 |
| Large N, low t | 2 | 10 | 9 | Vandermonde | 45 |
| Optimal t=3 | 3 | 5 | 1 | Hardcoded | 10 |
| Search matrix | 3 | 10 | 3 | Hardcoded (search) | 120 |
| Vandermonde t=3 | 3 | 10 | 64 | Vandermonde | 120 |
| Higher threshold | 4 | 8 | 125 | Vandermonde | 70 |
| Stress test | 5 | 10 | 1296 | Vandermonde | 252 |

## Benchmark Results

Hardware: Intel Core i7-8700K @ 4.40 GHz (overclocked), Linux (WSL2), Go 1.24.0.

### Per-Algorithm Timing

Mean over 5 runs (`-count=5 -benchtime=3s`).

| Operation | t=2, N=3 (B_W=1) | t=3, N=10 (B_W=3) | t=3, N=10 (Vand) | t=4, N=8 (Vand) | t=5, N=10 (Vand) |
|---|---|---|---|---|---|
| Share | 3.09 ms | 7.18 ms | 7.48 ms | 9.15 ms | 12.59 ms |
| PartDec | 21.04 ms | 21.24 ms | 23.72 ms | 23.52 ms | 24.68 ms |
| Combine | 0.91 ms | 0.99 ms | 0.93 ms | 1.15 ms | 1.15 ms |
| FinDec | 0.62 ms | 0.61 ms | 0.61 ms | 0.60 ms | 0.60 ms |

PartDec decomposition (t=3, N=10, B_W=3): ring multiply **1.10 ms** (5%) + smudging noise **18.59 ms** (88%) + overhead ~1.55 ms (7%).

Single-party BFV Decrypt baseline: **1.12 ms**. Algebraic threshold overhead (ring mul + Combine + FinDec) is **2.4x** the baseline.

### End-to-End Timing

| Configuration | Sequential | Parallel (goroutines) |
|---|---|---|
| t=2, N=3, B_W=1 | 76.7 ms | -- |
| t=3, N=5, B_W=1 | 121.5 ms | -- |
| t=3, N=10, B_W=3 | 228.2 ms | **67.2 ms** |
| t=3, N=10, Vandermonde | 252.3 ms | -- |
| t=4, N=8, Vandermonde | 205.9 ms | -- |
| t=5, N=10, Vandermonde | 259.0 ms | -- |

E2E includes Encode+Encrypt (~5.3 ms). Parallel PartDec across N goroutines gives 3.4x speedup.

### Memory (t=3, N=10)

| Object | Size |
|---|---|
| Share s_j | 256 KB |
| Partial decryption d_j | 256 KB |
| Ciphertext ct | 512 KB |
| This work per-party storage | **256 KB** |
| {0,1}-LSSS per-party storage | **~4.3 GB** |
| **Storage reduction** | **~17,500x** |

## Noise Margin Analysis

Correctness condition: |delta| * B_ct + Lambda_S * B_sm < Delta/2, where Delta/2 = Q/(2T) ~ 2^202.

| Configuration | B_W | log2(B_sm) | Worst Lambda_S | Margin (bits) |
|---|---|---|---|---|
| t=2, N=3, B_W=1 | 1 | 60 | 2 | **140** |
| t=2, N=10, Vand | 9 | 63 | 16 | **134** |
| t=3, N=5, B_W=1 | 1 | 60 | 2 | **140** |
| t=3, N=10, B_W=3 | 3 | 61 | 16 | **136** |
| t=3, N=10, Vand | 64 | 66 | 260 | **127** |
| t=4, N=8, Vand | 125 | 66 | 1068 | **125** |
| t=5, N=10, Vand | 1296 | 70 | 133866 | **114** |

The B_W=3 search matrix gains 9 bits of margin over Vandermonde at (t=3, N=10). Margin drops ~12-15 bits per unit increase in t with Vandermonde. The scheme remains practical through t=5 with Regime B parameters.

## Project Structure

```
compact-the/
  go.mod, go.sum
  CLAUDE.md                      # Project guide for Claude Code
  implementation-report.md       # v1 report
  implementation-report-v2.md    # v2 report (feedback response, B_W search)
  implementation-report-v3.md    # v3 report (benchmarks with statistics, decomposition, baselines)
  params/
    params.go            # PublicParams, DefaultRegimeB(), VerifyParams()
    params_test.go
  lsss/
    matrix.go            # VandermondeW, BuildM, SearchBW1, HardcodedBW1 (incl. B_W=3)
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
    scheme_test.go       # End-to-end + edge case tests (10 tests)
    sweep_test.go        # Parameter sweep: tightest viable LogQ
  noise/
    bounds.go            # ComputeBsmRegimeB, VerifyCorrectness, LambdaS
    smudging.go          # SampleSmudgingNoise (uniform)
  bench/
    bench_test.go        # 29 benchmark configurations (per-algo, E2E, decomposition, baselines)
  cmd/
    demo/main.go         # Full pipeline demo (N=5, t=3, B_W=1)
    bw1search/main.go    # B_W backtracking search utility
```

## Testing

```bash
# Full test suite (368+ qualifying sets verified across 7 configurations)
go test ./lsss/ ./params/ ./scheme/

# Verbose output
go test -v ./scheme/ ./lsss/ ./params/

# Run specific test
go test -v -run TestEndToEnd_T3_N10_BW3 ./scheme/

# Parameter sweep (finds tightest LogQ per config)
go test -v -run TestParameterSweep ./scheme/

# Benchmarks
go test -bench=. -benchtime=5s ./bench/

# Single benchmark
go test -bench=BenchmarkEndToEnd_T3_N10_BW3 -benchtime=5s ./bench/
```

### Test Coverage

| Test | Config | Sets Verified |
|---|---|---|
| TestEndToEnd_T2_N3_BW1 | t=2, N=3, B_W=1 | 3 |
| TestEndToEnd_T3_N5_BW1 | t=3, N=5, B_W=1 | 10 |
| TestEndToEnd_T2_N10_Vandermonde | t=2, N=10, B_W=9 | 45 |
| TestEndToEnd_T3_N10_BW3 | t=3, N=10, B_W=3 | 120 |
| TestEndToEnd_T3_N10_Vandermonde | t=3, N=10, B_W=64 | 120 |
| TestEndToEnd_T4_N8_Vandermonde | t=4, N=8, B_W=125 | 70 |
| TestMultipleDecryptions | t=2, N=3 (5 trials) | 5 |
| TestInsufficientParties | t=2, N=3 (1 party) | -- |
| TestCorruptedPartialDec | t=2, N=4 (corrupted) | 6 |
| TestNoiseMarginAnalysis | 7 configurations | -- |
| TestParameterSweep | 8 configs x 13 LogQ levels | -- |
| TestMemoryUsage | t=3, N=10, B_W=3 | -- |

## Known Deviations from Spec

1. **Plaintext modulus p=65537** (vs p=131072) -- Lattigo requires gcd(p,Q)=1 and p = 1 mod 2N. Paper corrected.
2. **No cryptographic erasure in Share** -- Go's GC prevents reliable memory wiping.
3. **FinDec uses BGV Encoder.Decode** instead of manual CRT -- required for Lattigo's slot-packed encoding.
4. **B_W=3 for (t=3, N=10)** instead of claimed B_W=1 -- exhaustive search proves B_W=1 impossible; B_W=3 found via randomised search. Noise margin improves from 127 bits (Vandermonde B_W=64) to 136 bits.
5. **Smudging noise is uniform** (not Gaussian) -- matches the spec's [-B_sm, B_sm] requirement.

See `implementation-report-v3.md` for full details including benchmarks with statistics, PartDec decomposition, memory analysis, and baselines.

## References

- CDM00 construction: LSSS-based compact threshold FHE with adjugate reconstruction
- [Lattigo v6.2.0](https://github.com/tuneinsight/lattigo) -- Lattice-based cryptographic library
