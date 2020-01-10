package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	sys "golang.org/x/sys/unix"
)

func main() {
	pid := os.Args[1]
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
