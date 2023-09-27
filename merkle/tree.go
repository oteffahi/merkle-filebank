package merkle

import (
	"encoding/hex"
	"errors"
	"sort"

	cr "github.com/oteffahi/merkle-filebank/cryptography"
)

type MerkleTree struct {
	Hashes [][32]byte
}

func (m MerkleTree) GetMerkleRoot() [32]byte {
	if len(m.Hashes) > 0 {
		return m.Hashes[0]
	}
	return [32]byte{}
}

func (m MerkleTree) GetTreeInHex() []string {
	var hexTree []string

	for _, hash := range m.Hashes {
		hexTree = append(hexTree, hex.EncodeToString(hash[:]))
	}

	return hexTree
}

func (m *MerkleTree) BuildMerkleTree(files [][]byte) error {
	if len(files) == 0 {
		return errors.New("Cannot create tree from empty slice")
	}
	var leafs [][32]byte

	for _, file := range files {
		leafs = append(leafs, cr.HashTwice(file))
	}

	tree := merkleTreeFromLeafs(leafs)

	m.Hashes = tree

	return nil
}

func (m MerkleTree) getNodeIndex(leaf [32]byte) (int, error) {
	size := len(m.Hashes)
	if size == 0 {
		return -1, errors.New("Cannot search in empty tree")
	}
	nbLeafs := (size + 1) / 2
	low := size - nbLeafs
	high := size - 1
	// binary search in leafs
	for low <= high {
		median := (low + high) / 2
		if m.Hashes[median] == leaf {
			return median, nil
		} else if !cr.CompareHashes(m.Hashes[median], leaf) {
			low = median + 1
		} else {
			high = median - 1
		}
	}
	if low == size || m.Hashes[low] != leaf {
		return -1, nil
	}
	return low, nil
}

func (m MerkleTree) GenerateProofForFile(file []byte) (*MerkleProof, error) {
	leaf := cr.HashTwice(file)

	proof, err := m.generateProof(leaf)
	if err != nil {
		return nil, err
	}

	return proof, nil
}

func merkleTreeFromLeafs(leafs [][32]byte) [][32]byte {
	// sort leafs
	sort.Slice(leafs, func(i, j int) bool {
		return cr.CompareHashes(leafs[i], leafs[j])
	})

	treeLen := len(leafs)*2 - 1
	tree := make([][32]byte, treeLen)

	// insert leafs at end of buffer in reverse order
	for i, leaf := range leafs {
		tree[treeLen-1-i] = leaf
	}

	// compute nodes
	for i := treeLen - len(leafs) - 1; i >= 0; i-- {
		tree[i] = concatAndHash(tree[2*i+1], tree[2*i+2])
	}

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
