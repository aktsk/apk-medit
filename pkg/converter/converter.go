package converter

import (
	"encoding/binary"
	"strconv"
)

// UTF8 string
func StringToBytes(arg string) ([]byte, error) {
	rs := []rune(arg)
	return []byte(string(rs)), nil
}

func WordToBytes(arg string) ([]byte, error) {
	searchBytes := make([]byte, 2)
	targetVal, err := strconv.ParseUint(arg, 10, 16)
	if err != nil {
		return nil, err
	}
	binary.LittleEndian.PutUint16(searchBytes[0:], uint16(targetVal))
	for i := len(searchBytes); i < 2; i++ {
		searchBytes = append(searchBytes, 0)
	}
	return searchBytes, nil
}

func DwordToBytes(arg string) ([]byte, error) {
	searchBytes := make([]byte, 4)
	targetVal, err := strconv.ParseUint(arg, 10, 32)
	if err != nil {
		return nil, err
	}
	binary.LittleEndian.PutUint32(searchBytes[0:], uint32(targetVal))
	return searchBytes, nil
}

func QwordToBytes(arg string) ([]byte, error) {
	searchBytes := make([]byte, 8)
	targetVal, err := strconv.ParseUint(arg, 10, 64)
	if err != nil {
		return nil, err
	}
	binary.LittleEndian.PutUint64(searchBytes[0:], uint64(targetVal))
	return searchBytes, nil
}
