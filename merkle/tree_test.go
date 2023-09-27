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
	err := tree.BuildMerkleTree(files)

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
	expectedRoot := "49e5171f64c94c819582d1b433156a604b916ef5774765be78c6dc646585a7fa"

	err := tree.BuildMerkleTree(files)

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
