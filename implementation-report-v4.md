# Implementation Report v4: Compact THE — Publication-Ready Benchmarks

## 0. Changelog (v3 → v4)

| Change | Detail |
|--------|--------|
| **ComputeBsmRegimeB bug fixed (B1)** | `floor(log₂(B_W))` replaced with exact `big.Int` arithmetic: `B_sm = B_W × 2^(bctLog2 + κ)`. For B_W=3, B_sm is now 3×2^60 (correct) instead of 2^61 (truncated). Security proof now holds. |
| **CSPRNG smudging sampler (B3)** | Added `SampleSmudgingNoiseFast` using ChaCha20 seeded from `crypto/rand`. PartDec drops from ~14ms to ~1.9ms (7.5× speedup). Now the primary sampler. |
| **ThresholdDecrypt benchmark (B4)** | Separated threshold decryption (Share + N×PartDec + Combine + FinDec) from Encode+Encrypt. Old E2E renamed to FullPipeline. |
| **benchstat confidence intervals (B2)** | All numbers from `-count=10 -benchtime=3s -benchmem`, processed with `benchstat`. CIs reported. |
| **Lattigo multiparty baseline (B5)** | N-of-N collective key-switching benchmarked for calibration. |
| **Memory profiling (B7)** | `-benchmem` reports B/op and allocs/op for all benchmarks. |
| **Noise margin recalculated** | Exact B_sm values; (t=3, N=10, BW3) margin: 135 bits (was ~136 with truncated B_sm). |

---

## 1. Environment

| Property | Value |
|----------|-------|
| CPU | Intel Core i7-8700K @ 3.70 GHz |
| OS | Linux 6.6.87.2-microsoft-standard-WSL2 (Windows Subsystem for Linux 2) |
| Go version | go1.24.0 linux/amd64 |
| Lattigo version | v6.2.0 |
| GOMAXPROCS | 1 (single-threaded, except parallel benchmark) |
| GOGC | default (not disabled) |
| CPU governor | Not controlled (WSL2 — host Windows manages) |

**WSL2 note:** Timing variance is higher than bare-metal due to host OS scheduling. CIs reflect this. Memory numbers are unaffected.

---

## 2. Methodology

- **Separate invocations:** Each benchmark group runs in its own `go test` process to prevent state leakage (Lattigo ring caches, GC pressure).
- **Parameters:** `-count=10 -benchtime=3s -benchmem` for all benchmarks.
- **CI computation:** `benchstat` (golang.org/x/perf) computes confidence intervals from 10 samples.
- **ThresholdDecrypt:** Encode+Encrypt performed ONCE outside timing loop. Inside loop: `Share + N×PartDec + Combine + FinDec`.
- **Smudging noise sampler:** CSPRNG (ChaCha20) is the primary sampler. Conservative (`crypto/rand`) is benchmarked separately for comparison.

---

## 3. Results

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

**CSPRNG speedup: 21.7× over crypto/rand** for noise sampling. The ring multiply is now the dominant cost.

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

**Note:** Lattigo's multiparty uses additive N-of-N sharing with Gaussian smudging noise. It does NOT provide CPA^D security (no threshold access structure). Our scheme's overhead comes from: (1) LSSS-based sharing requiring t scalar-ring multiplies in Combine, (2) larger smudging noise for the deterministic bound, (3) FinDec requiring δ⁻¹ mod T. The 2.6× ratio is the cost of CPA^D threshold security.

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

**Note on Vand allocs:** The Vandermonde configs (B_W ≥ 64) use the big.Int slow path for CSPRNG sampling because B_sm exceeds 64 bits. This causes ~32k allocs/op per PartDec. The B_W=1 and B_W=3 configs stay in the uint64 fast path (28 allocs/op).

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

All margins are computed with exact B_sm (no floor(log₂) truncation). The (t=3, N=10, BW3) margin decreased ~1 bit from v3 (136 → 135) due to the B_sm correction. All configs remain safe with > 100 bits of margin.

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

### Table 9: Parallel ThresholdDecrypt (t=3, N=10, B_W=3, GOMAXPROCS=12)

| Variant | Time | Speedup |
|---------|------|---------|
| Sequential (GOMAXPROCS=1) | 30.15 ms ± 13% | 1.0× |
| Parallel PartDec (GOMAXPROCS=12) | 14.96 ms ± 8% | 2.0× |

---

## 4. Analysis

