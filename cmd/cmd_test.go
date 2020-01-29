package cmd

import (
	"reflect"
	"testing"
)

func TestGetWritableAddrRanges(t *testing.T) {
	actual, _ := getWritableAddrRanges("testdata/proc_test_maps")
	expected := [][2]int64{{506124509184, 506124533760}, {548934774784, 548934774784}, {548934778880, 548943163392}}
	// "75d75f2000-75d75f8000", "7fcf0ff000-7fcf0ff000", "7fcf100000-7fcf8ff000"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got AddrRanges: %v\nexpected AddrRanges: %v", actual, expected)
	}
}
