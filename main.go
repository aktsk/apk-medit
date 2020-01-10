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
			fmt.Printf("cmd: %s, pid: %s\n", cmd, pid)
			pids = append(pids, pid)
		}
		line, err = out.ReadString('\n')
	}

	if len(pids) == 1 {
		fmt.Printf("attach target PID has been set to %s.\n", pids[0])
		app_pid = pids[0]
	}
	return nil
}

func attach(pid string) {
	fmt.Printf("Target PID: %s\n", pid)
	tidinfo, err := ioutil.ReadDir("/proc/" + pid + "/task")
	if err != nil {
		log.Fatal(err)
	}

	tids = []int{}
	for _, t := range tidinfo {
		tid, _ := strconv.Atoi(t.Name())
		tids = append(tids, tid)
	}

	for _, tid := range tids {
		sys.PtraceAttach(tid)
		fmt.Printf("Attached TID: %d\n", tid)
	}

}

func detach() {
	for _, tid := range tids {
		sys.PtraceDetach(tid)
		fmt.Printf("Detached TID: %d\n", tid)
	}
}

func executor(t string) {
	if t == "ps" {
		if err := plist(); err != nil {
			log.Fatal(err)
		}
	}

	if strings.HasPrefix(t, "attach") {
		slice := strings.Split(t, " ")
		if len(slice) > 1 {
			attach(slice[1])
		} else if app_pid != "" {
			attach(app_pid)
		} else {
			fmt.Printf("Cannot attach because PID cannot be specified.")
		}
	}

	if t == "detach" {
		detach()
	}

	if t == "exit" {
		os.Exit(0)
	}
	return
}

func completer(t prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "ps"},
		{Text: "attach <pid>"},
		{Text: "detach"},
		{Text: "exit"},
	}
}

func main() {
	p := prompt.New(
		executor,
		completer,
	)
	p.Run()
}
