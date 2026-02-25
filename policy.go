package policy

import (
	"context"
	"encoding/hex"
	"fmt"
)

// PolicyID is the hex-encoded Blake2b-224 hash of a serialized native script.
// It uniquely identifies a Cardano minting policy.
type PolicyID string

// String returns the hex string representation of the PolicyID.
func (p PolicyID) String() string { return string(p) }

// Bytes returns the raw 28-byte representation of the PolicyID.
func (p PolicyID) Bytes() ([]byte, error) {
	return hex.DecodeString(string(p))
}

// scriptTag is the CBOR tag for a native/simple script (tag 0).
// Cardano hashes the script with a 1-byte type prefix before the CBOR body.
// For native scripts (type 0): prefix = 0x00.
const nativeScriptHashPrefix = byte(0x00)

// Compute derives the PolicyID for the given Script.
// It serializes the script to its canonical Cardano CBOR encoding,
// prepends the native script type byte (0x00), then hashes with Blake2b-224.
//
// This matches the computation performed by cardano-cli and cardano-addresses.
//
// Example:
//
//	sig, _ := policy.NewSigScript("e09d36c79dec9bd1b3d9e152247701cd0bb0f68e8cd748d2a05")
//	id, err := policy.Compute(sig)
//	fmt.Println(id) // "3a4..."
func Compute(s Script) (PolicyID, error) {
	return ComputeWithContext(context.Background(), s)
}

// ComputeWithContext is like Compute but respects context cancellation.
func ComputeWithContext(ctx context.Context, s Script) (PolicyID, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if err := Validate(s); err != nil {
		return "", err
	}

	cbor, err := encodeScript(s)
	if err != nil {
		return "", fmt.Errorf("policy: failed to encode script: %w", err)
	}

	// Prepend the native script type byte
	payload := append([]byte{nativeScriptHashPrefix}, cbor...)
	hash := blake2b224(payload)
	return PolicyID(hex.EncodeToString(hash)), nil
}

// MustCompute is like Compute but panics on error.
// Useful for package-level policy ID constants in test or CLI code.
func MustCompute(s Script) PolicyID {
	id, err := Compute(s)
	if err != nil {
		panic(err)
	}
	return id
}

// ComputeMany derives PolicyIDs for multiple scripts in a single call.
// Returns a slice of PolicyIDs in the same order as the input scripts.
// If any script fails, the error is returned immediately.
func ComputeMany(scripts ...Script) ([]PolicyID, error) {
	ids := make([]PolicyID, 0, len(scripts))
	for i, s := range scripts {
		id, err := Compute(s)
		if err != nil {
			return nil, fmt.Errorf("script[%d]: %w", i, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// encodeScript produces the canonical CBOR encoding of a Cardano native script.
// This matches the encoding used by cardano-node for policy ID computation.
//
// CBOR encoding rules (from Cardano ledger spec):
//
//	sig:      [0, h'<keyHash>']           (2-element array)
//	all:      [1, [<scripts...>]]         (2-element array)
//	any:      [2, [<scripts...>]]
//	atLeast:  [3, n, [<scripts...>]]      (3-element array)
//	before:   [4, slot]
//	after:    [5, slot]
func encodeScript(s Script) ([]byte, error) {
	enc := newCborEncoder()
	if err := writeScript(enc, s); err != nil {
		return nil, err
	}
	return enc.bytes(), nil
}

func writeScript(enc *cborEncoder, s Script) error {
	switch s.Type {
	case ScriptTypeSig:
		kh, err := hex.DecodeString(s.KeyHash)
		if err != nil {
			return fmt.Errorf("invalid keyHash hex: %w", err)
		}
		enc.writeArray(2)
		enc.writeUint(0)
		enc.writeBytes(kh)

	case ScriptTypeAll:
		enc.writeArray(2)
		enc.writeUint(1)
		enc.writeArray(len(s.Scripts))
		for _, sub := range s.Scripts {
			if err := writeScript(enc, sub); err != nil {
				return err
			}
		}

	case ScriptTypeAny:
		enc.writeArray(2)
		enc.writeUint(2)
		enc.writeArray(len(s.Scripts))
		for _, sub := range s.Scripts {
			if err := writeScript(enc, sub); err != nil {
				return err
			}
		}

	case ScriptTypeAtLeast:
		enc.writeArray(3)
		enc.writeUint(3)
		enc.writeUint(uint64(s.Required))
		enc.writeArray(len(s.Scripts))
		for _, sub := range s.Scripts {
			if err := writeScript(enc, sub); err != nil {
				return err
			}
		}

	case ScriptTypeBefore:
		enc.writeArray(2)
		enc.writeUint(4)
		enc.writeUint(s.Slot)

	case ScriptTypeAfter:
		enc.writeArray(2)
		enc.writeUint(5)
		enc.writeUint(s.Slot)

	default:
		return fmt.Errorf("unknown script type: %s", s.Type)
	}
	return nil
}

// cborEncoder is a minimal, zero-allocation CBOR encoder for the subset of
// CBOR types needed by Cardano native scripts (arrays, unsigned ints, byte strings).
type cborEncoder struct {
	buf []byte
}

func newCborEncoder() *cborEncoder { return &cborEncoder{} }

func (e *cborEncoder) bytes() []byte { return e.buf }

// writeArray writes a CBOR definite-length array header.
func (e *cborEncoder) writeArray(n int) {
	e.writeHead(4, uint64(n)) // major type 4 = array
}

// writeUint writes a CBOR unsigned integer.
func (e *cborEncoder) writeUint(n uint64) {
	e.writeHead(0, n) // major type 0 = unsigned int
}

// writeBytes writes a CBOR byte string.
func (e *cborEncoder) writeBytes(b []byte) {
	e.writeHead(2, uint64(len(b))) // major type 2 = byte string
	e.buf = append(e.buf, b...)
}

// writeHead encodes a CBOR initial byte + length.
func (e *cborEncoder) writeHead(major byte, val uint64) {
	major <<= 5
	switch {
	case val <= 23:
		e.buf = append(e.buf, major|byte(val))
	case val <= 0xff:
		e.buf = append(e.buf, major|24, byte(val))
	case val <= 0xffff:
		e.buf = append(e.buf, major|25, byte(val>>8), byte(val))
	case val <= 0xffffffff:
		e.buf = append(e.buf, major|26, byte(val>>24), byte(val>>16), byte(val>>8), byte(val))
	default:
		e.buf = append(e.buf, major|27,
			byte(val>>56), byte(val>>48), byte(val>>40), byte(val>>32),
			byte(val>>24), byte(val>>16), byte(val>>8), byte(val))
	}
}
