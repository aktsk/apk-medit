package cmd

import (
	"bufio"
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
	tid_dir := fmt.Sprintf("/proc/%s/task", pid)
	if _, err := os.Stat(tid_dir); err == nil {
		tidinfo, err := ioutil.ReadDir(tid_dir)
		if err != nil {
			log.Fatal(err)
		}

		tids = []int{}
		for _, t := range tidinfo {
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

func Find(pid string, targetVal string) error {
	// search value in /proc/<pid>/mem
	mapsPath := fmt.Sprintf("/proc/%s/maps", pid)
	addrRanges, err := GetWritableAddrRanges(mapsPath)
	if err != nil {
		return err
	}
	fmt.Println(addrRanges)

	//memPath := fmt.Sprintf("/proc/%s/mem", pid)

	return nil
}

func GetWritableAddrRanges(mapsPath string) ([]string, error) {
	addrRanges := []string{}
	ignorePaths := []string{"/vendor/lib64/", "/system/lib64/", "/system/framework/"}
	file, err := os.Open(mapsPath)
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
				addrRanges = append(addrRanges, addrRange)
			}
		}
	}
	return addrRanges, nil
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

	wpid, err := sys.Wait4(pid, &s, sys.WALL, nil)
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
