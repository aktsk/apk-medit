package cmd

import (
	"reflect"
	"testing"
)

func TestStringToBytes(t *testing.T) {
	actual, _ := stringToBytes("147")
	expected := []byte{0x31, 0x34, 0x37}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got Bytes: %v\nexpected Bytes: %v", actual, expected)
	}
}

func TestWordToBytes(t *testing.T) {
	actual, _ := wordToBytes("19704")
	expected := []byte{0xf8, 0x4c}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got Bytes: %v\nexpected Bytes: %v", actual, expected)
	}
}

func TestFailWordToBytesByOutOfRange(t *testing.T) {
	_, err := wordToBytes("1193046") // 0x123456
	if err == nil {
		t.Errorf("Expected strconv.ParseUint: value out of range\n")
	}
}

func TestDwordToBytes(t *testing.T) {
	actual, _ := dwordToBytes("19704")
	expected := []byte{0xf8, 0x4c, 0, 0}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got Bytes: %v\nexpected Bytes: %v", actual, expected)
	}
}

func TestFailDwordToBytesByOutOfRange(t *testing.T) {
	_, err := dwordToBytes("78187493530") // 0x123456789A
	if err == nil {
		t.Errorf("Expected strconv.ParseUint: value out of range\n")
	}
}

func TestQwordToBytes(t *testing.T) {
	actual, _ := qwordToBytes("19704")
	expected := []byte{0xf8, 0x4c, 0, 0, 0, 0, 0, 0}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got Bytes: %v\nexpected Bytes: %v", actual, expected)
	}
}
