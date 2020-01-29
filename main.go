package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/aktsk/medit/cmd"

	prompt "github.com/c-bata/go-prompt"
)

var appPID string

func executor(in string) {
	if in == "ps" {
		if pid, err := cmd.Plist(); err != nil {
			log.Fatal(err)
		} else if pid != "" {
			appPID = pid
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
		cmd.Attach(pid)

	} else if strings.HasPrefix(in, "find") {
		slice := strings.Split(in, " ")
		var targetVal uint64
		if len(slice) > 1 {
			targetVal, _ = strconv.ParseUint(slice[1], 64, 10)
		} else {
			fmt.Println("Target value cannot be specified.")
		}
		cmd.Find(appPID, targetVal)

	} else if in == "detach" {
		if err := cmd.Detach(); err != nil {
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

	if pid, err := cmd.Plist(); err != nil {
		log.Fatal(err)
	} else if pid != "" {
		appPID = pid
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
