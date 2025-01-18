package cmd

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	sys "golang.org/x/sys/unix"

	"github.com/sterrasec/apk-medit/pkg/converter"
	"github.com/sterrasec/apk-medit/pkg/memory"
)

var tids []int
var isAttached = false

type Found struct {
	addrs     []int
	converter func(string) ([]byte, error)
	dataType  string
}

func Plist() (string, error) {
	cmd := exec.Command("ps", "-e")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}

	re := regexp.MustCompile(`\s+`)
	line, err := out.ReadString('\n')
	pids := make(map[string]string)
	for err == nil && len(line) != 0 {
		s := strings.Split(re.ReplaceAllString(string(line), " "), " ")
		pid := s[1]
		cmd := s[8]
		if pid != "PID" && cmd != "" && cmd != "ps" && cmd != "sh" && cmd != "medit" {
			fmt.Printf("Package: %s, PID: %s\n", cmd, pid)
			pids[cmd] = pid
		}
		line, err = out.ReadString('\n')
	}

	current_path, _ := os.Getwd()
	_, package_name := filepath.Split(current_path)
	for cmd, pid := range pids {
		if cmd == package_name {
			fmt.Printf("Target PID has been set to %s.\n", pid)
			return pid, nil
		}
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

func Find(pid string, targetVal string, dataType string) ([]Found, error) {
	founds := []Found{}
	// parse /proc/<pid>/map, and get writable area
	mapsPath := fmt.Sprintf("/proc/%s/maps", pid)
	addrRanges, err := memory.GetWritableAddrRanges(mapsPath)
	// search value in /proc/<pid>/mem
	memPath := fmt.Sprintf("/proc/%s/mem", pid)
	if err != nil {
		return nil, err
	}

	if dataType == "all" {
		// search string
		foundAddrs, err := memory.FindString(memPath, targetVal, addrRanges)
		if err == nil && len(foundAddrs) > 0 {
			founds = append(founds, Found{
				addrs:     foundAddrs,
				converter: converter.StringToBytes,
				dataType:  "UTF-8 string",
			})
		} else if _, ok := err.(memory.TooManyErr); ok {
			return founds, err
		}
		fmt.Println("------------------------")

		// search int
		foundAddrs, err = memory.FindWord(memPath, targetVal, addrRanges)
		if err == nil {
			if len(foundAddrs) > 0 {
				founds = append(founds, Found{
					addrs:     foundAddrs,
					converter: converter.WordToBytes,
					dataType:  "word",
				})
			}
			return founds, nil
		} else if _, ok := err.(memory.TooManyErr); ok {
			return founds, err
		}
		fmt.Println("------------------------")
		foundAddrs, err = memory.FindDword(memPath, targetVal, addrRanges)
		if err == nil {
			if len(foundAddrs) > 0 {
				founds = append(founds, Found{
					addrs:     foundAddrs,
					converter: converter.DwordToBytes,
					dataType:  "dword",
				})
			}
			return founds, nil
		} else if _, ok := err.(memory.TooManyErr); ok {
			return founds, err
		}
		fmt.Println("------------------------")
		foundAddrs, err = memory.FindQword(memPath, targetVal, addrRanges)
		if err == nil {
			if len(foundAddrs) > 0 {
				founds = append(founds, Found{
					addrs:     foundAddrs,
					converter: converter.QwordToBytes,
					dataType:  "qword",
				})
			}
			return founds, nil
		} else if _, ok := err.(memory.TooManyErr); ok {
			return founds, err
		}

	} else if dataType == "string" {
		foundAddrs, _ := memory.FindString(memPath, targetVal, addrRanges)
		if err == nil {
			if len(foundAddrs) > 0 {
				founds = append(founds, Found{
					addrs:     foundAddrs,
					converter: converter.StringToBytes,
					dataType:  "UTF-8 string",
				})
			}
			return founds, nil
		} else if _, ok := err.(memory.TooManyErr); ok {
			return founds, err
		}

	} else if dataType == "word" {
		foundAddrs, err := memory.FindWord(memPath, targetVal, addrRanges)
		if err == nil {
			if len(foundAddrs) > 0 {
				founds = append(founds, Found{
					addrs:     foundAddrs,
					converter: converter.WordToBytes,
					dataType:  "word",
				})
			}
			return founds, nil
		} else if _, ok := err.(memory.TooManyErr); ok {
			return founds, err
		}

	} else if dataType == "dword" {
		foundAddrs, err := memory.FindDword(memPath, targetVal, addrRanges)
		if err == nil {
			if len(foundAddrs) > 0 {
				founds = append(founds, Found{
					addrs:     foundAddrs,
					converter: converter.DwordToBytes,
					dataType:  "dword",
				})
			}
			return founds, nil
		} else if _, ok := err.(memory.TooManyErr); ok {
			return founds, err
		}

	} else if dataType == "qword" {
		foundAddrs, err := memory.FindQword(memPath, targetVal, addrRanges)
		if err == nil {
			if len(foundAddrs) > 0 {
				founds = append(founds, Found{
					addrs:     foundAddrs,
					converter: converter.QwordToBytes,
					dataType:  "qword",
				})
			}
			return founds, nil
		} else if _, ok := err.(memory.TooManyErr); ok {
			return founds, err
		}
	}

	return nil, errors.New("Error: specified datatype does not exist")
}

func Filter(pid string, targetVal string, prevFounds []Found) ([]Found, error) {
	founds := []Found{}
	mapsPath := fmt.Sprintf("/proc/%s/maps", pid)
	memPath := fmt.Sprintf("/proc/%s/mem", pid)
	writableAddrRanges, err := memory.GetWritableAddrRanges(mapsPath)
	if err != nil {
		return nil, err
	}
	addrRanges := [][2]int{}

	// check if previous result address exists in current memory map
	for i, prevFound := range prevFounds {
		targetBytes, _ := prevFound.converter(targetVal)
		targetLength := len(targetBytes)
		fmt.Printf("Check previous results of searching %s...\n", prevFound.dataType)
		fmt.Printf("Target Value: %s(%v)\n", targetVal, targetBytes)
		for _, prevAddr := range prevFound.addrs {
			for _, writable := range writableAddrRanges {
				if writable[0] < prevAddr && prevAddr < writable[1] {
					addrRanges = append(addrRanges, [2]int{prevAddr, prevAddr + targetLength})
				}
			}
		}
		foundAddrs, _ := memory.FindDataInAddrRanges(memPath, targetBytes, addrRanges)
		fmt.Printf("Found: %d!!\n", len(foundAddrs))
		if len(foundAddrs) < 10 {
			for _, v := range foundAddrs {
				fmt.Printf("Address: 0x%x\n", v)
			}
		}
		founds = append(founds, Found{
			addrs:     foundAddrs,
			converter: prevFound.converter,
			dataType:  prevFound.dataType,
		})
		if i != len(prevFounds)-1 {
			fmt.Println("------------------------")
		}
	}
	return founds, nil
}

func PatchWithoutPtrace(pid string, targetVal string, targetAddrs []Found) error {
	memPath := fmt.Sprintf("/proc/%s/mem", pid)
	f, err := os.OpenFile(memPath, os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, found := range targetAddrs {
		targetBytes, _ := found.converter(targetVal)
		for _, targetAddr := range found.addrs {
			err := memory.WriteMemory(f, targetAddr, targetBytes)
			if err != nil {
				return err
			}
		}
	}
	fmt.Println("Successfully patched!")
	return nil
}

func PatchWithPtrace(pid string, targetVal string, targetAddrs []Found) error {
	if !isAttached {
		if err := Attach(pid); err != nil {
			fmt.Println(err)
		}
	}
	for _, found := range targetAddrs {
		targetBytes, _ := found.converter(targetVal)
		for _, targetAddr := range found.addrs {
			tid_int, _ := strconv.Atoi(pid)
			_, err := sys.PtracePokeData(tid_int, uintptr(targetAddr), targetBytes)
			if err != nil {
				return err
			}
		}
	}
	Detach()
	fmt.Println("Successfully patched!")
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

func Dump(pid string, beginAddress int, endAddress int) error {
	memPath := fmt.Sprintf("/proc/%s/mem", pid)
	memFile, _ := os.Open(memPath)
	defer memFile.Close()

	memSize := endAddress - beginAddress
	buf := make([]byte, memSize)
	memory := memory.ReadMemory(memFile, buf, beginAddress, endAddress)
	fmt.Printf("Address range: 0x%x - 0x%x\n", beginAddress, endAddress)
	fmt.Println("--------------------------------------------")
	fmt.Printf("%s", hex.Dump(memory))
	return nil
}
