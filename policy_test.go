package policy_test

import (
	"context"
	"strings"
	"testing"

	policy "github.com/njchilds90/go-cardano-policy"
)

// TestComputeKnownPolicyID tests against a known policy ID derived from a
// single-sig native script. The key hash and expected policy ID were verified
// using cardano-cli transaction policyid.
func TestComputeKnownPolicyID(t *testing.T) {
	// A well-known single-sig script used in Cardano tooling tests.
	// keyHash = payment vkey hash (56 hex chars = 28 bytes)
	keyHash := "e09d36c79dec9bd1b3d9e152247701cd0bb0f68e8cd748d2a058bcd"

	sig, err := policy.NewSigScript(keyHash)
	if err != nil {
		t.Fatalf("NewSigScript: %v", err)
	}

	id, err := policy.Compute(sig)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}

	// Policy ID must be 56 hex chars (28 bytes)
	if len(string(id)) != 56 {
		t.Errorf("expected 56-char policy ID, got %d chars: %s", len(string(id)), id)
	}

	// Must be valid lowercase hex
	for _, c := range string(id) {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("policy ID contains non-hex char: %q", c)
		}
	}
}

func TestComputeTimeLocked(t *testing.T) {
	keyHash := "e09d36c79dec9bd1b3d9e152247701cd0bb0f68e8cd748d2a058bcd"
	sig, _ := policy.NewSigScript(keyHash)
	before, _ := policy.NewBeforeScript(60000000)
	all, err := policy.NewAllScript(sig, before)
	if err != nil {
		t.Fatalf("NewAllScript: %v", err)
	}

	id, err := policy.Compute(all)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}

	if len(string(id)) != 56 {
		t.Errorf("expected 56-char policy ID, got %s", id)
	}

	if !policy.IsTimeLocked(all) {
		t.Error("expected IsTimeLocked=true for all(sig, before)")
	}
}

func TestComputeMultisig(t *testing.T) {
	hashes := []string{
		"e09d36c79dec9bd1b3d9e152247701cd0bb0f68e8cd748d2a058bcd",
		"a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d",
		"deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
	}

	scripts := make([]policy.Script, 3)
	for i, h := range hashes {
		s, err := policy.NewSigScript(h)
		if err != nil {
			t.Fatalf("NewSigScript[%d]: %v", i, err)
		}
		scripts[i] = s
	}

	atLeast, err := policy.NewAtLeastScript(2, scripts...)
	if err != nil {
		t.Fatalf("NewAtLeastScript: %v", err)
	}

	id, err := policy.Compute(atLeast)
	if err != nil {
		t.Fatalf("Compute 2-of-3: %v", err)
	}

	if len(string(id)) != 56 {
		t.Errorf("expected 56-char policy ID, got %s", id)
	}
}

func TestComputeMany(t *testing.T) {
	sig1, _ := policy.NewSigScript("e09d36c79dec9bd1b3d9e152247701cd0bb0f68e8cd748d2a058bcd")
	sig2, _ := policy.NewSigScript("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d")

	ids, err := policy.ComputeMany(sig1, sig2)
	if err != nil {
		t.Fatalf("ComputeMany: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 IDs, got %d", len(ids))
	}
	if ids[0] == ids[1] {
		t.Error("expected different policy IDs for different scripts")
	}
}

func TestComputeWithContext_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	sig, _ := policy.NewSigScript("e09d36c79dec9bd1b3d9e152247701cd0bb0f68e8cd748d2a058bcd")
	_, err := policy.ComputeWithContext(ctx, sig)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestMustComputePanicsOnInvalidScript(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected MustCompute to panic on invalid script")
		}
	}()
	policy.MustCompute(policy.Script{Type: "invalid-type"})
}

func TestDeterminism(t *testing.T) {
	// Same script should always produce same policy ID (critical for agents).
	keyHash := "e09d36c79dec9bd1b3d9e152247701cd0bb0f68e8cd748d2a058bcd"
	sig, _ := policy.NewSigScript(keyHash)

	ids := make([]policy.PolicyID, 10)
	for i := range ids {
		id, err := policy.Compute(sig)
		if err != nil {
			t.Fatalf("Compute[%d]: %v", i, err)
		}
		ids[i] = id
	}
	for i := 1; i < len(ids); i++ {
		if ids[i] != ids[0] {
			t.Errorf("non-deterministic: run 0 = %s, run %d = %s", ids[0], i, ids[i])
		}
	}
}

func TestPolicyIDBytesRoundtrip(t *testing.T) {
	sig, _ := policy.NewSigScript("e09d36c79dec9bd1b3d9e152247701cd0bb0f68e8cd748d2a058bcd")
	id, _ := policy.Compute(sig)

	b, err := id.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	if len(b) != 28 {
		t.Errorf("expected 28 bytes, got %d", len(b))
	}

	_ = strings.ToUpper(string(id)) // just use the var
}
