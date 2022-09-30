package replay

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"os"
	"testing"
)

// Positive Test: Check whether GetFilename for index zero and one returns the first two
// operation filenames.
func TestPositiveGetFilename(t *testing.T) {
	fn := GetFilename(0)
	fmt.Printf("f: %v\n", fn)
	if fn != "sop-getstate.dat" {
		t.Fatalf("GetFilename(0) failed; returns %v", fn)
	}
	fn = GetFilename(1)
	if fn != "sop-setstate.dat" {
		t.Fatalf("GetFilename(1) failed; returns %v", fn)
	}
}

// Positive Test: Write/read test for GetStateOperation
func TestPositiveWriteReadGetState(t *testing.T) {
	filename := "./test.dat"

	//  write two test object to file
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatalf("Failed to open file for writing")
	}
	// write first object
	var sop = &GetStateOperation{ContractIndex: 1, StorageIndex: 2}
	sop.GetWriteable().Set(1001)
	sop.Write(f)
	// write second object
	sop.ContractIndex = 100
	sop.StorageIndex = 200
	sop.GetWriteable().Set(1010)
	sop.Write(f)
	err = f.Close()
	if err != nil {
		t.Fatalf("Failed to close file for writing")
	}

	// read test object from file
	f, err = os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open file for reading")
	}
	// read first object & compare
	data, err := ReadGetStateOperation(f)
	if err != nil {
		t.Fatalf("Failed to read from file")
	}
	if data.GetWriteable().Get() != 1001 || data.ContractIndex != 1 || data.StorageIndex != 2 {
		t.Fatalf("Failed comparison")
	}
	// read second object & compare
	data, err = ReadGetStateOperation(f)
	if err != nil {
		t.Fatalf("Failed to read from file")
	}
	if data.GetWriteable().Get() != 1010 || data.ContractIndex != 100 || data.StorageIndex != 200 {
		t.Fatalf("Failed comparison")
	}
	err = f.Close()
	if err != nil {
		t.Fatalf("Failed to close file for reading")
	}

	// read test object from file
	os.Remove(filename)
}

// Positive Test: Write/read test for SetStateOperation
func TestPositiveWriteReadSetState(t *testing.T) {
	filename := "./test.dat"

	//  write two test object to file
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatalf("Failed to open file for writing")
	}
	// write first object
	var sop = &SetStateOperation{ContractIndex: 1, StorageIndex: 2, Value: common.HexToHash("0x1000312211312312321312")}
	sop.GetWriteable().Set(1001)
	sop.Write(f)
	// write second object
	sop.ContractIndex = 100
	sop.StorageIndex = 200
	sop.Value = common.HexToHash("0x123111231231283012083")
	sop.GetWriteable().Set(1010)
	sop.Write(f)
	err = f.Close()
	if err != nil {
		t.Fatalf("Failed to close file for writing")
	}

	// read test object from file
	f, err = os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open file for reading")
	}
	// read first object & compare
	data, err := ReadSetStateOperation(f)
	if err != nil {
		t.Fatalf("Failed to read from file")
	}
	if data.GetWriteable().Get() != 1001 || data.ContractIndex != 1 || data.StorageIndex != 2 {
		t.Fatalf("Failed comparison")
	}
	// read second object & compare
	data, err = ReadSetStateOperation(f)
	if err != nil {
		t.Fatalf("Failed to read from file")
	}
	if data.GetWriteable().Get() != 1010 || data.ContractIndex != 100 || data.StorageIndex != 200 {
		t.Fatalf("Failed comparison")
	}
	err = f.Close()
	if err != nil {
		t.Fatalf("Failed to close file for reading")
	}

	// read test object from file
	os.Remove(filename)
}
