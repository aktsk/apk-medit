package cmd

import (
	"io"
	"os"
)

func readMemory(memFile *os.File, buffer []byte, beginAddr int, endAddr int) []byte {
	n := endAddr - beginAddr
	r := io.NewSectionReader(memFile, int64(beginAddr), int64(n))
	r.Read(buffer)
	return buffer
}

func writeMemory(memFile *os.File, targetAddr int, targetVal []byte) error {
	if _, err := memFile.WriteAt(targetVal, int64(targetAddr)); err != nil {
		return err
	}
	return nil
}
