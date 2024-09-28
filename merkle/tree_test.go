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

func TestKnownTree(t *testing.T) {
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

func TestTreeStructure(t *testing.T) {
	for i := 1; i <= 10; i++ {
		var files [][]byte
		for j := 0; j < i; j++ {
			files = append(files, []byte(fmt.Sprintf("TEST%d", j)))
		}
		var tree MerkleTree
		if err := tree.BuildMerkleTree(files); err != nil {
			t.Errorf("error when generating tree: %v", err)
			t.FailNow()
		}
		for index, node := range tree.Hashes {
			if node == [32]byte{} {
				// print tree for debug
				hexTree := tree.GetTreeInHex()
				for _, hash := range hexTree {
					fmt.Println(hash)
				}
				t.Errorf("error in tree structure: files=%v, index=%v", i, index)
				t.FailNow()
			}
		}
	}
}
