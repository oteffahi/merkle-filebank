package merkle

import (
	"fmt"
	"testing"
)

func TestNominalProof(t *testing.T) {
	// testData
	var files [][]byte
	for i := 0; i < 100; i++ {
		files = append(files, []byte(fmt.Sprintf("TEST%d", i)))
	}

	var tree MerkleTree
	if err := tree.BuildMerkleTree(files); err != nil {
		t.Errorf("error occured when generating tree: %v", err)
		t.FailNow()
	}
	for i, file := range files {
		proof, err := tree.GenerateProofForFile(file)
		if err != nil {
			t.Errorf("error occured when generating proof: %v", err)
			t.FailNow()
		}
		isValidProof, err := proof.VerifyFileProof(file, tree.GetMerkleRoot())
		if err != nil {
			t.Errorf("error occured when verifying proof: %v", err)
		} else if !isValidProof {
			t.Errorf("failed to verify proof for file %v", i)
		}
	}
}

func TestNoProofFromEmptyTree(t *testing.T) {
	var tree MerkleTree
	if _, err := tree.generateProof([32]byte{}); err == nil {
		t.Errorf("generateProof should return error when tree is empty")
	}
}

func TestFailVerification(t *testing.T) {
	// testData
	var files [][]byte
	for i := 0; i < 100; i++ {
		files = append(files, []byte(fmt.Sprintf("TEST%d", i)))
	}

	var tree MerkleTree
	if err := tree.BuildMerkleTree(files); err != nil {
		t.Errorf("error when generating tree: %v", err)
		return
	}
	proof, err := tree.GenerateProofForFile(files[0])
	if err != nil {
		t.Errorf("error when generating proof: %v", err)
		t.FailNow()
	}
	isValidProof, err := proof.VerifyFileProof(files[5], tree.GetMerkleRoot())
	if err != nil {
		t.Errorf("error when verifying proof: %v", err)
	} else if isValidProof {
		t.Errorf("expected proof verification to fail, got success")
	}
}
