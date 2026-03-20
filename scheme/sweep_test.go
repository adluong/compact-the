package scheme

import (
	"testing"

	"compact-the/lsss"
	"compact-the/noise"
	"compact-the/params"
)

// TestParameterSweep finds the tightest LogQ configuration that still provides
// a positive noise margin for each (t, N, B_W) configuration.
func TestParameterSweep(t *testing.T) {
	logQConfigs := []struct {
		name string
		logQ []int
		logP []int
	}{
		{"218 (4×~55)", []int{55, 55, 55, 53}, []int{55}},
		{"200 (4×50)", []int{50, 50, 50, 50}, []int{50}},
		{"180 (3×60)", []int{60, 60, 60}, []int{55}},
		{"165 (3×55)", []int{55, 55, 55}, []int{55}},
		{"150 (3×50)", []int{50, 50, 50}, []int{50}},
		{"140 (2×55+30)", []int{55, 55, 30}, []int{55}},
		{"130 (2×55+20)", []int{55, 55, 20}, []int{55}},
		{"120 (2×60)", []int{60, 60}, []int{55}},
		{"110 (2×55)", []int{55, 55}, []int{55}},
		{"100 (2×50)", []int{50, 50}, []int{50}},
		{"90 (2×45)", []int{45, 45}, []int{45}},
		{"80 (2×40)", []int{40, 40}, []int{40}},
		{"60 (1×60)", []int{60}, []int{55}},
	}

	configs := []struct {
		name string
		N, T int
		bw   int
	}{
		{"t=2, N=3, B_W=1", 3, 2, 1},
		{"t=2, N=4, B_W=1", 4, 2, 1},
		{"t=2, N=10, Vand", 10, 2, 0},
		{"t=3, N=5, B_W=1", 5, 3, 1},
		{"t=3, N=10, B_W=3", 10, 3, 1},
		{"t=3, N=10, Vand", 10, 3, 0},
		{"t=4, N=8, Vand", 8, 4, 0},
		{"t=5, N=10, Vand", 10, 5, 0},
	}

	t.Log("=== Parameter Sweep: tightest valid LogQ per config ===")

	for _, cfg := range configs {
		// Find the smallest LogQ that still works
		bestLogQ := ""
		bestMargin := -1
		bestLogQBits := 0
		bestBW := 0
		bestBsm := 0

		for i := len(logQConfigs) - 1; i >= 0; i-- {
			lq := logQConfigs[i]
			bgvParams, err := params.NewBGVParams(13, lq.logQ, lq.logP, 65537)
			if err != nil {
				continue
			}
			pp, err := Setup(bgvParams, cfg.N, cfg.T, cfg.bw, 40)
			if err != nil {
				continue
			}

			logQ := bgvParams.RingQ().ModulusAtLevel[bgvParams.MaxLevel()].BitLen()
			logT := 16
			halfDeltaLog2 := logQ - logT - 1

			sets := lsss.AllQualifyingSets(cfg.N, cfg.T)
			worstMargin := halfDeltaLog2
			allPass := true

			for _, S := range sets {
				MS := lsss.ExtractSubmatrix(pp.M, S)
				det, cofactors := lsss.FirstRowCofactors(MS, cfg.T)
				if det == 0 {
					allPass = false
					break
				}
				absDet := det
				if absDet < 0 {
					absDet = -absDet
				}
				lambdaS := lsss.LambdaS(cofactors)

				ok := noise.VerifyCorrectness(absDet, lambdaS, pp.Bsm, 20, logQ, logT)
				if !ok {
					allPass = false
					break
				}

				noiseBits := 0
				l := lambdaS
				for l > 0 {
					noiseBits++
					l >>= 1
				}
				margin := halfDeltaLog2 - (noiseBits + pp.Bsm.BitLen())
				if margin < worstMargin {
					worstMargin = margin
				}
			}

			if allPass && worstMargin >= 0 {
				bestLogQ = lq.name
				bestMargin = worstMargin
				bestLogQBits = logQ
				bestBW = pp.BW
				bestBsm = pp.Bsm.BitLen()
				break // Found the smallest valid LogQ
			}
		}

		if bestLogQ != "" {
			t.Logf("%s: tightest logQ≈%s (%d bits), margin=%d bits, B_W=%d, log₂Bsm=%d",
				cfg.name, bestLogQ, bestLogQBits, bestMargin, bestBW, bestBsm)
		} else {
			t.Logf("%s: no valid configuration found", cfg.name)
		}
	}

	// Also show default config margins
	t.Log("")
	t.Log("=== Default config (logQ≈218) margins ===")
	bgvParams, _ := params.DefaultRegimeB()
	logQ := bgvParams.RingQ().ModulusAtLevel[bgvParams.MaxLevel()].BitLen()
	logT := 16
	halfDeltaLog2 := logQ - logT - 1
	t.Logf("logQ=%d, log₂(Δ/2)=%d", logQ, halfDeltaLog2)

	for _, cfg := range configs {
		pp, err := Setup(bgvParams, cfg.N, cfg.T, cfg.bw, 40)
		if err != nil {
			t.Logf("%s: SETUP FAILED: %v", cfg.name, err)
			continue
		}
		sets := lsss.AllQualifyingSets(cfg.N, cfg.T)
		worstMargin := halfDeltaLog2
		for _, S := range sets {
			MS := lsss.ExtractSubmatrix(pp.M, S)
			_, cofactors := lsss.FirstRowCofactors(MS, cfg.T)
			lambdaS := lsss.LambdaS(cofactors)
			noiseBits := 0
			l := lambdaS
			for l > 0 {
				noiseBits++
				l >>= 1
			}
			margin := halfDeltaLog2 - (noiseBits + pp.Bsm.BitLen())
			if margin < worstMargin {
				worstMargin = margin
			}
		}
		t.Logf("%s: margin=%d bits, B_W=%d, log₂Bsm=%d",
			cfg.name, worstMargin, pp.BW, pp.Bsm.BitLen())
	}
}
