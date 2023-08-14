package merkle

import (
	"fmt"
	"testing"
)

func TestNominalProof(t *testing.T) {
	var tree MerkleTree

	// testData
	a := []byte("TEST1")
	b := []byte("TEST2")
	c := []byte("TEST3")
	d := []byte("TEST4")
	e := []byte("TEST5")
	files := [][]byte{e, b, c, d, a}

	err := tree.BuildMerkeTree(files)

	proof, err := tree.GenerateProofForFile(a)
	if err != nil {
		t.Errorf("Error occured when generating proof: %v", err)
		return
	}

	gotProof := proof.GetProofInHex()
	// debugging
	fmt.Println(gotProof)

	expectedProof := []string{
		"0ff22d9aeb032108d73111e206db99a18758707742ebb6ef3ce7a11ce238b3ee",
		"7896435a0622d2f07caafa76d28c23e2bce225aefceeef013495a828f4492436",
		"d22151099bfd35343be610eb424dd95024626ad051a4f2bc1dc76d47adbca40b",
	}

	if len(gotProof) != len(expectedProof) {
		t.Errorf("Generated proof is not equal to expected")
		return
	}
	for i := range expectedProof {
		if expectedProof[i] != gotProof[i] {
			t.Errorf("Generated proof is not equal to expected")
			return
		}
	}
}
