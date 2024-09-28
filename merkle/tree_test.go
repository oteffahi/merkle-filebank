package merkle

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestNoEmptyTree(t *testing.T) {
	// testData
	var files [][]byte

	var tree MerkleTree
	if err := tree.BuildMerkleTree(files); err == nil {
		t.Errorf("merkle tree builder should return error when tree is empty")
	}
}

func TestNominalTree(t *testing.T) {
	// testData
	a := []byte("TEST1")
	b := []byte("TEST2")
	c := []byte("TEST3")
	d := []byte("TEST4")
	e := []byte("TEST5")
	files := [][]byte{e, b, c, d, a}
	expectedRoot := "49e5171f64c94c819582d1b433156a604b916ef5774765be78c6dc646585a7fa"

	var tree MerkleTree
	if err := tree.BuildMerkleTree(files); err != nil {
		t.Errorf("error occured when building tree: %v", err)
		t.FailNow()
	}
	root := tree.GetMerkleRoot()
	gotRoot := hex.EncodeToString(root[:])
	if expectedRoot != gotRoot {
		// print full tree for debugging
		hexTree := tree.GetTreeInHex()
		for _, hash := range hexTree {
			fmt.Println(hash)
		}
		t.Errorf("merkle roots do not match. Expected %v, got %v", expectedRoot, gotRoot)
	}
}
