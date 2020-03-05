package memory

import (
	"io"
	"os"
)

func ReadMemory(memFile *os.File, buffer []byte, beginAddr int, endAddr int) []byte {
	n := endAddr - beginAddr
	r := io.NewSectionReader(memFile, int64(beginAddr), int64(n))
	r.Read(buffer)
	return buffer
}

func WriteMemory(memFile *os.File, targetAddr int, targetVal []byte) error {
	if _, err := memFile.WriteAt(targetVal, int64(targetAddr)); err != nil {
		return err
	}
	return nil
}
