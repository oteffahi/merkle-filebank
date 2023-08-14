package merkle

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestNoEmptyTree(t *testing.T) {
	var tree MerkleTree

	// testData
	var files [][]byte
	err := tree.BuildMerkeTree(files)

	if err == nil {
		t.Errorf("Expected error, got no error")
	}
}

func TestNominalTree(t *testing.T) {
	var tree MerkleTree

	// testData
	a := []byte("TEST1")
	b := []byte("TEST2")
	c := []byte("TEST3")
	d := []byte("TEST4")
	e := []byte("TEST5")
	files := [][]byte{e, b, c, d, a}
	expectedRoot := "d49293f7d646a5ceb762e3f36ceeef2d0d83918ddae6a610961d2fa231b2a7fa"

	err := tree.BuildMerkeTree(files)

	if err != nil {
		t.Errorf("Error occured when creating tree: %v", err)
	} else {
		gotRoot := tree.GetMerkleRoot()

		// print full tree for debugging
		hexTree := tree.GetTreeInHex()
		for _, hash := range hexTree {
			fmt.Println(hash)
		}

		if expectedRoot != hex.EncodeToString(gotRoot[:]) {
			t.Errorf("Merkle roots do not match. Expected %v, got %v", expectedRoot, gotRoot)
		}
	}
}
