package memory

import (
	"os"
	"reflect"
	"testing"
)

func TestReadMemory(t *testing.T) {
	memFile, _ := os.Open("testdata/proc_test_mem")
	defer memFile.Close()
	saved := make([]byte, 5)
	actual := ReadMemory(memFile, saved, 0x3, 0x8) // Memory address is zero origin.
	expected := []byte{0x3, 0x4, 0x5, 0x6, 0x7}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got memory bytes: %v\nexpected memory bytes: %v", actual, expected)
	}
}
