// Package policy provides zero-dependency utilities for Cardano native script
// construction and policy ID computation, compatible with cardano-cli JSON format.
//
// Supported script types follow the Cardano Simple Script specification:
// https://github.com/input-output-hk/cardano-node/blob/master/doc/reference/simple-scripts.md
package policy

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ScriptType identifies the kind of native script clause.
type ScriptType string

const (
	// ScriptTypeSig requires a signature from a specific key hash.
	ScriptTypeSig ScriptType = "sig"
	// ScriptTypeAll requires ALL sub-scripts to be satisfied.
	ScriptTypeAll ScriptType = "all"
	// ScriptTypeAny requires at least ONE sub-script to be satisfied.
	ScriptTypeAny ScriptType = "any"
	// ScriptTypeAtLeast requires at least N sub-scripts to be satisfied.
	ScriptTypeAtLeast ScriptType = "atLeast"
	// ScriptTypeBefore requires the transaction to be submitted before a slot.
	ScriptTypeBefore ScriptType = "before"
	// ScriptTypeAfter requires the transaction to be submitted after a slot.
	ScriptTypeAfter ScriptType = "after"
)

// Script represents a Cardano native script (simple script).
// It maps directly to the JSON format accepted by cardano-cli.
type Script struct {
	Type     ScriptType `json:"type"`
	KeyHash  string     `json:"keyHash,omitempty"`
	Required int        `json:"required,omitempty"`
	Slot     uint64     `json:"slot,omitempty"`
	Scripts  []Script   `json:"scripts,omitempty"`
}

// ErrInvalidScript is returned when a script fails structural validation.
type ErrInvalidScript struct {
	Type   ScriptType
	Reason string
}

func (e *ErrInvalidScript) Error() string {
	return fmt.Sprintf("invalid script (type=%s): %s", e.Type, e.Reason)
}

// NewSigScript creates a script that requires a signature from the given key hash.
// keyHash must be a 56-character lowercase hex string (28 bytes).
//
// Example:
//
//	s, err := policy.NewSigScript("e09d36c79dec9bd1b3d9e152247701cd0bb0f68e8cd748d2a05")
func NewSigScript(keyHash string) (Script, error) {
	if err := validateKeyHash(keyHash); err != nil {
		return Script{}, err
	}
	return Script{Type: ScriptTypeSig, KeyHash: keyHash}, nil
}

// MustSigScript creates a sig script, panicking on invalid input.
// Useful for initializing package-level or test constants.
func MustSigScript(keyHash string) Script {
	s, err := NewSigScript(keyHash)
	if err != nil {
		panic(err)
	}
	return s
}

// NewAllScript creates a script requiring ALL sub-scripts to be satisfied.
//
// Example:
//
//	s, err := policy.NewAllScript(sig1, timeLock)
func NewAllScript(scripts ...Script) (Script, error) {
	if len(scripts) == 0 {
		return Script{}, &ErrInvalidScript{Type: ScriptTypeAll, Reason: "must have at least one sub-script"}
	}
	return Script{Type: ScriptTypeAll, Scripts: scripts}, nil
}

// NewAnyScript creates a script requiring at least ONE sub-script to be satisfied.
func NewAnyScript(scripts ...Script) (Script, error) {
	if len(scripts) == 0 {
		return Script{}, &ErrInvalidScript{Type: ScriptTypeAny, Reason: "must have at least one sub-script"}
	}
	return Script{Type: ScriptTypeAny, Scripts: scripts}, nil
}

// NewAtLeastScript creates a script requiring at least n sub-scripts to be satisfied.
//
// Example:
//
//	s, err := policy.NewAtLeastScript(2, sig1, sig2, sig3) // 2-of-3 multisig
func NewAtLeastScript(n int, scripts ...Script) (Script, error) {
	if n <= 0 {
		return Script{}, &ErrInvalidScript{Type: ScriptTypeAtLeast, Reason: "required must be > 0"}
	}
	if len(scripts) == 0 {
		return Script{}, &ErrInvalidScript{Type: ScriptTypeAtLeast, Reason: "must have at least one sub-script"}
	}
	if n > len(scripts) {
		return Script{}, &ErrInvalidScript{Type: ScriptTypeAtLeast, Reason: fmt.Sprintf("required (%d) exceeds script count (%d)", n, len(scripts))}
	}
	return Script{Type: ScriptTypeAtLeast, Required: n, Scripts: scripts}, nil
}

