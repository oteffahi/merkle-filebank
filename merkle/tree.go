package merkle

import (
	"encoding/hex"
	"errors"
	"sort"
	"sync/atomic"

	cr "github.com/oteffahi/merkle-filebank/cryptography"
)

type MerkleTree struct {
	Hashes [][32]byte
}

func (m *MerkleTree) BuildMerkleTree(files [][]byte) error {
	if len(files) == 0 {
		return errors.New("cannot create tree from empty slice")
	}
	var leafs [][32]byte
	for _, file := range files {
		leafs = append(leafs, cr.HashTwice(file))
	}
	tree := merkleTreeFromLeafs(leafs)
	m.Hashes = tree
	return nil
}

func (m MerkleTree) GetMerkleRoot() [32]byte {
	if len(m.Hashes) > 0 {
		return m.Hashes[0]
	}
	return [32]byte{}
}

func (m MerkleTree) GenerateProofForFile(file []byte) (*MerkleProof, error) {
	leaf := cr.HashTwice(file)
	proof, err := m.generateProof(leaf)
	if err != nil {
		return nil, err
	}
	return proof, nil
}

func (m MerkleTree) GetTreeInHex() []string {
	var hexTree []string
	for _, hash := range m.Hashes {
		hexTree = append(hexTree, hex.EncodeToString(hash[:]))
	}
	return hexTree
}

func (m MerkleTree) getNodeIndex(leaf [32]byte) (int, error) {
	size := len(m.Hashes)
	if size == 0 {
		return -1, errors.New("cannot search in empty tree")
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

func merkleTreeFromLeafs(leafs [][32]byte) [][32]byte {
	if len(leafs) == 1 {
		return leafs[:1]
	}
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
	atomicBuffers := make([]atomic.Pointer[[32]byte], treeLen-len(leafs))
	notifyEnd := make(chan struct{})
	defer close(notifyEnd)
	for i := treeLen - 1; i > treeLen-len(leafs); i -= 2 {
		go merkleBuildWorker(tree, atomicBuffers, i, notifyEnd)
	}
	<-notifyEnd
	return tree
}

func merkleBuildWorker(tree [][32]byte, atomicBuffers []atomic.Pointer[[32]byte], index1 int, ch chan<- struct{}) {
	var (
		index2      int
		resultIndex int
		elm1        *[32]byte
		elm2        *[32]byte
	)
	// compute other index
	if index1%2 == 0 {
		index2 = index1 - 1
	} else {
		index2 = index1 + 1
	}
	// fetch nodes from leafs/atomicBuffer
	mapIndexes := func(a, b int, f func(int) *[32]byte) (elm1 *[32]byte, elm2 *[32]byte) {
		return f(a), f(b)
	}
	elm1, elm2 = mapIndexes(index1, index2, func(i int) *[32]byte {
		if i >= len(atomicBuffers) {
			return &tree[i]
		} else {
			return atomicBuffers[i].Load()
		}
	})
	if elm1 == nil || elm2 == nil {
		// This goroutine has computed one of the two buffers in the previous call
		// if the other buffer is empty, the goroutine computing that buffer will continue
		return
	}
	// compute next node
	resultIndex = min(index1, index2) / 2
	tree[resultIndex] = concatAndHash(*elm1, *elm2)
	atomicBuffers[resultIndex].Store(&tree[resultIndex])
	// start computation of next level, or terminate if root reached
	if resultIndex == 0 {
		ch <- struct{}{}
	} else {
		merkleBuildWorker(tree, atomicBuffers, resultIndex, ch)
	}
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
