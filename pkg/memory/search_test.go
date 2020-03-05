package memory

import (
	"reflect"
	"testing"
)

func TestGetWritableAddrRanges(t *testing.T) {
	actual, _ := GetWritableAddrRanges("testdata/proc_test_maps")
	expected := [][2]int{{506124509184, 506124533760}, {548934774784, 548934774784}, {548934778880, 548943163392}}
	// "75d75f2000-75d75f8000", "7fcf0ff000-7fcf0ff000", "7fcf100000-7fcf8ff000"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got AddrRanges: %v\nexpected AddrRanges: %v", actual, expected)
	}
}

func TestFindDataInSplittedMemory(t *testing.T) {
	memory := []byte{0x10, 0x11, 0x12, 0x10, 0x10, 0x11, 0x12, 0x11, 0x10, 0x11, 0x12, 0x12}
	searchBytes := []byte{0x10, 0x11, 0x12}
	actual := []int{}
	findDataInSplittedMemory(&memory, searchBytes, len(searchBytes), 0x100, 0, &actual)
	expected := []int{0x100, 0x104, 0x108}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got addr slice: %v\nexpected addr slice: %v", actual, expected)
	}
}

func TestFindEmptyInSplittedMemory(t *testing.T) {
	memory := []byte{0x10}
	searchBytes := []byte{0xAA, 0xBB, 0xCC}
	actual := []int{}
	findDataInSplittedMemory(&memory, searchBytes, len(searchBytes), 0x100, 0, &actual)
	expected := []int{}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got addr slice: %v\nexpected addr slice: %v", actual, expected)
	}
}
