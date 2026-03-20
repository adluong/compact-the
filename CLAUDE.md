# Compact THE — Project Guide

## What This Is

A Go proof-of-concept of a compact (t,N) threshold homomorphic encryption scheme using LSSS-based secret sharing with adjugate reconstruction. The paper lives at `../myscheme-latex.md`.

Built on Lattigo v6.2.0 for BGV/BFV ring operations.

## Architecture

Five core algorithms in `scheme/`:
- `setup.go` — Algorithm 1: constructs share matrix M, validates qualifying sets, computes noise bound
- `share.go` — Algorithm 2: splits secret key into N shares via M
- `partdec.go` — Algorithm 3: partial decryption d_j = c_1·s_j + smudging_noise
- `combine.go` — Algorithm 4: b' = δ·c_0 + Σ λ̂_j·d_j using adjugate cofactors
- `findec.go` — Algorithm 5: absorbs determinant δ via δ⁻¹ mod p, uses BGV Encoder.Decode

Supporting packages:
- `lsss/` — matrix construction (Vandermonde + hardcoded search matrices), qualifying set enumeration, determinant/cofactor computation
- `params/` — Lattigo BGV parameter wrapper, `DefaultRegimeB()` returns standard params
- `noise/` — smudging noise sampling (uniform, not Gaussian), noise bound computation

## Key Conventions

- `bw=1` in Setup means "use the best available matrix" (SearchBW1 → HardcodedBW1 → Vandermonde fallback). The actual B_W is computed from the matrix via `MaxAbsEntry(W)`.
- `bw=0` in Setup means "always use Vandermonde".
- All polynomials are in coefficient form unless explicitly noted (NTT conversions happen inside algorithms).
- Qualifying set indices are 0-based. Parties 0..t-2 are F-parties (free), t-1..N-1 are L-parties (linear).
- The determinant δ is passed from Combine to FinDec as a plain int64. FinDec computes δ⁻¹ mod T.

## Critical Constraints

- **p = 65537** (Fermat prime), NOT 131072. Lattigo requires gcd(p,Q)=1 and p ≡ 1 (mod 2n) for SIMD slot packing.
- **B_W=1 does NOT exist for (t=3, N=10).** Exhaustively proven via backtracking search. B_W=3 matrix is hardcoded.
- **B_W=1 at t=2 is limited to N≤4** (pigeonhole on {±1,0}² rows).
- **Regime B only.** The implementation uses the all-F RLWE noise model (B_sm = B_W × B_ct × 2^κ). Regime A (statistical smudging) would need different, larger parameters.
- **ComputeBsmRegimeB uses exact big.Int arithmetic**: `B_sm = B_W × 2^(bctLog2 + κ)`. For B_W=3, B_sm = 3×2^60 (exact). The `PublicParams.Bsm` field is `*big.Int`.
- **PartDec uses ChaCha20 CSPRNG** (`SampleSmudgingNoiseFast`): ~1.9ms per call. Conservative `crypto/rand` sampler (`SampleSmudgingNoise`) is kept as fallback (~14ms). For B_W ≤ 3, the CSPRNG uses uint64 fast path; for larger B_W, falls back to big.Int with top-byte masking.

## Testing

```bash
go test ./lsss/ ./params/ ./scheme/   # unit + integration (368+ qualifying sets)
go test -v -run TestNoiseMarginAnalysis ./scheme/  # noise margin for all configs
go test -v -run TestParameterSweep ./scheme/  # tightest viable logQ per config
go test -v -run TestMemoryUsage ./scheme/  # byte sizes of crypto objects
go test -bench=. -count=5 -benchtime=3s ./bench/  # benchmarks with stats
```

Use `benchstat` for benchmark analysis:
```bash
go test -bench=. -count=10 -benchtime=5s ./bench/ > bench_raw.txt
benchstat bench_raw.txt
```

## Test Patterns

- `testEndToEnd(t, N, threshold, bw, kappa)` — reusable helper that tests all C(N,t) qualifying sets
- Noise margin analysis iterates all configs and all sets, computes worst-case margin
- Parameter sweep tries 13 LogQ levels × 8 configs to find minimum viable parameters

## File Quick Reference

| Need to... | Look at |
|------------|---------|
| Add a new (t,N) hardcoded matrix | `lsss/matrix.go` → `HardcodedBW1()` |
| Change BGV parameters | `params/params.go` → `DefaultRegimeB()` |
| Understand noise bounds | `noise/bounds.go` → `ComputeBsmRegimeB()`, `VerifyCorrectness()` |
| Add a benchmark | `bench/bench_test.go` — follow existing pattern with `setupBench` + `b.ResetTimer()` |
| Add an end-to-end test | `scheme/scheme_test.go` — call `testEndToEnd()` |
| Run the B_W search | `cmd/bw1search/main.go` — backtracking with row-by-row pruning |

## Paper Relationship

The paper source is `../myscheme-latex.md` (single-file LaTeX). Paper fixes are tracked in `../report-feedback.md` (v1), `../report-feedback-v2.md` (v2), and `../report-feedback-v3.md` (v3). Implementation reports are `implementation-report.md` (v1), `implementation-report-v2.md` (v2), `implementation-report-v3.md` (v3), `implementation-report-v4.md` (v4 — publication-ready benchmarks).

When reporting noise ratios R(t, B_W) for the B_W=3 matrix, always specify "empirical" — the value R=48 uses actual worst Λ_S=16, not the Hadamard bound Λ_S≤54 which gives R≤162.
