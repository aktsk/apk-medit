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

	prompt "github.com/c-bata/go-prompt"
	sys "golang.org/x/sys/unix"
)

var app_pid string
var tids []int
var is_attached = false

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
		app_pid = pids[0]
	}
	return nil
}

func attach(pid string) {
	if is_attached {
		fmt.Println("Already attached.")
		return
	}

	fmt.Printf("Target PID: %s\n", pid)
	tid_dir := "/proc/" + pid + "/task"
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
			err := sys.PtraceAttach(tid)
			if err == nil {
				fmt.Printf("Attached TID: %d\n", tid)
			} else {
				fmt.Println(err)
			}
		}
		is_attached = true
	} else if os.IsNotExist(err) {
		fmt.Println("PID must be an integer that exists.")
	}
}

func detach() {
	if !is_attached {
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
	is_attached = false
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
		} else if app_pid != "" {
			attach(app_pid)
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
