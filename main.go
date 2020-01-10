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
	for err == nil && len(line) != 0 {
		s := strings.Split(re.ReplaceAllString(string(line), " "), " ")
		pid := s[1]
		cmd := s[8]
		if pid != "PID" && cmd != "" && cmd != "ps" && cmd != "sh" && cmd != "medit" {
			fmt.Printf("pid: %s, cmd: %s\n", pid, cmd)
		}
		line, err = out.ReadString('\n')
	}

	return nil
}

func attach(pid string) {
	fmt.Printf("Target PID: %s\n", pid)
	tidinfo, err := ioutil.ReadDir("/proc/" + pid + "/task")
	if err != nil {
		log.Fatal(err)
	}

	tids := []int{}
	for _, t := range tidinfo {
		tid, _ := strconv.Atoi(t.Name())
		tids = append(tids, tid)
	}

	for _, tid := range tids {
		sys.PtraceAttach(tid)
		fmt.Printf("Attached TID: %d\n", tid)
	}

	fmt.Println("5s sleep.....")
	time.Sleep(5 * time.Second)

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
		attach(slice[1])
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
