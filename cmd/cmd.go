package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	sys "golang.org/x/sys/unix"
)

var tids []int
var isAttached = false

func Plist() (string, error) {
	cmd := exec.Command("ps", "-e")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}

	re := regexp.MustCompile(`\s+`)
	line, err := out.ReadString('\n')
	pids := []string{}
	for err == nil && len(line) != 0 {
		s := strings.Split(re.ReplaceAllString(string(line), " "), " ")
		pid := s[1]
		cmd := s[8]
		if pid != "PID" && cmd != "" && cmd != "ps" && cmd != "sh" && cmd != "medit" {
			fmt.Printf("Target Package: %s, PID: %s\n", cmd, pid)
			pids = append(pids, pid)
		}
		line, err = out.ReadString('\n')
	}

	if len(pids) == 1 {
		fmt.Printf("Attach target PID has been set to %s.\n", pids[0])
		return pids[0], nil
	}
	return "", nil
}

func Attach(pid string) error {
	if isAttached {
		fmt.Println("Already attached.")
		return nil
	}

	fmt.Printf("Target PID: %s\n", pid)
	tidDir := fmt.Sprintf("/proc/%s/task", pid)
	if _, err := os.Stat(tidDir); err == nil {
		tidInfo, err := ioutil.ReadDir(tidDir)
		if err != nil {
			log.Fatal(err)
		}

		tids = []int{}
		for _, t := range tidInfo {
			tid, _ := strconv.Atoi(t.Name())
			tids = append(tids, tid)
		}

		for _, tid := range tids {
			if err := sys.PtraceAttach(tid); err == nil {
				fmt.Printf("Attached TID: %d\n", tid)
			} else {
				fmt.Printf("attach failed: %s\n", err)
			}
			if err := wait(tid); err != nil {
				fmt.Printf("Failed wait TID: %d, %s\n", tid, err)
			}
		}

		isAttached = true

	} else if os.IsNotExist(err) {
		fmt.Println("PID must be an integer that exists.")
	}
	return nil
}

func Find(pid string, targetVal string) ([]int, error) {
	// search value in /proc/<pid>/mem
	mapsPath := fmt.Sprintf("/proc/%s/maps", pid)
	addrRanges, err := getWritableAddrRanges(mapsPath)
	if err != nil {
		return nil, err
	}

	memPath := fmt.Sprintf("/proc/%s/mem", pid)
	//targetBytes := stringToBytes(targetVal)
	targetBytes, _ := wordToBytes(targetVal)
	fmt.Printf("Target Value: %s(%v)\n", targetVal, targetBytes)
	foundAddrs, _ := findDataInAddrRanges(memPath, targetBytes, addrRanges)
	fmt.Printf("Found: 0x%x!!!\n", len(foundAddrs))
	if len(foundAddrs) < 10 {
		for _, v := range foundAddrs {
			fmt.Printf("Address: 0x%x\n", v)
		}
	}
	fmt.Println(foundAddrs)
	return foundAddrs, nil
}

func Filter(pid string, targetVal string, prevAddrs []int) ([]int, error) {
	mapsPath := fmt.Sprintf("/proc/%s/maps", pid)
	writableAddrRanges, err := getWritableAddrRanges(mapsPath)
	if err != nil {
		return nil, err
	}
	//targetBytes := stringToBytes(targetVal)
	targetBytes, _ := wordToBytes(targetVal)
	targetLength := len(targetBytes)
	addrRanges := [][2]int{}
	// check if previous result address exists in current memory map
	for _, prevAddr := range prevAddrs {
		for _, writable := range writableAddrRanges {
			if writable[0] < prevAddr && prevAddr < writable[1] {
				addrRanges = append(addrRanges, [2]int{prevAddr, prevAddr + targetLength})
			}
		}
	}
	fmt.Printf("Target Value: %s(%v)\n", targetVal, targetBytes)
	memPath := fmt.Sprintf("/proc/%s/mem", pid)
	foundAddrs, _ := findDataInAddrRanges(memPath, targetBytes, addrRanges)
	fmt.Printf("Found: 0x%x!!!\n", len(foundAddrs))
	if len(foundAddrs) < 10 {
		for _, v := range foundAddrs {
			fmt.Printf("Address: 0x%x\n", v)
		}
	}
	return foundAddrs, nil
}

func Patch(pid string, targetVal string, targetAddrs []int) error {
	memPath := fmt.Sprintf("/proc/%s/mem", pid)
	f, err := os.OpenFile(memPath, os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Println(targetAddrs)
	targetBytes, _ := stringToBytes(targetVal)
	fmt.Println(targetBytes)
	for _, v := range targetAddrs {
		err := writeMemory(f, v, targetBytes)
		if err != nil {
			return err
		}
	}
	return nil
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

var splitSize = 0x50000000
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, splitSize)
	},
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

func Detach() error {
	if !isAttached {
		fmt.Println("Already detached.")
		return nil
	}

	for _, tid := range tids {
		if err := sys.PtraceDetach(tid); err != nil {
			return fmt.Errorf("%d detach failed. %s\n", tid, err)
		} else {
			fmt.Printf("Detached TID: %d\n", tid)
		}
	}

	isAttached = false
	return nil
}

func wait(pid int) error {
	var s sys.WaitStatus

	// sys.WALL = 0x40000000 on Linux(ARM64)
	// Using sys.WALL does not pass test on macOS.
	// https://github.com/golang/go/blob/50bd1c4d4eb4fac8ddeb5f063c099daccfb71b26/src/syscall/zerrors_linux_arm.go#L1203
	wpid, err := sys.Wait4(pid, &s, 0x40000000, nil)
	if err != nil {
		return err
	}

	if wpid != pid {
		return fmt.Errorf("wait failed: wpid = %d", wpid)
	}
	if !s.Stopped() {
		return fmt.Errorf("wait failed: status is not stopped: ")
	}

	return nil
}
