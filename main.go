package main

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
	"time"

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
	err := cmd.Run()
	if err != nil {
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

func attach(pid string) {
	if isAttached {
		fmt.Println("Already attached.")
		return
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
				fmt.Println(err)
			}
			if err := wait(tid); err != nil {
				fmt.Printf("Failed wait TID: %d\n", tid)
			}
		}

		//syscall.RawSyscall(syscall.SYS_WAITPID, uintptr(-1), 0, sys.WALL) ない！！！
		/*
			        for _, tid := range tids {
						if err := wait(tid); err == nil {
							fmt.Printf("Wait TID: %d\n", tid)
						} else {
							fmt.Printf("Failed wait TID: %d\n", tid)
						}
			        }*/

		isAttached = true

	} else if os.IsNotExist(err) {
		fmt.Println("PID must be an integer that exists.")
	}
}

func detach() {
	if !isAttached {
		fmt.Println("Already detached.")
		return
	}

	for _, tid := range tids {
		err := sys.PtraceDetach(tid)
		if err == nil {
			fmt.Printf("Detached TID: %d\n", tid)
		} else {
			fmt.Println(err)
		}
	}
	isAttached = false
}

func status(pid int, comm string) rune {
	f, err := os.Open(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return '\000'
	}
	defer f.Close()

	var (
		p     int
		state rune
	)

	// The second field of /proc/pid/stat is the name of the task in parenthesis.
	// The name of the task is the base name of the executable for this process limited to TASK_COMM_LEN characters
	// Since both parenthesis and spaces can appear inside the name of the task and no escaping happens we need to read the name of the executable first
	// See: include/linux/sched.c:315 and include/linux/sched.c:1510
	fmt.Fscanf(f, "%d ("+comm+")  %c", &p, &state)
	return state
}

func wait(pid int) error {
	var s sys.WaitStatus
	for {
		wpid, err := sys.Wait4(pid, &s, sys.WNOHANG|0x40000000|0, nil)
		if err != nil {
			return err
		}
		if wpid != 0 {
			return err
		}
		// Status == Zombie
		// if status(pid, dbp.os.comm) == 'Z' {
		// 	return nil
		// }
		time.Sleep(200 * time.Millisecond)
	}
}

func executor(in string) {
	if in == "ps" {
		if err := plist(); err != nil {
			log.Fatal(err)
		}
	} else if strings.HasPrefix(in, "attach") {
		slice := strings.Split(in, " ")
		if len(slice) > 1 {
			attach(slice[1])
		} else if appPID != "" {
			attach(appPID)
		} else {
			fmt.Println("Cannot attach because PID cannot be specified.")
		}
	} else if in == "detach" {
		detach()
	} else if in == "exit" {
		os.Exit(0)
	} else {
		fmt.Println("Command not found.")
	}
	return
}

func completer(t prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "attach"},
		{Text: "attach <pid>"},
		{Text: "detach"},
		{Text: "exit"},
		{Text: "ps"},
	}
}

func main() {
	if err := plist(); err != nil {
		log.Fatal(err)
	}

	p := prompt.New(
		executor,
		completer,
	)
	p.Run()
}