// NewBeforeScript creates a time-lock script valid only before the given slot.
//
// Example:
//
//	s, err := policy.NewBeforeScript(60000000) // lock minting before slot 60000000
func NewBeforeScript(slot uint64) (Script, error) {
	if slot == 0 {
		return Script{}, &ErrInvalidScript{Type: ScriptTypeBefore, Reason: "slot must be > 0"}
	}
	return Script{Type: ScriptTypeBefore, Slot: slot}, nil
}

// NewAfterScript creates a time-lock script valid only after the given slot.
func NewAfterScript(slot uint64) (Script, error) {
	if slot == 0 {
		return Script{}, &ErrInvalidScript{Type: ScriptTypeAfter, Reason: "slot must be > 0"}
	}
	return Script{Type: ScriptTypeAfter, Slot: slot}, nil
}

// Validate checks that a Script is structurally valid.
// It recursively validates all sub-scripts.
func Validate(s Script) error {
	switch s.Type {
	case ScriptTypeSig:
		return validateKeyHash(s.KeyHash)
	case ScriptTypeAll, ScriptTypeAny:
		if len(s.Scripts) == 0 {
			return &ErrInvalidScript{Type: s.Type, Reason: "must have at least one sub-script"}
		}
		for i, sub := range s.Scripts {
			if err := Validate(sub); err != nil {
				return fmt.Errorf("sub-script[%d]: %w", i, err)
			}
		}
	case ScriptTypeAtLeast:
		if s.Required <= 0 {
			return &ErrInvalidScript{Type: s.Type, Reason: "required must be > 0"}
		}
		if s.Required > len(s.Scripts) {
			return &ErrInvalidScript{Type: s.Type, Reason: fmt.Sprintf("required (%d) exceeds script count (%d)", s.Required, len(s.Scripts))}
		}
		for i, sub := range s.Scripts {
			if err := Validate(sub); err != nil {
				return fmt.Errorf("sub-script[%d]: %w", i, err)
			}
		}
	case ScriptTypeBefore, ScriptTypeAfter:
		if s.Slot == 0 {
			return &ErrInvalidScript{Type: s.Type, Reason: "slot must be > 0"}
		}
	default:
		return &ErrInvalidScript{Type: s.Type, Reason: "unknown script type"}
	}
	return nil
}

// ToJSON serializes the script to the cardano-cli compatible JSON format.
//
// Example:
//
//	jsonBytes, err := policy.ToJSON(script)
func ToJSON(s Script) ([]byte, error) {
	if err := Validate(s); err != nil {
		return nil, err
	}
	return json.MarshalIndent(s, "", "  ")
}

// FromJSON deserializes a script from cardano-cli compatible JSON.
func FromJSON(data []byte) (Script, error) {
	var s Script
	if err := json.Unmarshal(data, &s); err != nil {
		return Script{}, fmt.Errorf("failed to parse script JSON: %w", err)
	}
	if err := Validate(s); err != nil {
		return Script{}, err
	}
	return s, nil
}

// validateKeyHash checks that a key hash string is a valid 56-char lowercase hex string.
func validateKeyHash(kh string) error {
	if len(kh) != 56 {
		return &ErrInvalidScript{
			Type:   ScriptTypeSig,
			Reason: fmt.Sprintf("keyHash must be 56 hex chars (28 bytes), got %d", len(kh)),
		}
	}
	for _, c := range kh {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return &ErrInvalidScript{
				Type:   ScriptTypeSig,
				Reason: fmt.Sprintf("keyHash contains invalid character: %q", c),
			}
		}
	}
	return nil
}

// IsTimeLocked returns true if the script (or any sub-script) contains a Before or After clause.
func IsTimeLocked(s Script) bool {
	if s.Type == ScriptTypeBefore || s.Type == ScriptTypeAfter {
		return true
	}
	for _, sub := range s.Scripts {
		if IsTimeLocked(sub) {
			return true
		}
	}
	return false
}

// ErrEmptyScript is returned when a nil or empty script is provided.
var ErrEmptyScript = errors.New("script is empty")
