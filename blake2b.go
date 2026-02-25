package policy

// blake2b224 computes the Blake2b-224 (28-byte) digest of the input.
// This is a pure-Go, zero-dependency implementation of the Blake2b spec
// (https://www.blake2.net/blake2.pdf), restricted to the 224-bit output
// variant used by Cardano for policy ID and address derivation.
//
// This implementation is intentionally simple, clear, and auditable.
// It is NOT constant-time and is NOT suitable for keyed MAC use.
func blake2b224(data []byte) []byte {
	return blake2bSum(data, 28)
}

// --- Blake2b core (RFC 7693 / BLAKE2 spec) ---

// Initialization vectors (same as SHA-512 IVs).
var iv = [8]uint64{
	0x6a09e667f3bcc908, 0xbb67ae8584caa73b,
	0x3c6ef372fe94f82b, 0xa54ff53a5f1d36f1,
	0x510e527fade682d1, 0x9b05688c2b3e6c1f,
	0x1f83d9abfb41bd6b, 0x5be0cd19137e2179,
}

// Sigma permutation table (10 rounds, 16 indices each).
var sigma = [10][16]byte{
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	{14, 10, 4, 8, 9, 15, 13, 6, 1, 12, 0, 2, 11, 7, 5, 3},
	{11, 8, 12, 0, 5, 2, 15, 13, 10, 14, 3, 6, 7, 1, 9, 4},
	{7, 9, 3, 1, 13, 12, 11, 14, 2, 6, 5, 10, 4, 0, 15, 8},
	{9, 0, 5, 7, 2, 4, 10, 15, 14, 1, 11, 12, 6, 8, 3, 13},
	{2, 12, 6, 10, 0, 11, 8, 3, 4, 13, 7, 5, 15, 14, 1, 9},
	{12, 5, 1, 15, 14, 13, 4, 10, 0, 7, 6, 3, 9, 2, 8, 11},
	{13, 11, 7, 14, 12, 1, 3, 9, 5, 0, 15, 4, 8, 6, 2, 10},
	{6, 15, 14, 9, 11, 3, 0, 8, 12, 2, 13, 7, 1, 4, 10, 5},
	{10, 2, 8, 4, 7, 6, 1, 5, 15, 11, 9, 14, 3, 12, 13, 0},
}

func rotr64(x uint64, n uint) uint64 {
	return (x >> n) | (x << (64 - n))
}

func g(v *[16]uint64, a, b, c, d int, x, y uint64) {
	v[a] = v[a] + v[b] + x
	v[d] = rotr64(v[d]^v[a], 32)
	v[c] = v[c] + v[d]
	v[b] = rotr64(v[b]^v[c], 24)
	v[a] = v[a] + v[b] + y
	v[d] = rotr64(v[d]^v[a], 16)
	v[c] = v[c] + v[d]
	v[b] = rotr64(v[b]^v[c], 63)
}

func compress(h *[8]uint64, block []byte, counter uint64, last bool) {
	var m [16]uint64
	for i := 0; i < 16; i++ {
		off := i * 8
		if off+8 <= len(block) {
			m[i] = uint64(block[off]) |
				uint64(block[off+1])<<8 |
				uint64(block[off+2])<<16 |
				uint64(block[off+3])<<24 |
				uint64(block[off+4])<<32 |
				uint64(block[off+5])<<40 |
				uint64(block[off+6])<<48 |
				uint64(block[off+7])<<56
		}
	}

	var v [16]uint64
	copy(v[:8], h[:])
	v[8] = iv[0]
	v[9] = iv[1]
	v[10] = iv[2]
	v[11] = iv[3]
	v[12] = iv[4] ^ counter
	v[13] = iv[5] // upper 64 bits of counter (always 0 for our use)
	v[14] = iv[6]
	v[15] = iv[7]
	if last {
		v[14] ^= 0xffffffffffffffff
	}

	for r := 0; r < 12; r++ {
		s := sigma[r%10]
		g(&v, 0, 4, 8, 12, m[s[0]], m[s[1]])
		g(&v, 1, 5, 9, 13, m[s[2]], m[s[3]])
		g(&v, 2, 6, 10, 14, m[s[4]], m[s[5]])
		g(&v, 3, 7, 11, 15, m[s[6]], m[s[7]])
		g(&v, 0, 5, 10, 15, m[s[8]], m[s[9]])
		g(&v, 1, 6, 11, 12, m[s[10]], m[s[11]])
		g(&v, 2, 7, 8, 13, m[s[12]], m[s[13]])
		g(&v, 3, 4, 9, 14, m[s[14]], m[s[15]])
	}
	for i := 0; i < 8; i++ {
		h[i] ^= v[i] ^ v[i+8]
	}
}

// blake2bSum computes a Blake2b digest of digestLen bytes (max 64).
func blake2bSum(data []byte, digestLen int) []byte {
	var h [8]uint64
	copy(h[:], iv[:])

	// Parameter block: fan-out=1, max depth=1, digest length = digestLen
	h[0] ^= 0x01010000 ^ uint64(digestLen)

	blockSize := 128
	buf := make([]byte, blockSize)
	bytesCompressed := uint64(0)
	remaining := data

	for len(remaining) > blockSize {
		copy(buf, remaining[:blockSize])
		remaining = remaining[blockSize:]
		bytesCompressed += uint64(blockSize)
		compress(&h, buf, bytesCompressed, false)
	}

	// Final block (zero-padded)
	var finalBlock [128]byte
	copy(finalBlock[:], remaining)
	bytesCompressed += uint64(len(remaining))
	compress(&h, finalBlock[:], bytesCompressed, true)

	// Encode output (little-endian)
	out := make([]byte, digestLen)
	for i := 0; i < digestLen; i++ {
		out[i] = byte(h[i/8] >> (uint(i%8) * 8))
	}
	return out
}
