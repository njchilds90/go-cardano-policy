package policy_test

import (
	"encoding/json"
	"strings"
	"testing"

	policy "github.com/njchilds90/go-cardano-policy"
)

const validKeyHash = "e09d36c79dec9bd1b3d9e152247701cd0bb0f68e8cd748d2a058bcd"

func TestNewSigScript(t *testing.T) {
	tests := []struct {
		name    string
		keyHash string
		wantErr bool
	}{
		{"valid", validKeyHash, false},
		{"too short", "abc", true},
		{"too long", validKeyHash + "00", true},
		{"uppercase rejected", strings.ToUpper(validKeyHash), true},
		{"non-hex", "gggggggggggggggggggggggggggggggggggggggggggggggggggggggg", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := policy.NewSigScript(tt.keyHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSigScript(%q) error = %v, wantErr %v", tt.keyHash, err, tt.wantErr)
			}
		})
	}
}

func TestNewAllScript_RequiresScripts(t *testing.T) {
	_, err := policy.NewAllScript()
	if err == nil {
		t.Error("expected error for empty all-script")
	}
}

func TestNewAtLeastScript_Validation(t *testing.T) {
	sig, _ := policy.NewSigScript(validKeyHash)

	tests := []struct {
		name    string
		n       int
		scripts []policy.Script
		wantErr bool
	}{
		{"valid 1-of-1", 1, []policy.Script{sig}, false},
		{"valid 2-of-3", 2, []policy.Script{sig, sig, sig}, false},
		{"n=0", 0, []policy.Script{sig}, true},
		{"n > count", 3, []policy.Script{sig, sig}, true},
		{"no scripts", 1, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := policy.NewAtLeastScript(tt.n, tt.scripts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAtLeastScript(%d, ...) error = %v, wantErr %v", tt.n, err, tt.wantErr)
			}
		})
	}
}

func TestBeforeAfterScript(t *testing.T) {
	_, err := policy.NewBeforeScript(0)
	if err == nil {
		t.Error("expected error for slot=0 before-script")
	}
	_, err = policy.NewAfterScript(0)
	if err == nil {
		t.Error("expected error for slot=0 after-script")
	}

	b, err := policy.NewBeforeScript(100)
	if err != nil {
		t.Fatalf("NewBeforeScript: %v", err)
	}
	if !policy.IsTimeLocked(b) {
		t.Error("expected IsTimeLocked=true for before-script")
	}
}

func TestToJSONFromJSON(t *testing.T) {
	sig, _ := policy.NewSigScript(validKeyHash)
	before, _ := policy.NewBeforeScript(50000000)
	all, _ := policy.NewAllScript(sig, before)

	jsonBytes, err := policy.ToJSON(all)
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}

	// Must be valid JSON
	var raw map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &raw); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}

	// Round-trip
	parsed, err := policy.FromJSON(jsonBytes)
	if err != nil {
		t.Fatalf("FromJSON: %v", err)
	}

	// Policy IDs must match after round-trip
	id1, _ := policy.Compute(all)
	id2, _ := policy.Compute(parsed)
	if id1 != id2 {
		t.Errorf("policy ID changed after JSON round-trip: %s vs %s", id1, id2)
	}
}

func TestValidate_UnknownType(t *testing.T) {
	s := policy.Script{Type: "bogus"}
	if err := policy.Validate(s); err == nil {
		t.Error("expected error for unknown type")
	}
}

func TestIsTimeLocked(t *testing.T) {
	sig, _ := policy.NewSigScript(validKeyHash)
	if policy.IsTimeLocked(sig) {
		t.Error("sig script should not be time locked")
	}

	after, _ := policy.NewAfterScript(1000)
	any_, _ := policy.NewAnyScript(sig, after)
	if !policy.IsTimeLocked(any_) {
		t.Error("any(sig, after) should be time locked")
	}
}

func TestMustSigScript_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	policy.MustSigScript("bad")
}
