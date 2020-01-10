package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-ps"

	prompt "github.com/c-bata/go-prompt"
	sys "golang.org/x/sys/unix"
)

func plist() {
	processes, err := ps.Processes()
	if err != nil {
		return
	}

	for i, p := range processes {
		fmt.Printf("%d : %+v\n", i, p)
	}
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
		plist()
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
		{Text: "attach"},
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
