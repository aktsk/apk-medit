package test

import (
	"reflect"
	"testing"

	"github.com/aktsk/medit/cmd"
)

func TestGetWritableAddrRanges(t *testing.T) {
	actual, _ := cmd.GetWritableAddrRanges("./proc_test_maps")
	expected := []string{"75d75f2000-75d75f8000", "7fcf0ff000-7fcf0ff000", "7fcf100000-7fcf8ff000"}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got AddrRanges: %s\nexpected AddrRanges: %s", actual, expected)
	}
}
