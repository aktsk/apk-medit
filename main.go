package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	sys "golang.org/x/sys/unix"
)

var appPID string
var tids []int
var isAttached = false

func plist() error {
	cmd := exec.Command("ps", "-e")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return err
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
		appPID = pids[0]
	}
	return nil
}

func attach(pid string) error {
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

func find(pid string, targetVal string) error {
	// search value in /proc/<pid>/maps
	mapsPath := fmt.Sprintf("/proc/%s/maps", pid)
	addrRanges, err := getWritableAddrRanges(mapsPath)
	if err != nil {
		return err
	}
	fmt.Println(addrRanges)
	return nil
}

func getWritableAddrRanges(mapsPath string) ([]string, error) {
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

func detach() error {
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

func executor(in string) {
	if in == "ps" {
		if err := plist(); err != nil {
			log.Fatal(err)
		}

	} else if strings.HasPrefix(in, "attach") {
		slice := strings.Split(in, " ")
		var pid string
		if len(slice) > 1 {
			pid = slice[1]
		} else if appPID != "" {
			pid = appPID
		} else {
			fmt.Println("PID cannot be specified.")
		}
		attach(pid)

	} else if strings.HasPrefix(in, "find") {
		slice := strings.Split(in, " ")
		var targetVal string
		if len(slice) > 1 {
			targetVal = slice[1]
		} else {
			fmt.Println("Target value cannot be specified.")
		}
		find(appPID, targetVal)

	} else if in == "detach" {
		if err := detach(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	} else if in == "exit" {
		os.Exit(0)

	} else if in == "" {

	} else {
		fmt.Println("Command not found.")
	}
	return
}

func completer(t prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "attach", Description: "Attach to the specified process."},
		{Text: "attach <pid>", Description: "Attach to the process specified on the command line."},
		{Text: "find <int>", Description: "TODO"},
		{Text: "detach", Description: "Detach from the attached process."},
		{Text: "ps", Description: "Find the target process and if there is only one, specify it as the target."},
		{Text: "exit"},
	}
}

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := plist(); err != nil {
		log.Fatal(err)
	}

	p := prompt.New(
		executor,
		completer,
		prompt.OptionTitle("medit: simple MEmory eDIT tool"),
		prompt.OptionPrefix("> "),
		prompt.OptionInputTextColor(prompt.Cyan),
		prompt.OptionPrefixTextColor(prompt.DarkBlue),
		prompt.OptionPreviewSuggestionTextColor(prompt.Green),
		prompt.OptionDescriptionTextColor(prompt.DarkGray),
	)
	p.Run()
}
