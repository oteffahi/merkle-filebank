package cryptography

import "crypto/sha256"

func HashTwice(data []byte) [32]byte {
	hash := sha256.Sum256(data)
	return sha256.Sum256(hash[:])
}

func HashOnce(data []byte) [32]byte {
	return sha256.Sum256(data)
}

func CompareHashes(a [32]byte, b [32]byte) bool {
	for i := 0; i < 32; i++ {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}
