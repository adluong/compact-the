package scheme

import (
	"fmt"

	"compact-the/lsss"
	"compact-the/noise"
	"compact-the/params"

	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

// Setup implements Algorithm 1: Setup(1^λ, t, N, B_W) → pp.
// It constructs the public parameters including the share-generation matrix M.
// If bw == 1, attempts to use a B_W=1 matrix; otherwise uses Vandermonde.
func Setup(bgvParams bgv.Parameters, numParties, threshold, bw, kappa int) (*params.PublicParams, error) {
	if threshold < 2 {
		return nil, fmt.Errorf("threshold must be ≥ 2, got %d", threshold)
	}
	if threshold > numParties {
		return nil, fmt.Errorf("threshold %d > number of parties %d", threshold, numParties)
	}
	if threshold > 3 {
		fmt.Printf("WARNING: threshold %d > 3 — cofactor bounds grow super-exponentially; noise margin may be insufficient\n", threshold)
	}

	// Verify BGV parameters
	if err := params.VerifyParams(bgvParams, numParties); err != nil {
		return nil, fmt.Errorf("parameter verification failed: %w", err)
	}

	// Construct W matrix
	var W [][]int64
	var err error
	if bw == 1 {
		// bw=1 requests "search for smallest B_W matrix available"
		W, err = lsss.SearchBW1(numParties, threshold)
		if err != nil {
			// Fall back to Vandermonde
			W = lsss.VandermondeW(numParties, threshold)
		}
		bw = int(lsss.MaxAbsEntry(W))
	} else {
		W = lsss.VandermondeW(numParties, threshold)
		bw = int(lsss.MaxAbsEntry(W))
	}

	// Build M from W
	M := lsss.BuildM(W, numParties, threshold)

	// Validate all qualifying sets
	sets := lsss.AllQualifyingSets(numParties, threshold)
	qModuli := bgvParams.RingQ().ModuliChain()[:bgvParams.MaxLevel()+1]
	if err := lsss.ValidateAllSets(M, threshold, sets, qModuli); err != nil {
		return nil, fmt.Errorf("qualifying set validation failed: %w", err)
	}

	// Compute B_sm for Regime B
	bctLog2 := 20 // approximate ciphertext noise bound
	bsmLog2 := noise.ComputeBsmRegimeB(bw, bctLog2, kappa)

	pp := &params.PublicParams{
		BGVParams: bgvParams,
		N:         numParties,
		T:         threshold,
		BW:        bw,
		M:         M,
		W:         W,
		Kappa:     kappa,
		BsmLog2:   bsmLog2,
	}

	return pp, nil
}
