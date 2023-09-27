package merkle

import (
	"fmt"
	"testing"
)

func TestNominalProof(t *testing.T) {
	var tree MerkleTree

	// testData
	var files [][]byte
	for i := 0; i < 100; i++ {
		files = append(files, []byte(fmt.Sprintf("TEST%d", i)))
	}

	err := tree.BuildMerkleTree(files)

	if err != nil {
		t.Errorf("Error occured when generating tree: %v", err)
		return
	}

	for i, file := range files {
		proof, err := tree.GenerateProofForFile(file)
		if err != nil {
			t.Errorf("Error occured when generating proof: %v", err)
			return
		}
		isValidProof, err := proof.VerifyFileProof(file, tree.GetMerkleRoot())
		if err != nil {
			t.Errorf("Error occured when verifying proof: %v", err)
			return
		}
		if !isValidProof {
			t.Errorf("Failed to verify proof for file %v", i)
			return
		}
	}
}

func TestFailVerification(t *testing.T) {
	var tree MerkleTree

	// testData
	var files [][]byte
	for i := 0; i < 100; i++ {
		files = append(files, []byte(fmt.Sprintf("TEST%d", i)))
	}

	err := tree.BuildMerkleTree(files)

	if err != nil {
		t.Errorf("Error occured when generating tree: %v", err)
		return
	}

	proof, err := tree.GenerateProofForFile(files[0])
	if err != nil {
		t.Errorf("Error occured when generating proof: %v", err)
		return
	}
	isValidProof, err := proof.VerifyFileProof(files[5], tree.GetMerkleRoot())
	if err != nil {
		t.Errorf("Error occured when verifying proof: %v", err)
		return
	}
	if isValidProof {
		t.Errorf("Expected proof verification to fail, got success")
		return
	}
}
