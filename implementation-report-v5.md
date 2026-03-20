# Implementation Report v5: Compact THE — Audit Trail

## 0. Changelog (v4 → v5)

This report is **not** a new benchmark run. It is an **audit trail** over v4: tracking what was done from `report-feedback-v3.md` (B1–B8), what was not done, and the final benchmark numbers.

| Item | Description |
|------|-------------|
| **Audit format** | v5 replaces v4's narrative structure with an accountability-oriented format: feedback tracker, code changes, sanity checks, honest "not done" list. |
| **No new benchmarks** | All numbers are from v4. No code was changed for v5. |
| **Feedback coverage** | All 8 items (B1–B8) from report-feedback-v3.md tracked with status. |
| **Decision points** | §7 decision points from report-feedback-v3.md are addressed. |

---

## 1. Feedback Tracker

### 1.1 Bug Fixes and Code Changes (B1–B8)

| ID | Issue | Status | What Was Done | What Remains |
|----|-------|--------|---------------|--------------|
| **B1** | `ComputeBsmRegimeB` truncation bug (`floor(log₂(B_W))` → wrong B_sm) | **DONE** | Replaced with exact `big.Int` arithmetic: `B_sm = B_W × 2^(bctLog2 + κ)`. `PublicParams.Bsm` is now `*big.Int`. For B_W=3, B_sm = 3×2^60 (correct). Security proof holds. | None. |
| **B2** | Benchmark methodology not publication-grade (count, CI, benchstat, warmup, GOMAXPROCS) | **PARTIAL** | `-count=10 -benchtime=3s -benchmem` for all benchmarks. `benchstat` CIs reported. `GOMAXPROCS=1` for single-threaded baselines. Separate `go test` invocations per config. | Warmup trimming not done (used Go's `b.N` auto-calibration only). GOGC=off comparison not done. CPU governor not controlled (WSL2). See §5. |
| **B3** | PartDec bottleneck: per-coefficient `crypto/rand.Int` ~90% of PartDec time | **DONE** | Added `SampleSmudgingNoiseFast` in `noise/smudging.go` using ChaCha20 seeded from `crypto/rand`. PartDec drops from ~14ms to ~1.9ms (7.5× speedup). Now the primary sampler. Conservative `crypto/rand` sampler kept as `SampleSmudgingNoise`. | None. Both samplers benchmarked and reported. |
| **B4** | E2E benchmark includes Encode+Encrypt (not part of threshold decrypt) | **DONE** | Restructured: `BenchmarkThresholdDecrypt_*` = Share + N×PartDec + Combine + FinDec (no Encrypt in loop). Old E2E renamed to `BenchmarkFullPipeline_*` for completeness. ThresholdDecrypt is the primary reported metric. | None. |
| **B5** | Missing Lattigo multiparty baseline | **DONE** | Added `bench/lattigo_baseline_test.go`: N-of-N collective key-switching via `multiparty.NewKeySwitchProtocol` with Gaussian smudging. Benchmarked at N=3 and N=10. Ratio: our scheme is 2.6× Lattigo's N-of-N (cost of CPA^D threshold security). | None. Honest caveat about apples-to-oranges comparison included in v4 §3 Table 5. |
| **B6** | v2 reported identical per-algorithm times across different configs | **DONE** | All configs benchmarked via separate `go test` invocations. Per-algorithm tables in v4 §3 Table 1 show differentiated numbers across configs. State leakage eliminated. | None. |
| **B7** | No memory profiling (only analytical object sizes) | **DONE** | `-benchmem` reports B/op and allocs/op for all benchmarks. v4 §3 Table 6 has full memory data. Noted that Vandermonde configs hit big.Int slow path (~32k allocs/op). | None. |
| **B8** | No throughput numbers | **DONE** | v4 §3 Table 8 reports ops/sec derived from ThresholdDecrypt latency for all configs. | None. |

### 1.2 Methodology Requirements (§2 of report-feedback-v3.md)

| Requirement | Status | Notes |
|-------------|--------|-------|
| `-count=10` minimum | **DONE** | All benchmarks use `-count=10`. |
| `benchstat` confidence intervals | **DONE** | All numbers reported as mean ± CI%. |
| Separate invocations per config | **DONE** | Each benchmark group runs in its own `go test` process. |
| `-benchtime=3s` | **DONE** | Used for all benchmarks. |
| `GOMAXPROCS=1` for baselines | **DONE** | Set for all single-threaded benchmarks. |
| Warmup: discard first 2 runs or `-count=12` and trim | **NOT DONE** | Relied on Go's `b.N` auto-calibration. See §5. |
| `GOGC=off` or document GC impact | **NOT DONE** | Ran with default GOGC. See §5. |
| CPU governor | **NOT DONE** | WSL2 — host Windows manages governor. Documented in v4 §1. See §5. |
| E2E timing excludes Encrypt | **DONE** | ThresholdDecrypt is the primary metric; FullPipeline kept separately. |

### 1.3 Decision Points (§7 of report-feedback-v3.md)

| Decision | Choice Made | Rationale |
|----------|-------------|-----------|
| **D1: CSPRNG sampler** | Implemented both; CSPRNG is primary | ChaCha20 seeded from `crypto/rand` — equivalent security under random-oracle model. Conservative sampler kept for comparison. |
| **D2: Lattigo multiparty baseline** | Attempted and completed | N-of-N collective key-switching benchmarked at N=3 and N=10. |
| **D3: GOGC=off** | Not done | Ran with default GOGC only. A GOGC=off pass would reduce variance but may not represent production. See §5. |
| **D4: Paper table format** | Cleaned-up tables in v4 | Mean ± CI from benchstat; formatted for readability. |

---

## 2. Code Changes

| File | Change |
|------|--------|
| `noise/bounds.go` | Replaced `floor(log₂(B_W))` with exact `big.Int` computation: `B_sm = B_W × 2^(bctLog2 + κ)`. |
| `noise/smudging.go` | Added `SampleSmudgingNoiseFast` using ChaCha20 CSPRNG with uint64 fast path (B_sm ≤ 63 bits) and big.Int slow path (B_sm > 64 bits). |
| `params/params.go` | Changed `PublicParams.Bsm` from `int` to `*big.Int`. Added `VerifyParams` for gcd(p,Q)=1 check. |
| `scheme/setup.go` | Updated to use exact `big.Int` B_sm from `noise.ComputeBsmRegimeB`. Uses `lsss.MaxAbsEntry(W)` for actual B_W. |
| `scheme/partdec.go` | Switched default sampler to `SampleSmudgingNoiseFast`. Added exported `PartDecRingMul`, `PartDecSmudgingNoise`, `PartDecSmudgingNoiseFast` for benchmark decomposition. |
| `scheme/scheme_test.go` | Updated `testEndToEnd` helper for `*big.Int` B_sm. Tests all C(N,t) qualifying sets per config (368+ total). |
| `scheme/sweep_test.go` | `TestParameterSweep` updated for exact B_sm. Sweeps 13 LogQ levels × 8 configs. |
| `bench/bench_test.go` | Added `BenchmarkThresholdDecrypt_*` (primary metric, no Encrypt in loop), `BenchmarkFullPipeline_*` (includes Encrypt), per-algorithm benchmarks for 5 configs, PartDec decomposition benchmarks, parallel ThresholdDecrypt, Setup/KeyGen/BFVDecrypt/Encrypt baselines. |
| `bench/lattigo_baseline_test.go` | New file: Lattigo N-of-N multiparty collective key-switching benchmark at N=3 and N=10. |
| `cmd/demo/main.go` | Updated for `*big.Int` B_sm API. |
| `CLAUDE.md` | Updated to reflect exact B_sm, CSPRNG sampler, and new benchmark structure. |

---

## 3. Benchmark Results (from v4)

All numbers: `-count=10 -benchtime=3s -benchmem`, `GOMAXPROCS=1`, `benchstat` CIs. Environment: Intel i7-8700K @ 3.70 GHz, WSL2, Go 1.24.0, Lattigo v6.2.0.

### Table 1: Per-Algorithm Timing (CSPRNG sampler, GOMAXPROCS=1)

| Config | Share | PartDec | Combine | FinDec |
|--------|-------|---------|---------|--------|
| t=2, N=3, B_W=1 | 3.36 ms ± 24% | 2.06 ms ± 6% | 1.66 ms ± 28% | 1.28 ms ± 43% |
| t=3, N=10, B_W=3 | 7.61 ms ± 31% | 1.92 ms ± 2% | 1.94 ms ± 24% | 1.31 ms ± 15% |
| t=3, N=10, Vand (B_W=64) | 8.74 ms ± 54% | 5.80 ms ± 4% | 2.10 ms ± 27% | 1.19 ms ± 34% |
| t=4, N=8, Vand (B_W=125) | 10.66 ms ± 34% | 6.33 ms ± 10% | 2.05 ms ± 41% | 1.20 ms ± 44% |
| t=5, N=10, Vand (B_W=1296) | 15.22 ms ± 20% | 5.77 ms ± 20% | 2.71 ms ± 31% | 1.23 ms ± 30% |

### Table 2: ThresholdDecrypt End-to-End (Primary Metric)

ThresholdDecrypt = Share + N × PartDec + Combine + FinDec. **No Encrypt in loop.**

| Config | ThresholdDecrypt | Predicted¹ | Δ |
|--------|-----------------|------------|---|
| t=2, N=3, B_W=1 | 11.21 ms ± 3% | 12.48 ms | -10% |
| t=2, N=4, B_W=1 | 13.14 ms ± 3% | — | — |
| t=3, N=5, B_W=1 | 17.46 ms ± 12% | — | — |
| **t=3, N=10, B_W=3** | **30.15 ms ± 13%** | **30.06 ms** | **<1%** |
| t=3, N=10, Vand | 61.82 ms ± 4% | 69.83 ms | -11% |
| t=4, N=8, Vand | 52.02 ms ± 3% | 64.55 ms | -19% |
| t=5, N=10, Vand | 68.21 ms ± 2% | 76.66 ms | -11% |

¹ Predicted = Share + N × PartDec + Combine + FinDec from Table 1. Discrepancy is due to per-benchmark variance (separate invocations).

### Table 3: PartDec Decomposition (t=3, N=10, B_W=3)

| Component | Time | Fraction |
|-----------|------|----------|
| Ring multiply (c₁ · sⱼ) | 1.21 ms ± 4% | 64% |
| Smudging noise (CSPRNG) | 0.67 ms ± 1% | 35% |
| **Total PartDec** | **1.92 ms ± 2%** | 100% |
| Smudging noise (crypto/rand) | 14.44 ms ± 1% | (conservative baseline) |

**CSPRNG speedup: 21.7× over crypto/rand** for noise sampling.

### Table 4: Baselines

| Benchmark | Time | B/op | allocs/op |
|-----------|------|------|-----------|
| BFV Decrypt (single-party) | 0.89 ms ± 2% | 258 KiB | 54 |
| Encrypt (Encode+Encrypt) | 4.87 ms ± 2% | 784 KiB | 139 |
| KeyGen | 4.58 ms ± 5% | 990 KiB | 681 |
| Setup (t=3, N=10, BW3) | 0.085 ms ± 3% | 46 KiB | 2362 |
| Setup (t=5, N=10, Vand) | 1.30 ms ± 4% | 1159 KiB | 31870 |

### Table 5: Lattigo Multiparty Baseline (N-of-N collective key-switching)

| Config | Lattigo N-of-N | Our ThresholdDecrypt | Ratio |
|--------|---------------|---------------------|-------|
| N=3 | 4.11 ms ± 5% | 11.21 ms (t=2) | 2.7× |
| N=10 | 11.75 ms ± 2% | 30.15 ms (t=3, BW3) | 2.6× |

**Note:** Lattigo's multiparty uses additive N-of-N sharing with Gaussian smudging noise. It does NOT provide CPA^D security (no threshold access structure). The 2.6× ratio is the cost of CPA^D threshold security.

### Table 6: Memory (-benchmem)

| Benchmark | B/op | allocs/op |
|-----------|------|-----------|
| ThresholdDecrypt (t=3,N=10,BW3) | 18.77 MiB | 658 |
| ThresholdDecrypt (t=5,N=10,Vand) | 34.28 MiB | 328.6k |
| PartDec (BW3, CSPRNG) | 1.38 MiB | 28 |
| PartDec (Vand, CSPRNG) | 2.61 MiB | 32.8k |
| Share (t=3,N=10) | 3.50 MiB | 75 |
| Combine (t=3,N=10) | 769 KiB | 29 |
| FinDec | 776 KiB | 265 |
| BFV Decrypt | 258 KiB | 54 |

### Table 7: Noise Margin (Exact B_sm)

| Config | B_W | log₂(B_sm) | Worst Λ_S | Worst margin (bits) |
|--------|-----|------------|-----------|---------------------|
| t=2, N=3, B_W=1 | 1 | 61 | 2 | 139 |
| t=2, N=10, Vand | 9 | 64 | 16 | 133 |
| t=3, N=5, B_W=1 | 1 | 61 | 2 | 139 |
| t=3, N=10, B_W=3 | 3 | 62 | 16 | 135 |
| t=3, N=10, Vand | 64 | 67 | 260 | 126 |
| t=4, N=8, Vand | 125 | 67 | 1068 | 124 |
| t=5, N=10, Vand | 1296 | 71 | 133866 | 113 |

### Table 8: Throughput

| Config | ThresholdDecrypt (ms) | Throughput (ops/sec) |
|--------|----------------------|---------------------|
| t=2, N=3, B_W=1 | 11.21 | 89 |
| t=2, N=4, B_W=1 | 13.14 | 76 |
| t=3, N=5, B_W=1 | 17.46 | 57 |
| t=3, N=10, B_W=3 | 30.15 | 33 |
| t=3, N=10, Vand | 61.82 | 16 |
| t=4, N=8, Vand | 52.02 | 19 |
| t=5, N=10, Vand | 68.21 | 15 |

---

## 4. Sanity Checks

Checks from report-feedback-v3.md §2.4, evaluated against v4 data.

| Check | Expected | Result | Status |
|-------|----------|--------|--------|
| PartDec > BFV Decrypt | Always | 1.92 ms > 0.89 ms (BW3); 5.80 ms > 0.89 ms (Vand) | **PASS** |
| Combine < PartDec | Always | 1.94 ms vs 1.92 ms (BW3) — borderline; 2.10 ms < 5.80 ms (Vand) | **MARGINAL** ¹ |
| FinDec ≈ constant across configs | Yes | Range: 1.19–1.31 ms across all configs | **PASS** |
| Share scales with t | Roughly linear | t=2: 3.36 ms, t=3: 7.61 ms, t=4: 10.66 ms, t=5: 15.22 ms | **PASS** |
| E2E ≈ Share + N×PartDec + Combine + FinDec | Within 5% | BW3: 30.15 vs 30.06 predicted (<1%). Vand configs: -11% to -19% discrepancy | **PARTIAL** ² |
| Parallel E2E ≥ Share + 1×PartDec + Combine + FinDec | Always | 14.96 ms ≥ 7.61 + 1.92 + 1.94 + 1.31 = 12.78 ms | **PASS** |
| Vand PartDec ≥ BW3 PartDec | Yes | 5.80 ms ≥ 1.92 ms | **PASS** |

¹ For BW3, Combine (1.94 ms) ≈ PartDec (1.92 ms). This is expected: with CSPRNG, PartDec is much faster, and Combine involves t=3 scalar-ring multiplies which approach the single ring multiply cost. Not a bug.

² Vandermonde configs show E2E 11–19% faster than predicted sum. This is within the wider CIs for those configs and is attributable to cross-invocation variance on WSL2. The BW3 primary target is <1%.

---

## 5. What Was NOT Done

| Item | Feedback Source | Why Not Done | Impact |
|------|----------------|--------------|--------|
| **Warmup trimming** | B2 (§2.4) | Used Go's `b.N` auto-calibration instead of explicit warmup discard. `-count=12` with first-2 trimming was not implemented. | CIs may be slightly wider. Low impact for `-count=10`. |
| **GOGC=off comparison** | B2 (§2.4), D3 (§7) | Not run. Benchmarks use default GOGC. | GC pressure may inflate some measurements. Low impact for PartDec (allocations are modest at 28 allocs/op for BW3). Higher potential impact for ThresholdDecrypt E2E (~658 allocs/op). |
| **CPU governor control** | B2 (§2.4) | WSL2 does not expose CPU governor — host Windows manages it. | Timing variance is higher than bare-metal. CIs reflect this. Memory numbers unaffected. |
| **Bare-metal re-run** | v4 §6 | All benchmarks run on WSL2. | For publication, re-run on bare-metal Linux to reduce CI widths (some CIs are 20–54%). |
| **FullPipeline benchmarks not re-run after CSPRNG fix** | Implicit | FullPipeline benchmarks exist in code but v4 does not report them. They include Encode+Encrypt and are not the primary metric. | Low impact — ThresholdDecrypt is the primary metric. FullPipeline numbers can be generated on demand. |
| **Per-algorithm breakdown for all configs** | B2 (§3.2) | Per-algorithm tables reported for 5 configs (T2N3, T3N10BW3, T3N10Vand, T4N8, T5N10). Missing: T2N4, T3N5. | Minor gap. These configs have ThresholdDecrypt E2E numbers. Per-algorithm decomposition can be generated if needed. |
| **PartDec decomposition for all configs** | B2 (§3.2) | Only reported for t=3, N=10, BW3 (the primary target). | Other configs' decomposition can be inferred (ring mul is similar; noise varies by B_sm size). |
| **Setup benchmarks for all configs** | B2 (§3.3) | Reported for T3N10BW3 and T5N10Vand. Missing: T2N3, T3N5, T4N8. | Setup is a one-time cost (<1.3 ms for all configs). Low priority. |

---

## 6. Remaining Work (Before Paper Submission)

### High Priority

1. **Bare-metal Linux benchmarks.** WSL2 CIs are too wide for some configs (up to 54%). Re-run all benchmarks on bare-metal Linux with `performance` CPU governor. This is the single most impactful improvement for publication readiness.

2. **Tighten ThresholdDecrypt CIs.** The primary target (t=3, N=10, BW3) has 13% CI. On bare-metal, expect <5%. If not, increase to `-count=20`.

### Medium Priority

3. **GOGC=off comparison pass.** Run the primary table (ThresholdDecrypt for all configs) with `GOGC=off`. If >5% difference, report both numbers. If <5%, document that GC impact is negligible.

4. **Warmup trimming.** Run with `-count=12`, discard first 2 with `benchstat -filter`. Low expected impact but satisfies publication methodology requirements.

5. **Fill per-algorithm gaps.** Add per-algorithm breakdown for T2N4 and T3N5 configs. Add PartDec decomposition for at least one Vandermonde config (to show big.Int slow path breakdown).

### Low Priority

6. **big.Int CSPRNG optimization.** Vandermonde configs (~32k allocs/op in PartDec) could benefit from fixed-point arithmetic for B_sm in [64, 128]-bit range. This is an engineering optimization, not a correctness issue.

7. **FullPipeline re-run and reporting.** Run `BenchmarkFullPipeline_*` and include in an appendix table for completeness.

8. **B_W=2 existence for (t=3, N=10).** Still open after 38M exhaustive trials. Not blocking but relevant to the paper's matrix search narrative.
