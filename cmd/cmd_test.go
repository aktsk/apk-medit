package cmd

import (
	"reflect"
	"testing"
)

func TestGetWritableAddrRanges(t *testing.T) {
	actual, _ := GetWritableAddrRanges("testdata/proc_test_maps")
	expected := []string{"75d75f2000-75d75f8000", "7fcf0ff000-7fcf0ff000", "7fcf100000-7fcf8ff000"}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got AddrRanges: %s\nexpected AddrRanges: %s", actual, expected)
	}
}
