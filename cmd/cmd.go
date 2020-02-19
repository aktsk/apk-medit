package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	sys "golang.org/x/sys/unix"
)

var tids []int
var isAttached = false

type Found struct {
	addrs     []int
	converter func(string) ([]byte, error)
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

func Find(pid string, targetVal string) ([]Found, error) {
	founds := []Found{}
	// search value in /proc/<pid>/mem
	mapsPath := fmt.Sprintf("/proc/%s/maps", pid)
	memPath := fmt.Sprintf("/proc/%s/mem", pid)
	addrRanges, err := getWritableAddrRanges(mapsPath)
	if err != nil {
		return nil, err
	}

	targetBytes, _ := stringToBytes(targetVal)
	fmt.Println("Search UTF-8 String...")
	fmt.Printf("Target Value: %s(%v)\n", targetVal, targetBytes)
	foundBeginAddrs, _ := findDataInAddrRanges(memPath, targetBytes, addrRanges)
	fmt.Printf("Found: %d!!!\n", len(foundBeginAddrs))
	if len(foundBeginAddrs) < 10 {
		for _, v := range foundBeginAddrs {
			fmt.Printf("Address: 0x%x\n", v)
		}
	}
	fmt.Println(foundBeginAddrs)
	founds = append(founds, Found{
		addrs:     foundBeginAddrs,
		converter: stringToBytes,
	})

	targetBytes, _ = dwordToBytes(targetVal)
	fmt.Println("Search Double Word...")
	fmt.Printf("Target Value: %s(%v)\n", targetVal, targetBytes)
	foundBeginAddrs, _ = findDataInAddrRanges(memPath, targetBytes, addrRanges)
	fmt.Printf("Found: %d!!!\n", len(foundBeginAddrs))
	if len(foundBeginAddrs) < 10 {
		for _, v := range foundBeginAddrs {
			fmt.Printf("Address: 0x%x\n", v)
		}
	}
	fmt.Println(foundBeginAddrs)
	founds = append(founds, Found{
		addrs:     foundBeginAddrs,
		converter: dwordToBytes,
	})
	return founds, nil
}

func Filter(pid string, targetVal string, prevAddrs []Found) ([]Found, error) {
	founds := []Found{}
	mapsPath := fmt.Sprintf("/proc/%s/maps", pid)
	memPath := fmt.Sprintf("/proc/%s/mem", pid)
	writableAddrRanges, err := getWritableAddrRanges(mapsPath)
	if err != nil {
		return nil, err
	}
	addrRanges := [][2]int{}
	// check if previous result address exists in current memory map
	for _, foundAddr := range prevAddrs {
		targetBytes, _ := foundAddr.converter(targetVal)
		targetLength := len(targetBytes)
		fmt.Printf("Target Value: %s(%v)\n", targetVal, targetBytes)
		for _, prevAddr := range foundAddr.addrs {
			for _, writable := range writableAddrRanges {
				if writable[0] < prevAddr && prevAddr < writable[1] {
					addrRanges = append(addrRanges, [2]int{prevAddr, prevAddr + targetLength})
				}
			}
		}
		foundBeginAddrs, _ := findDataInAddrRanges(memPath, targetBytes, addrRanges)
		fmt.Printf("Found: 0x%x!!!\n", len(foundBeginAddrs))
		if len(foundBeginAddrs) < 10 {
			for _, v := range foundBeginAddrs {
				fmt.Printf("Address: 0x%x\n", v)
			}
		}
		founds = append(founds, Found{
			addrs:     foundBeginAddrs,
			converter: foundAddr.converter,
		})
	}
	return founds, nil
}

func Patch(pid string, targetVal string, targetAddrs []Found) error {
	memPath := fmt.Sprintf("/proc/%s/mem", pid)
	f, err := os.OpenFile(memPath, os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, found := range targetAddrs {
		targetBytes, _ := found.converter(targetVal)
		for _, targetAddr := range found.addrs {
			err := writeMemory(f, targetAddr, targetBytes)
			if err != nil {
				return err
			}
		}
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