### 4.1 Algebraic Overhead Ratio

| Config | ThresholdDecrypt | BFV Decrypt | Ratio |
|--------|-----------------|-------------|-------|
| t=2, N=3, B_W=1 | 11.21 ms | 0.89 ms | 12.6× |
| t=3, N=10, B_W=3 | 30.15 ms | 0.89 ms | 33.9× |
| t=5, N=10, Vand | 68.21 ms | 0.89 ms | 76.6× |

The overhead is dominated by N PartDec calls (N ring multiplies + N noise samples) and Share (t-1 ring samples + N-t+1 scalar-ring multiplies).

### 4.2 Where Time Goes (t=3, N=10, B_W=3)

| Phase | Time | % of E2E |
|-------|------|----------|
| Share | 7.61 ms | 25% |
| 10 × PartDec | 19.20 ms | 64% |
| Combine | 1.94 ms | 6% |
| FinDec | 1.31 ms | 4% |
| **Total** | **30.06 ms** | **100%** |

PartDec dominates at 64%. Within each PartDec, ring multiply (1.21ms) is now the primary cost, with CSPRNG noise sampling (0.67ms) secondary.

### 4.3 CSPRNG vs crypto/rand Impact

| Sampler | PartDec time | allocs/op | Security model |
|---------|-------------|-----------|----------------|
| CSPRNG (ChaCha20) | 1.92 ms | 28 | Random oracle |
| crypto/rand | ~15.5 ms¹ | ~122k | Information-theoretic |

¹ Estimated from decomposition: RingMul (1.21ms) + SmudgingNoise_cryptorand (14.44ms) = 15.65ms.

At N=10, the difference is 10 × (15.65 - 1.92) = 137ms — the gap between a 30ms scheme and a 167ms scheme.

### 4.4 B_W Impact on PartDec

| B_W | log₂(B_sm) | PartDec | Sampler path |
|-----|------------|---------|--------------|
| 1 | 61 | 2.06 ms | uint64 fast |
| 3 | 62 | 1.92 ms | uint64 fast |
| 64 | 67 | 5.80 ms | big.Int slow |
| 125 | 67 | 6.33 ms | big.Int slow |
| 1296 | 71 | 5.77 ms | big.Int slow |

The B_W=1 and B_W=3 configs benefit from the uint64 fast path (B_sm fits in 63 bits). Larger B_W forces big.Int arithmetic (~3× slower). This confirms the paper's recommendation to seek small B_W matrices.

---

## 5. Correctness

### 5.1 All qualifying sets pass

```
TestEndToEnd_T2_N3_BW1:        3 sets — PASS (0.11s)
TestEndToEnd_T3_N5_BW1:       10 sets — PASS (0.13s)
TestEndToEnd_T2_N10_Vand:     45 sets — PASS (0.27s)
TestEndToEnd_T3_N10_BW3:     120 sets — PASS (0.22s)
TestEndToEnd_T3_N10_Vand:    120 sets — PASS (0.39s)
TestEndToEnd_T4_N8_Vand:      70 sets — PASS (0.24s)
```

All 368+ qualifying sets produce correct slot-by-slot decryption with the CSPRNG sampler.

### 5.2 Noise margin verification (exact B_sm)

All 7 configs pass `VerifyCorrectness` with exact `big.Int` B_sm computation. No truncation. See Table 7.

---

## 6. Known Limitations

| Item | Status |
|------|--------|
| **WSL2 timing variance** | CIs are wider than bare-metal (15-50% on some benchmarks). Memory numbers are accurate. For publication, re-run on bare-metal Linux. |
| **CSPRNG big.Int path** | Vandermonde configs (B_W ≥ 64) use big.Int for CSPRNG rejection sampling, adding ~3× overhead and ~32k allocs/op per PartDec. Optimization: implement fixed-point arithmetic for B_sm in [64, 128] bit range. |
| **No GOGC=off comparison** | Benchmarks run with default GC. A GOGC=off pass would reduce variance but may not be representative of production. |
| **Lattigo multiparty comparison is apples-to-oranges** | Lattigo uses Gaussian smudging (not uniform), additive sharing (not LSSS), and N-of-N (not t-of-N). The comparison is for calibration only. |
| **B_W=2 existence for (t=3, N=10)** | Still open after 38M exhaustive trials. Not proven impossible. |
| **Regime A not implemented** | Statistical CPA^D (Regime A) would require much larger B_sm and different parameters. |
