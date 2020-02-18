package cmd

import (
	"reflect"
	"testing"
)

func TestStringToBytes(t *testing.T) {
	actual, _ := stringToBytes("147")
	expected := []byte{0x31, 0x34, 0x37}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got AddrRanges: %v\nexpected AddrRanges: %v", actual, expected)
	}
}

func TestWordToBytes(t *testing.T) {
	actual, _ := wordToBytes("19704")
	expected := []byte{0xf8, 0x4c}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got AddrRanges: %v\nexpected AddrRanges: %v", actual, expected)
	}
}

func TestDwordToBytes(t *testing.T) {
	actual, _ := dwordToBytes("19704")
	expected := []byte{0xf8, 0x4c, 0, 0}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got AddrRanges: %v\nexpected AddrRanges: %v", actual, expected)
	}
}

func TestQwordToBytes(t *testing.T) {
	actual, _ := qwordToBytes("19704")
	expected := []byte{0xf8, 0x4c, 0, 0, 0, 0, 0, 0}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got AddrRanges: %v\nexpected AddrRanges: %v", actual, expected)
	}
}
