package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

var splitSize = 0x50000000
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, splitSize)
	},
}

func getWritableAddrRanges(mapsPath string) ([][2]int, error) {
	addrRanges := [][2]int{}
	ignorePaths := []string{"/vendor/lib64/", "/system/lib64/", "/system/bin/", "/system/framework/", "/data/dalvik-cache/"}
	file, err := os.OpenFile(mapsPath, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		meminfo := strings.Fields(line)
		addrRange := meminfo[0]
		permission := meminfo[1]
		if permission[0] == 'r' && permission[1] == 'w' && permission[3] != 's' {
			ignoreFlag := false
			if len(meminfo) >= 6 {
				filePath := meminfo[5]
				for _, ignorePath := range ignorePaths {
					if strings.HasPrefix(filePath, ignorePath) {
						ignoreFlag = true
						break
					}
				}
			}

			if !ignoreFlag {
				addrs := strings.Split(addrRange, "-")
				beginAddr, _ := strconv.ParseInt(addrs[0], 16, 64)
				endAddr, _ := strconv.ParseInt(addrs[1], 16, 64)
				addrRanges = append(addrRanges, [2]int{int(beginAddr), int(endAddr)})
			}
		}
	}
	return addrRanges, nil
}

func findDataInAddrRanges(memPath string, targetBytes []byte, addrRanges [][2]int) ([]int, error) {
	foundAddrs := []int{}
	f, err := os.OpenFile(memPath, os.O_RDONLY, 0600)
	defer f.Close()

	searchLength := len(targetBytes)
	for _, s := range addrRanges {
		beginAddr := s[0]
		endAddr := s[1]
		memSize := endAddr - beginAddr
		if err != nil {
			fmt.Println(err)
		}
		for i := 0; i < (memSize/splitSize)+1; i++ {
			splitIndex := (i + 1) * splitSize
			splittedBeginAddr := beginAddr + i*splitSize
			splittedEndAddr := endAddr
			if splitIndex < memSize {
				splittedEndAddr = beginAddr + splitIndex
			}
			b := bufferPool.Get().([]byte)[:(splittedEndAddr - splittedBeginAddr)]
			readMemory(f, b, splittedBeginAddr, splittedEndAddr)
			//fmt.Printf("Memory size: 0x%x bytes\n", len(b))
			//fmt.Printf("Begin Address: 0x%x, End Address 0x%x\n", splittedBeginAddr, splittedEndAddr)
			findDataInSplittedMemory(&b, targetBytes, searchLength, splittedBeginAddr, 0, &foundAddrs)
			bufferPool.Put(b)
			if len(foundAddrs) > 10000 {
				fmt.Println("Too many addresses with target data found...")
				return foundAddrs, nil
			}
		}
	}

	return foundAddrs, nil
}

func findDataInSplittedMemory(memory *[]byte, targetBytes []byte, searchLength int, beginAddr int, offset int, results *[]int) {
	// use Rabin-Karp string search algorithm in bytes.Index
	index := bytes.Index((*memory)[offset:], targetBytes)
	if index == -1 {
		return
	} else {
		resultAddr := beginAddr + index + offset
		*results = append(*results, resultAddr)
		offset += index + searchLength
		findDataInSplittedMemory(memory, targetBytes, searchLength, beginAddr, offset, results)
	}
}
