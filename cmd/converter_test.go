package cmd

import (
	"reflect"
	"testing"
)

func TestIntToUTF8bytes(t *testing.T) {
	actual := intToUTF8bytes("147")
	expected := []byte{0x31, 0x34, 0x37}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got AddrRanges: %v\nexpected AddrRanges: %v", actual, expected)
	}
}
