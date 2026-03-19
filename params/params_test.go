package params

import (
	"math/big"
	"testing"
)

func TestDefaultRegimeB(t *testing.T) {
	params, err := DefaultRegimeB()
	if err != nil {
		t.Fatalf("DefaultRegimeB failed: %v", err)
	}

	// Check ring degree
	n := params.N()
	if n != 8192 {
		t.Errorf("expected N=8192, got %d", n)
	}

	// Check plaintext modulus
	p := params.PlaintextModulus()
	if p != 65537 {
		t.Errorf("expected PlaintextModulus=65537, got %d", p)
	}

	// Check gcd(p, Q) = 1
	pBig := new(big.Int).SetUint64(p)
	qBig := params.RingQ().ModulusAtLevel[params.MaxLevel()]
	g := new(big.Int).GCD(nil, nil, pBig, qBig)
	if g.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("gcd(p, Q) = %s, want 1", g.String())
	}

	// Check logQ ≈ 218
	logQ := qBig.BitLen()
	if logQ < 200 || logQ > 240 {
		t.Errorf("logQ = %d, expected ≈218", logQ)
	}
	t.Logf("N=%d, logQ=%d, p=%d, maxLevel=%d", n, logQ, p, params.MaxLevel())
}

func TestVerifyParams(t *testing.T) {
	params, err := DefaultRegimeB()
	if err != nil {
		t.Fatal(err)
	}

	if err := VerifyParams(params, 10); err != nil {
		t.Errorf("VerifyParams failed for N=10: %v", err)
	}

	// Should fail if numParties >= p
	if err := VerifyParams(params, 70000); err == nil {
		t.Error("expected error for numParties >= p")
	}
}
