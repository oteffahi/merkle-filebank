package merkle

import (
	"encoding/hex"
	"errors"
	"sort"

	cr "github.com/oteffahi/merkle-filebank/cryptography"
)

type MerkleTree struct {
	hashes [][32]byte
}

func (m MerkleTree) GetMerkleRoot() [32]byte {
	if len(m.hashes) > 0 {
		return m.hashes[0]
	}
	return [32]byte{}
}

func (m MerkleTree) GetTreeInHex() []string {
	var hexTree []string

	for _, hash := range m.hashes {
		hexTree = append(hexTree, hex.EncodeToString(hash[:]))
	}

	return hexTree
}

func (m *MerkleTree) BuildMerkeTree(files [][]byte) error {
	if len(files) == 0 {
		return errors.New("Cannot create tree from empty slice")
	}
	var leafs [][32]byte

	for _, file := range files {
		leafs = append(leafs, cr.HashTwice(file))
	}

	tree := merkleTreeFromLeafs(leafs)

	m.hashes = tree

	return nil
}

func merkleTreeFromLeafs(leafs [][32]byte) [][32]byte {
	// sort leafs
	sort.Slice(leafs, func(i, j int) bool {
		return cr.CompareHashes(leafs[i], leafs[j])
	})

	var tree [][32]byte

	level := leafs

	for len(level) > 1 {
		var newLevel [][32]byte

		for i := 0; i < len(level)-1; i += 2 {
			newNode := concatAndHash(level[i], level[i+1])
			newLevel = append(newLevel, newNode)
		}

		if len(level)%2 != 0 { // uneven number of nodes : final node must be moved to newLevel
			newLevel = append(newLevel, level[len(level)-1])
			level = level[:len(level)-1]
		}
		// commit current level to tree
		tree = append(level, tree...)
		// next iteration on newLevel
		level = newLevel
	}
	// append merkle root to tree
	tree = append(level, tree...)

	return tree
}

func concatAndHash(hash1 [32]byte, hash2 [32]byte) [32]byte {
	var buffer []byte
	if cr.CompareHashes(hash1, hash2) {
		buffer = append(buffer, hash1[:]...)
		buffer = append(buffer, hash2[:]...)
	} else {
		buffer = append(buffer, hash2[:]...)
		buffer = append(buffer, hash1[:]...)
	}
	return cr.HashOnce(buffer)
}
