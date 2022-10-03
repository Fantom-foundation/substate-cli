package tracer

import (
	"os"
	"testing"
)

// Add()
// Positive Test: Add a new set of file positions and compare the size of position index map
func TestPositiveFilePositionIndexAdd(t *testing.T) {
	var blk1 uint64 = 1
	var blk2 uint64 = 2
	var pos1 [NumWriteOperations]uint64
	var pos2 [NumWriteOperations]uint64
	fpi := NewFilePositionIndex()

	for i := 0; i < NumWriteOperations; i++ {
		pos1[i] = uint64(i)
		pos2[i] = uint64(i) + NumWriteOperations
	}
	err1 := fpi.Add(blk1, pos1)
	if err1 != nil {
		t.Fatalf("Failed to add new block: %v", err1)
	}
	err2 := fpi.Add(blk2, pos2)
	if err2 != nil {
		t.Fatalf("Failed to add new block: %v", err2)
	}
	want := 2
	have := len(fpi.blockToFilePos)
	if have != want {
		t.Fatalf("Unexpected map size")
	}
}

// Negative Test: Add a duplicate set of indices and compare whether the values are added twice
func TestNegativeFilePositionIndexAdd(t *testing.T) {
	var blk uint64 = 1
	var pos [NumWriteOperations]uint64
	fpi := NewFilePositionIndex()

	for i := 0; i < NumWriteOperations; i++ {
		pos[i] = uint64(i)
	}
	err1 := fpi.Add(blk, pos)
	if err1 != nil {
		t.Fatalf("Failed to add new block: %v", err1)
	}
	err2 := fpi.Add(blk, pos)
	if err2 == nil {
		t.Fatalf("Failed to report error when adding an existing block")
	}

	want := 1
	have := len(fpi.blockToFilePos)
	if have != want {
		t.Fatalf("Unexpectd map size")
	}
}

// Get()
// Positive Test: Get file positions from FilePositionIndex and compare index postions
func TestPositiveFilePositionIndexGet(t *testing.T) {
	var blk uint64 = 1
	var pos [NumWriteOperations]uint64
	fpi := NewFilePositionIndex()
	for i := 0; i < NumWriteOperations; i++ {
		pos[i] = uint64(i)
	}

	fpi.Add(blk, pos)
	filepos, err := fpi.Get(blk)
	if err != nil || len(pos) != len(filepos) {
		t.Fatalf("Failed to get block %v", blk)
	}

	for i := 0; i < NumWriteOperations; i++ {
		if pos[i] != filepos[i] {
			t.Fatalf("Index mismatched")
		}
	}
}

// Negative Test: Get file positions of a block whcih is not in FilePositionIndex
func TestNegativeFilePositionIndexGet(t *testing.T) {
	var blk uint64 = 1
	var pos [NumWriteOperations]uint64
	fpi := NewFilePositionIndex()
	for i := 0; i < NumWriteOperations; i++ {
		pos[i] = uint64(i)
	}

	fpi.Add(blk, pos)
	_, err := fpi.Get(blk + 1)
	if err == nil {
		t.Fatalf("Failed to report error. Block %v doesn't exist", blk+1)
	}
}

// Read and Write()
// Positive Tetst: Write a set of postion index to a binary file and read from it.
// Compare whether indices are the same.
func TestPositiveFilePositionIndexReadWrite(t *testing.T) {
	var blk1 uint64 = 1
	var blk2 uint64 = 2
	var pos1 [NumWriteOperations]uint64
	var pos2 [NumWriteOperations]uint64
	filename := "./test.dat"

	wFpi := NewFilePositionIndex()
	for i := 0; i < NumWriteOperations; i++ {
		pos1[i] = uint64(i)
		pos2[i] = uint64(i) + NumWriteOperations
	}
	wFpi.Add(blk1, pos1)
	wFpi.Add(blk2, pos2)

	err1 := wFpi.Write(filename)
	defer os.Remove(filename)
	if err1 != nil {
		t.Fatalf("Failed to write file. %v", err1)
	}
	rFpi := NewFilePositionIndex()
	err2 := rFpi.Read(filename)
	if err2 != nil {
		t.Fatalf("Failed to read file. %v", err2)
	}
	filepos1, err3 := rFpi.Get(blk1)
	if err3 != nil || len(pos1) != len(filepos1) {
		t.Fatalf("Failed to get block %v with error: %v", blk1, err3)
	}
	for i := 0; i < NumWriteOperations; i++ {
		if pos1[i] != filepos1[i] {
			t.Fatalf("Index mismatched")
		}
	}
	filepos2, err4 := rFpi.Get(blk2)
	if err4 != nil || len(pos2) != len(filepos2) {
		t.Fatalf("Failed to get block %v with error: %v", blk2, err4)
	}
	for i := 0; i < NumWriteOperations; i++ {
		if pos2[i] != filepos2[i] {
			t.Fatalf("Index mismatched")
		}
	}

}

// Positive Tetst: Create
// Negative Tetst: Write a corrupted file and read from it.
func TestNegativeFilePositionIndexWrite(t *testing.T) {
	filename := "./test.dict"
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatalf("Failed to open file")
	}
	defer os.Remove(filename)
	// write corrupted entry
	data := []byte("hello")
	if _, err := f.Write(data); err != nil {
		t.Fatalf("Failed to write data")
	}
	err = f.Close()
	if err != nil {
		t.Fatalf("Failed to close file")
	}
	fpi := NewFilePositionIndex()
	err = fpi.Read(filename)
	if err == nil {
		t.Fatalf("Failed to report error")
	}
}
