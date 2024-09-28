package merkle

import (
	"encoding/hex"
	"errors"

	cr "github.com/oteffahi/merkle-filebank/cryptography"
)

type MerkleProof struct {
	Leaf   [32]byte
	Hashes [][32]byte
}

func (m MerkleTree) generateProof(leaf [32]byte) (*MerkleProof, error) {
	if len(m.Hashes) == 0 {
		return nil, errors.New("cannot generate proof from empty tree")
	}
	// get leaf position in tree
	leafIndex, err := m.getNodeIndex(leaf)
	if err != nil {
		return nil, err
	}
	if leafIndex == 0 {
		return &MerkleProof{
			Leaf:   leaf,
			Hashes: [][32]byte{},
		}, nil
	}
	if leafIndex == -1 {
		return nil, errors.New("leaf is not part of the tree")
	}
	// TODO: check that found index is a leaf
	var proof [][32]byte
	currentIndex := leafIndex
	for currentIndex != 0 {
		proof = append(proof, m.Hashes[getNodeSiblingIndex(currentIndex)])
		currentIndex = getNodeParentIndex(currentIndex)
	}
	return &MerkleProof{
		Leaf:   leaf,
		Hashes: proof,
	}, nil
}

func (p MerkleProof) GetProofInHex() []string {
	var hexProof []string
	for _, hash := range p.Hashes {
		hexProof = append(hexProof, hex.EncodeToString(hash[:]))
	}
	return hexProof
}

func (p MerkleProof) VerifyFileProof(file []byte, merkleRoot [32]byte) (bool, error) {
	leaf := cr.HashTwice(file)
	return p.verifyLeafProof(leaf, merkleRoot), nil
}

func (p MerkleProof) verifyLeafProof(leaf [32]byte, merkleRoot [32]byte) bool {
	buff := leaf
	for _, hash := range p.Hashes {
		buff = concatAndHash(buff, hash)
	}
	return buff == merkleRoot
}

func getNodeParentIndex(index int) int {
	return (index - 1) / 2
}

func getNodeSiblingIndex(index int) int {
	if index%2 == 0 {
		return index - 1
	}
	return index + 1
}
