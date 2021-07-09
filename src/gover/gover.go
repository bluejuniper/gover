package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/akbarnes/gover/src/options"
	"github.com/akbarnes/gover/src/snapshots"
)

func ReadFilters() []string {
	filterPath := ".gover_ignore.json"
	var filters []string
	f, err := os.Open(filterPath)

	if err != nil {
		// panic(fmt.Sprintf("Error: Could not read snapshot file %s", snapshotPath))
		return []string{}
	}

	myDecoder := json.NewDecoder(f)

	if err := myDecoder.Decode(&filters); err != nil {
		panic(fmt.Sprintf("Error:could not decode filter file %s", filterPath))
	}

	f.Close()
	return filters
}

var Message string
var OutputFolder string

func init() {
	flag.BoolVar(&options.JsonMode, "json", false, "print json")
	flag.BoolVar(&options.JsonMode, "j", false, "print json")
	flag.BoolVar(&options.VerboseMode, "verbose", false, "verbose mode")
	flag.BoolVar(&options.VerboseMode, "v", false, "verbose mode")
}

func main() {
	commitCmd := flag.NewFlagSet("commit", flag.ExitOnError)
	statusCmd := flag.NewFlagSet("status", flag.ExitOnError)
	logCmd := flag.NewFlagSet("log", flag.ExitOnError)
	checkoutCmd := flag.NewFlagSet("checkout", flag.ExitOnError)

	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Println("Expected subcommand")
		os.Exit(1)
	}

	cmd := os.Args[1]

	if cmd == "commit" || cmd == "ci" {
		commitCmd.Parse(os.Args[2:])
		filters := ReadFilters()

		if commitCmd.NArg() >= 1 {
			Message = commitCmd.Arg(0)
		}

		snapshots.CommitSnapshot(Message, filters)
	} else if cmd == "status" || cmd == "st" {
		statusCmd.Parse(os.Args[2:])
		filters := ReadFilters()

		if statusCmd.NArg() >= 1 {
			snapshots.DiffSnapshot(statusCmd.Arg(0), filters)
		} else {
			snapshots.DiffSnapshot("", filters)
		}
	} else if cmd == "log" {
		logCmd.Parse(os.Args[2:])

		if logCmd.NArg() >= 1 {
			snapshotNum, _ := strconv.Atoi(logCmd.Arg(0))
			snapshots.LogSingleSnapshot(snapshotNum)
		} else {
			snapshots.LogAllSnapshots()
		}
	} else if cmd == "checkout" || cmd == "co" {
		checkoutCmd.StringVar(&OutputFolder, "out", "", "output folder")
		checkoutCmd.StringVar(&OutputFolder, "o", "", "output folder")
		checkoutCmd.Parse(os.Args[2:])
		snapshotNum, _ := strconv.Atoi(checkoutCmd.Arg(0))
		snapshots.CheckoutSnaphot(snapshotNum, OutputFolder)
	} else {
		fmt.Println("Unknown subcommand")
		os.Exit(1)
	}
}
