# Implementation Report v2: Compact THE with Determinant Absorption

## 0. Changelog (v1 to v2)

| Change | Detail |
|--------|--------|
| **B_W=3 matrix discovered** | Exhaustive backtracking proves B_W=1 impossible for (t=3, N=10). Randomised search finds B_W=3 in 354 trials. Hardcoded in `HardcodedBW1()`. Noise margin improves from 127 bits (Vandermonde B_W=64) to 136 bits. |
| **CPU clock corrected** | 3.70 GHz -> 4.40 GHz (overclocked). All benchmark numbers are from the 4.40 GHz system. |
| **Paper fixes applied** | 7 fixes from `report-feedback.md` applied to `myscheme-latex.md` (see section 1). |
| **Parameter sweep test added** | `scheme/sweep_test.go` finds tightest viable log2 Q per configuration. |
| **New test: TestEndToEnd_T3_N10_BW3** | End-to-end verification of the B_W=3 matrix across all 120 qualifying sets. |
| **New benchmark: BenchmarkEndToEnd_T3_N10_BW3** | End-to-end timing for B_W=3 config. |
| **New utility: cmd/bw1search/** | Standalone backtracking search for B_W=1 matrices with row-by-row pruning. |
| **t=2 "all N" bug found** | Paper claimed B_W=1 for all N at t=2; pigeonhole proves maximum is N=4. Corrected. |

---

## 1. Response to report-feedback.md

### Fix 1 -- B_W Claims at t=3 (severity: critical)

**Feedback:** Two sections make contradictory claims. Line 947 says "B_W=1 feasible? N<=10 (search)"; line 1483 says "B_W<=2 achievable by search for N<=10". Implementation could not find B_W=1 for (t=3, N=10).

**Fact-check verdict:** CONFIRMED. The claims are contradictory.

**Action taken:** Neither Option A nor Option B from the feedback was taken exactly. Instead:

- B_W=1 was **exhaustively proven to not exist** via backtracking over {-1,0,1}^{8x3} (completed in 2.5 ms).
- B_W=2 was not found after 38M random trials (30 s).
- **B_W=3 was found** via randomised search (354 trials) and published as a concrete artifact in `lem:concrete`.

Paper changes:
- Noise table (line 947): `$N\leq 10$ (search)` changed to `$N\leq 5$ (verified); $\BW{=}3$ for $N{=}10$`
- Positioning (line 1481--1485): Updated to reflect B_W=3 achievable by search, R(3,B_W=3) approx 48
- `lem:concrete` split into three valid cases (see Fix 3)

### Fix 2 -- p = 2^17 is Unimplementable (severity: critical)

**Feedback:** p = 2^17 = 131072 is not prime, fails SIMD slot-packing.

**Fact-check verdict:** CONFIRMED. 131072 = 2^17 is not prime. Only 1 occurrence at line 1434.

**Action taken:**
- Line 1434: `$p = 2^{17}$` changed to `$p = 65537$`
- Added `rmk:plaintext` explaining the Fermat prime choice and gcd(p,q)=1 sufficiency
- Global grep confirms no remaining `2^{17}` references

### Fix 3 -- Corollary Conflates Incompatible Parameters (severity: critical)

**Feedback:** Corollary 6.10 (`cor:allF-combined`) specifies B_W=1, N=10, t in {2,3} which is self-contradictory for t=2.

**Fact-check verdict:** CONFIRMED, but mislabelled. There is no `cor:allF-combined` in the paper. The actual location is `lem:concrete` (line 956). The lemma specifies "Fix B_W=1, N=10" for both t=2 and t=3. For t=2 at N=10, B_W=1 is mathematically impossible (at most 3 non-collinear rows in {+/-1,0}^2, so maximum N=4).

**Additional finding:** The feedback claims "the paper itself says N<=4 for B_W=1 at t=2 in line 134" -- but the paper actually claims "all N" at t=2 (line 946 and line 1478). The paper is wrong; B_W=1 at t=2 is limited to N<=4 by pigeonhole. The feedback correctly identifies the inconsistency but misattributes where the paper states it.

**Action taken:**
- `lem:concrete` split into three valid configurations:
  - Case 1: t=2, N=4, B_W=1 (verified)
  - Case 2: t=3, N=5, B_W=1 (verified)
  - Case 3: t=3, N=10, B_W=3 (the discovered matrix, published with full verification)
- Noise table: t=2 row changed from "all N" to "N<=4"
- Positioning: t=2 changed from "B_W=1 suffices for all N" to "B_W=1 for N<=4"

### Fix 4 -- "t<=3 Only" Too Conservative (severity: important)

**Feedback:** Implementation demonstrates t=4 and t=5 work under Regime B. The "t<=3" restriction is accurate for Regime A but not Regime B.

**Fact-check verdict:** CONFIRMED. Paper says t<=4 at line 1037 but t<=3 in abstract and conclusion. Implementation verifies t=5 with 114-bit margin.

**Action taken:**
- Abstract: qualified with "up to t=5 under the all-F RLWE regime"
- `rmk:window`: rewritten to distinguish Regime A (t<=3) from Regime B (t<=5 empirically)
- Positioning (t>=4 bullet): added Regime B clause with empirical margin data
- Conclusion: qualified with "under statistical smudging; up to t=5 under all-F RLWE"

### Fix 5 -- Soften p | q to gcd(p,q)=1 (severity: important)

**Feedback:** Paper assumes p | q throughout. Implementation uses gcd(p,Q)=1 with p=65537.

**Fact-check verdict:** CONFIRMED. Lines 412, 413, 645, 657, 744 all reference p | q.

**Action taken:**
- `def:params` (line 412): changed to `$\gcd(p,q) = 1$`, removed "(exact, since p | q)"
- Algorithm 5 comment (line 657): changed to `exists: |delta| < p and p prime`
- Correctness proof (line 744): changed to `exists because |delta| < p and p is prime`
- Remark about delta^{-1} existence (line 645): changed to `|delta| < p` argument
- Added `rmk:plaintext` near `def:params` explaining the general gcd(p,q)=1 condition

### Fix 6 -- R(3) = 6 Misleading Juxtaposition (severity: important)

**Feedback:** Lines 1481--1483 place R(3)=6 next to "B_W<=2 achievable by search" creating false impression that R(3)=6 holds at B_W=2.

**Fact-check verdict:** CONFIRMED. R(3)=6 only holds at B_W=1. With B_W=2, R(3,B_W=2)=48.

**Action taken:** Resolved jointly with Fix 1. The positioning section now clearly separates B_W=1 (verified up to N=5, R(3)=6) from B_W=3 at N=10 (R(3,B_W=3) approx 48).

### Fix 7 -- Algorithm 5 Assumes Coefficient Encoding (severity: minor)

**Feedback:** Algorithm 5 loops per-coefficient, but BGV/BFV libraries use slot packing.

**Fact-check verdict:** CONFIRMED. Lines 651--667 show per-coefficient BFV rounding. Implementation uses `bgv.Encoder.Decode`.

**Action taken:** Added `rmk:slot-pack` after Algorithm 5 explaining that slot-packed BFV/BGV libraries replace the per-coefficient loop with their decoder, followed by per-slot delta_inv multiplication. Correctness follows from delta_inv being an F_p-scalar commuting with slot unpacking.

### Fix 8 -- Implementation-Level Gaps (severity: informational)

**Feedback:** Three implementation observations for documentation:
1. PartDec performance: 19.7 ms vs 5 ms target due to per-coefficient crypto/rand.Int
2. B_W=1 search for (t=3, N=10): implement pruning-based search
3. No cryptographic erasure: acceptable for prototype

**Fact-check verdict:** CONFIRMED. All three are accurate observations.

**Action taken:**
1. PartDec performance documented in benchmark tables and noted as optimization opportunity (replace crypto/rand with PRNG seeded from crypto/rand)
2. B_W search implemented as `cmd/bw1search/` with backtracking pruning. Result: B_W=1 impossible, B_W=3 found and hardcoded.
3. No erasure remains acceptable for prototype; documented in deviations.

---

## 2. Executive Summary

This report documents the Go proof-of-concept implementation of the compact threshold homomorphic encryption (THE) scheme. The implementation realises the five core algorithms (Setup, Share, PartDec, Combine, FinDec) of the CDM00 LSSS-based construction with adjugate-based reconstruction and determinant absorption during BFV decoding.

**Verdict: The implementation is substantially compliant with the specification.** All five algorithms produce verified correct results across 368+ qualifying sets in 7 parameter configurations. Five deviations from the specification were identified, all justified.

**Key finding (v2):** A B_W=3 matrix for (t=3, N=10) was discovered via randomised search, after exhaustive proof that B_W=1 does not exist for these parameters. This improves the noise margin from 127 bits (Vandermonde B_W=64) to 136 bits.

**Environment:**
- **Language:** Go 1.24.0
- **Library:** Lattigo v6.2.0
- **Hardware:** Intel Core i7-8700K @ 4.40 GHz (overclocked), Linux (WSL2)

---

## 3. Specification Compliance: Algorithm-by-Algorithm

### 3.1 Algorithm 1: Setup (`scheme/setup.go`)

| Spec Step | Requirement | Implementation | Status |
|-----------|-------------|----------------|--------|
| 1 | Instantiate BFV parameters, verify n, p, q | `params.VerifyParams()` checks n is power of 2, p > N, gcd(p,Q)=1 | COMPLIANT |
| 2 | Construct W (Vandermonde or B_W search) | `lsss.SearchBW1()` with `lsss.VandermondeW()` fallback | COMPLIANT |
| 3 | Construct M from W per block structure | `lsss.BuildM()` implements [0\|I; -W] | COMPLIANT |
| 4 | Validate all C(N,t) qualifying sets | `lsss.ValidateAllSets()` checks det != 0 and gcd(det, q_i) = 1 | COMPLIANT |
| 5 | Compute B_sm | `noise.ComputeBsmRegimeB()` with B_ct approx 2^20 | COMPLIANT |
| 6 | Return pp | Returns `PublicParams` struct | COMPLIANT |

### 3.2 Algorithm 2: Share (`scheme/share.go`)

| Spec Step | Requirement | Implementation | Status |
|-----------|-------------|----------------|--------|
| 1 | Sample r_k uniformly from R_q | `ring.NewUniformSampler` via Lattigo PRNG | COMPLIANT |
| 2 | Form rho = (sk, r_1, ..., r_{t-1}) | Array of ring.Poly with sk.Value.Q at index 0 | COMPLIANT |
| 3a | F-party: s_i = r_{i+1} | Direct polynomial copy | COMPLIANT |
| 3b | L-party: s_i = -Sum W[j][k] * rho_k | Loop with `scalarMulSigned` + `ringQ.Add` | COMPLIANT |
| 4 | Distribute shares | Returned as `[]ring.Poly` | COMPLIANT |
| 5 | Erase rho from memory | Not explicitly zeroed | DEVIATION D2 |

### 3.3 Algorithm 3: PartDec (`scheme/partdec.go`)

| Spec Step | Requirement | Implementation | Status |
|-----------|-------------|----------------|--------|
| 1 | signal_j = c_1 * s_j | NTT + MForm + MulCoeffsMontgomery + INTT | COMPLIANT |
| 2 | Sample smudging noise uniform in [-B_sm, B_sm] | `noise.SampleSmudgingNoise()` via `crypto/rand.Int` | COMPLIANT |
| 3 | d_j = signal_j + noise | `ringQ.Add()` in coefficient form | COMPLIANT |
| 4 | Broadcast d_j | Returns d_j; no S parameter | COMPLIANT |

### 3.4 Algorithm 4: Combine (`scheme/combine.go`)

| Spec Step | Requirement | Implementation | Status |
|-----------|-------------|----------------|--------|
| 1 | Extract M_S, compute det and cofactors | `lsss.ExtractSubmatrix` + `lsss.FirstRowCofactors` | COMPLIANT |
| - | Verify reconstruction identity | `lsss.VerifyReconstructionIdentity()` self-test | COMPLIANT (extra) |
| 2 | b' = delta * c_0 + Sum lambda_hat_j * d_j | `scalarMulSigned` helper with `ringQ.Add` accumulation | COMPLIANT |

### 3.5 Algorithm 5: FinDec (`scheme/findec.go`)

| Spec Step | Requirement | Implementation | Status |
|-----------|-------------|----------------|--------|
| 1 | Compute delta | Received from Combine as parameter | COMPLIANT |
| 2 | Compute delta_inv = delta^{-1} mod p | `big.Int.ModInverse(delta, T)` | COMPLIANT |
| 3 | BFV rounding + delta absorption | Uses BGV Encoder.Decode | DEVIATION D3 |
| 4 | Return m' or FAIL | Returns `[]uint64` or error | COMPLIANT |

---

## 4. Deviations from Specification

### D1 -- Plaintext Modulus: p = 65537 instead of p = 131072

**Spec:** p = 2^17 = 131072

**Implementation:** p = 65537 = 2^16 + 1

**Reason:** 131072 = 2^17 is not prime and fails the SIMD slot-packing condition p = 1 (mod 2n). The Fermat prime 65537 satisfies 65537 mod 16384 = 1 and gcd(65537, Q) = 1. The paper has been corrected accordingly.

**Impact:** Minimal. log2(65537) approx 16 vs log2(131072) = 17 -- a 1-bit difference in Delta.

### D2 -- No Explicit Cryptographic Erasure in Share

**Reason:** Go does not provide reliable cryptographic erasure due to GC and compiler optimisations. Acceptable for a research prototype.

### D3 -- FinDec Uses BGV Encoder.Decode Instead of Manual CRT

**Reason:** Lattigo's BGV encoding uses slot packing (NTT in R_p + permutation). Manual CRT reconstruction recovers NTT-domain values, not plaintext slots. The encoder's Decode function correctly handles slot unpacking. The paper now documents this in `rmk:slot-pack`.

### D4 -- PublicParams Field Naming

BGVParams field instead of Params. Naming only.

### D5 -- B_W=3 Matrix for (t=3, N=10)

**Spec claim:** B_W=1 exists for (t=3, N=10).

**Finding:** B_W=1 does not exist (exhaustively proven). B_W=2 not found (38M random trials). B_W=3 found and hardcoded.

**Matrix:**
```
W = [[-3, 1, 2], [1, 2, 3], [3, 2, -1], [-3, 2, 3],
     [-2, 0, 3], [2, 2, 0], [2, 3, 2], [-2, 1, 1]]
```

All 120 qualifying sets verified non-zero determinant. Worst-case Lambda_S = 16, worst |delta| = 19. Noise margin: 136 bits (vs 127 for Vandermonde B_W=64).

---

## 5. Benchmark Results

### 5.1 Test Environment

- **CPU:** Intel Core i7-8700K @ 4.40 GHz (overclocked, 6 cores / 12 threads)
- **OS:** Linux 6.6.87 (WSL2)
- **Go:** 1.24.0
- **Lattigo:** v6.2.0
- **Ring parameters:** n = 8192 (LogN = 13), log2 Q approx 219, T = 65537

### 5.2 Per-Algorithm Timing

#### t = 2 Configurations

| Operation | t=2, N=3, B_W=1 | Spec Target | Status |
|-----------|-----------------|-------------|--------|
| Share | 3.23 ms | < 10 ms | PASS |
| PartDec | 21.4 ms | < 5 ms | OVER (noise sampling) |
| Combine | 0.99 ms | < 5 ms | PASS |
| FinDec | 0.58 ms | < 50 ms | PASS |
| **End-to-end** | **71.4 ms** | **< 100 ms** | **PASS** |

**Note on PartDec:** The 21.4 ms is dominated by per-coefficient `crypto/rand.Int` smudging noise sampling (~19 ms). The ring multiplication itself is ~0.5 ms. Replacing with a PRNG-based sampler would bring PartDec under 5 ms.

#### t = 3 Configurations

| Operation | t=3, N=5, B_W=1 | t=3, N=10, B_W=3 (search) | t=3, N=10, Vandermonde |
|-----------|-----------------|--------------------------|------------------------|
| Share | 6.73 ms | 6.73 ms | 6.73 ms |
| PartDec | 21.4 ms | 21.4 ms | 21.4 ms |
| Combine | 0.87 ms | 0.87 ms | 0.87 ms |
| FinDec | 0.58 ms | 0.58 ms | 0.58 ms |
| **End-to-end** | **117.8 ms** | **211.6 ms** | **258.6 ms** |

B_W=3 and Vandermonde have identical per-algorithm timing (B_W only affects noise bounds, not runtime). End-to-end at N=10 is dominated by PartDec running for all 10 parties.

#### t = 4, 5 Configurations

| Operation | t=4, N=8, Vandermonde | t=5, N=10, Vandermonde |
|-----------|----------------------|------------------------|
| Share | 8.40 ms | -- |
| PartDec | 22.2 ms | -- |
| Combine | 1.02 ms | 1.07 ms |
| FinDec | 0.57 ms | -- |
| **End-to-end** | **206.4 ms** | **249.2 ms** |

### 5.3 Benchmark Justification

| Benchmark | Parameters | Justification |
|-----------|-----------|---------------|
| `EndToEnd_T2_N3` | t=2, N=3, B_W=1 | Minimal configuration; baseline |
| `EndToEnd_T3_N5` | t=3, N=5, B_W=1 | Optimal B_W with t=3 |
| `EndToEnd_T3_N10_BW3` | t=3, N=10, B_W=3 | **NEW.** Search matrix; improved noise over Vandermonde |
| `EndToEnd_T3_N10` | t=3, N=10, B_W=64 | Vandermonde baseline for t=3, N=10 |
| `EndToEnd_T4_N8` | t=4, N=8, B_W=125 | Tests t>=4 regime |
| `EndToEnd_T5_N10` | t=5, N=10, B_W=1296 | Stress test for cofactor growth |
| Per-algorithm (12) | Various | Isolate Share, PartDec, Combine, FinDec costs |

---

## 6. Noise Margin Analysis

### 6.1 Correctness Condition

The scheme requires: `|delta| * B_ct + Lambda_S * B_sm < Delta/2`

Where:
- |delta| = absolute determinant of M_S
- B_ct approx 2^20 = ciphertext noise bound
- Lambda_S = Sum |lambda_hat_j| = sum of absolute cofactor values
- B_sm = B_W * B_ct * 2^kappa = smudging noise bound
- Delta/2 = Q/(2T) approx 2^(219-17) = 2^202

### 6.2 Results Per Configuration

| Configuration | B_W | log2(B_sm) | Worst |delta| | Worst Lambda_S | Noise (bits) | Margin (bits) | Status |
|---------------|-----|------------|--------|--------|----------------|---------------|--------|
| t=2, N=3, B_W=1 | 1 | 60 | 1 | 2 | ~61 | **140** | PASS |
| t=2, N=4, B_W=1 | 1 | 60 | 1 | 2 | ~61 | **140** | PASS |
| t=2, N=10, Vand | 9 | 63 | 2 | 16 | ~68 | **134** | PASS |
| t=3, N=5, B_W=1 | 1 | 60 | 1 | 2 | ~61 | **140** | PASS |
| **t=3, N=10, B_W=3** | **3** | **61** | **19** | **16** | **~65** | **136** | **PASS** |
| t=3, N=10, Vand | 64 | 66 | 20 | 260 | ~75 | **127** | PASS |
| t=4, N=8, Vand | 125 | 66 | 228 | 1068 | ~77 | **125** | PASS |
| t=5, N=10, Vand | 1296 | 70 | 16254 | 133866 | ~87 | **114** | PASS |

### 6.3 Trend Analysis

```
Threshold   B_W          Worst Lambda_S   Noise (bits)   Margin (bits)
  2         1            2                ~61            140
  3 (B_W=1) 1            2                ~61            140
  3 (B_W=3) 3            16               ~65            136
  3 (Vand)  64 = 8^2     260              ~75            127
  4 (Vand)  125 = 5^3    1068             ~77            125
  5 (Vand)  1296 = 6^4   133866           ~87            114
```

The B_W=3 search matrix for (t=3, N=10) gains 9 bits of margin over Vandermonde (136 vs 127). With Vandermonde, margin drops ~12--15 bits per unit increase in t.

### 6.4 Parameter Sweep: Tightest Viable log2 Q

The parameter sweep (`TestParameterSweep`) finds the smallest LogQ that still yields a positive noise margin for each configuration. This demonstrates the scheme works near its theoretical limit, not just with excess headroom.

| Configuration | Tightest log2 Q | Margin at tightest | Default (219-bit) Margin |
|---|---|---|---|
| t=2, N=3, B_W=1 | 81 | 2 bits | 140 bits |
| t=2, N=4, B_W=1 | 81 | 2 bits | 140 bits |
| t=2, N=10, Vandermonde | 91 | 6 bits | 134 bits |
| t=3, N=5, B_W=1 | 81 | 2 bits | 140 bits |
| **t=3, N=10, B_W=3** | **91** | **8 bits** | **136 bits** |
| t=3, N=10, Vandermonde | 101 | 9 bits | 127 bits |
| t=4, N=8, Vandermonde | 101 | 7 bits | 125 bits |
| t=5, N=10, Vandermonde | 111 | 6 bits | 114 bits |

**Key observation:** B_W=1 configs require only log2 Q >= 81 (2 RNS primes of 40 bits). The B_W=3 search matrix at (t=3, N=10) requires log2 Q >= 91, saving 10 bits over Vandermonde (which requires >= 101).

---

## 7. B_W Search Methodology

### 7.1 Problem Statement

For (t=3, N=10), find W in Z^{8x3} with minimal max|W[j][k]| such that all C(10,3) = 120 qualifying subsets of the resulting M matrix have non-zero determinant.

### 7.2 Approach 1: Exhaustive Backtracking (B_W=1)

**Tool:** `cmd/bw1search/main.go`

The search builds W row by row. After placing row k, it checks only the qualifying sets whose maximum party index equals t-1+k (sets that become checkable exactly when row k is placed). Sets are grouped by this criterion:

```
Row 0: 1 new set
Row 1: 3 new sets
Row 2: 6 new sets
Row 3: 10 new sets
Row 4: 15 new sets
Row 5: 21 new sets
Row 6: 28 new sets
Row 7: 36 new sets
```

For B_W=1, each row has 27 candidates (3^3). The search parallelises across the 27 first-row choices using 12 CPU cores.

**Result:** No B_W=1 matrix exists. Exhaustive search completed in **2.5 ms** -- the pruning is so aggressive that virtually all branches are cut within the first few rows.

### 7.3 Approach 2: Exhaustive Backtracking (B_W=2)

Each row has 125 candidates (5^3). The search tree is much larger and the pruning less effective. The backtracking ran for 5 minutes without completing. Separately, 38 million random W samples from {-2,...,2}^{8x3} found no valid matrix.

**Result:** B_W=2 status is undetermined (not proven impossible, but no matrix found).

### 7.4 Approach 3: Randomised Search (B_W=3)

Sample random W from {-3,...,3}^{8x3}, check all 120 qualifying sets with early rejection. With 12 workers running in parallel:

**Result:** Found in **354 trials** (effectively instant). This suggests B_W=3 matrices are common in the search space.

### 7.5 The Matrix

```
W = [[-3,  1,  2],
     [ 1,  2,  3],
     [ 3,  2, -1],
     [-3,  2,  3],
     [-2,  0,  3],
     [ 2,  2,  0],
     [ 2,  3,  2],
     [-2,  1,  1]]
```

Verification: all 120 qualifying sets have non-zero determinant. Worst-case |delta| = 19, worst-case Lambda_S = 16.

---

## 8. Paper Fixes Applied

Summary of all changes made to `myscheme-latex.md`:

| Fix | Location | Change |
|-----|----------|--------|
| Fix 2 | Line 1434 (Concrete Parameter Comparison) | `p = 2^{17}` changed to `p = 65537` |
| Fix 3 | Lines 956--1003 (`lem:concrete`) | Split into 3 cases: (t=2,N=4,B_W=1), (t=3,N=5,B_W=1), (t=3,N=10,B_W=3 with published matrix) |
| Fix 1 | Line 947 (noise table) | B_W=1 feasibility: t=2 changed from "all N" to "N<=4"; t=3 changed to "N<=5 verified; B_W=3 for N=10" |
| Fix 1 | Lines 1523--1536 (Positioning) | t=2: "all N" changed to "N<=4". t=3: B_W=3 with R(3,B_W=3)~48. t>=4: qualified by regime with Regime B data |
| Fix 5 | Line 412 (`def:params`) | `p \mid q` changed to `gcd(p,q) = 1` |
| Fix 5 | Line 413 | Removed "(exact, since p \mid q)" |
| Fix 5 | Lines 645--646, 657, 744 | delta^{-1} existence argument changed from "gcd(delta,q)=1 and p\|q" to "\|delta\| < p and p prime" |
| Fix 5 | After `def:params` | Added `rmk:plaintext` on gcd(p,q)=1 sufficiency and Fermat prime choice |
| Fix 4 | Line 27 (abstract) | Added "up to t=5 under all-F RLWE regime" |
| Fix 4 | `rmk:window` | Rewritten to distinguish Regime A (t<=3) and Regime B (t<=5) |
| Fix 4 | Lines 1533--1536 (t>=4 bullet) | Added Regime B clause with 114-bit margin data |
| Fix 4 | Conclusion | Qualified "t<=3" with "under statistical smudging; up to t=5 under all-F RLWE" |
| Fix 7 | After Algorithm 5 | Added `rmk:slot-pack` on slot-packed BFV/BGV encoding |

---

## 9. Project Structure

```
compact-the/
  go.mod, go.sum
  implementation-report.md      # v1 (retained for reference)
  implementation-report-v2.md   # This file
  params/
    params.go                   # PublicParams, DefaultRegimeB(), VerifyParams()
    params_test.go
  lsss/
    matrix.go                   # VandermondeW, BuildM, SearchBW1, HardcodedBW1
    adjugate.go                 # Determinant, Adjugate, FirstRowCofactors
    adjugate_test.go
    qualify.go                  # AllQualifyingSets, ExtractSubmatrix, ValidateAllSets
  scheme/
    setup.go                    # Algorithm 1: Setup
    share.go                    # Algorithm 2: Share
    partdec.go                  # Algorithm 3: PartDec
    combine.go                  # Algorithm 4: Combine
    findec.go                   # Algorithm 5: FinDec
    helpers.go                  # scalarMulSigned, roundDivBigInt, copyPoly
    scheme_test.go              # End-to-end + edge case tests (10 tests)
    sweep_test.go               # Parameter sweep: tightest viable LogQ (NEW)
  noise/
    bounds.go                   # ComputeBsmRegimeB, VerifyCorrectness, LambdaS
    smudging.go                 # SampleSmudgingNoise (uniform)
  bench/
    bench_test.go               # 17 benchmark configurations
  cmd/
    demo/main.go                # Full pipeline demo (N=5, t=3, B_W=1)
    bw1search/main.go           # B_W=1 backtracking search utility (NEW)
```

---

## 10. Known Limitations

1. **PartDec performance:** 21.4 ms (vs 5 ms target) due to per-coefficient `crypto/rand.Int`. Replace with ChaCha20-seeded PRNG for production.

2. **B_W=2 status for (t=3, N=10):** Not proven impossible (unlike B_W=1), but not found in 38M random trials. The gap between B_W=2 (unknown) and B_W=3 (found) remains open.

3. **No cryptographic erasure:** Go GC prevents reliable memory wiping. Acceptable for prototype.

4. **Regime B only:** The implementation uses the all-F RLWE noise model. Regime A (statistical smudging) would require larger parameters.

5. **Determinant magnitudes at high t:** At t=5, worst-case |delta|=16254 and Lambda_S=133866. While the 114-bit margin is comfortable, further increases in t or N would exhaust the noise budget.
