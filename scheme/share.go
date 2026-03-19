package scheme

import (
	"compact-the/params"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/utils/sampling"
)

// Share implements Algorithm 2: Share(pp, sk) → (s_1, ..., s_N).
// Computes shares s_i = Σ_k M[i][k] · ρ_k where ρ = (sk, r_1, ..., r_{t-1}).
// All shares are returned in coefficient (non-NTT) form.
func Share(pp *params.PublicParams, sk *rlwe.SecretKey) ([]ring.Poly, error) {
	ringQ := pp.BGVParams.RingQ()
	t := pp.T
	N := pp.N

	// The secret key polynomial may be in NTT+Montgomery form.
	// Convert to coefficient form for our operations.
	skPoly := ringQ.NewPoly()
	skPoly.CopyLvl(ringQ.Level(), sk.Value.Q)
	if sk.Value.Q.N() > 0 {
		// Secret keys in Lattigo are stored in NTT form by default.
		// We need coefficient form for scalar-ring multiplication.
		ringQ.INTT(skPoly, skPoly)
		ringQ.IMForm(skPoly, skPoly)
	}

	// Sample t-1 uniform random polynomials r_1, ..., r_{t-1}
	prng, err := sampling.NewPRNG()
	if err != nil {
		return nil, err
	}
	sampler := ring.NewUniformSampler(prng, ringQ)

	rho := make([]ring.Poly, t) // ρ = (sk, r_1, ..., r_{t-1})
	rho[0] = skPoly
	for k := 1; k < t; k++ {
		rho[k] = sampler.ReadNew()
	}

	// Compute shares
	shares := make([]ring.Poly, N)
	tmp := ringQ.NewPoly()

	for i := 0; i < N; i++ {
		shares[i] = ringQ.NewPoly()

		if i < t-1 {
			// F-party: s_i = r_{i+1} (M[i] is unit vector e_{i+1})
			shares[i].CopyLvl(ringQ.Level(), rho[i+1])
		} else {
			// L-party: s_i = Σ_k M[i][k] · ρ_k = -Σ_k W[j][k] · ρ_k
			j := i - (t - 1) // W-row index
			for k := 0; k < t; k++ {
				w := pp.W[j][k]
				if w == 0 {
					continue
				}
				// M[i][k] = -W[j][k], so we compute -W[j][k] * ρ[k]
				scalarMulSigned(ringQ, rho[k], -w, tmp)
				ringQ.Add(shares[i], tmp, shares[i])
			}
		}
	}

	return shares, nil
}
