package snapshots

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
	"strings"
	"strconv"
	"path/filepath"
	"encoding/json"
	"github.com/bmatcuk/doublestar/v4"
)

func LogAllSnapshots() {
	if JsonMode {
		type Snap struct {
			Time string
			Message string
		}

		snaps := []Snap{}
		
		snapshotGlob := filepath.Join(".gover","snapshots","*.json")
		snapshotPaths, err := filepath.Glob(snapshotGlob)
		check(err)

		for _, snapshotPath := range snapshotPaths {
			snapshot := ReadSnapshotFile(snapshotPath)
			snap := Snap{Time: snapshot.Time, Message: snapshot.Message}
			snaps = append(snaps, snap)
		}

		PrintJson(snaps)
	} else {
		snapshotGlob := filepath.Join(".gover", "snapshots", "*.json")
		snapshotPaths, err := filepath.Glob(snapshotGlob)
		check(err)

		for i, snapshotPath := range snapshotPaths {
			snap := ReadSnapshotFile(snapshotPath)
			// Time: 2021/05/08 08:57:46
			// Message: specify workdir path explicitly
			fmt.Printf("%3d) Time: %s\n", i + 1, snap.Time)

			if len(snap.Message) > 0 {
				fmt.Printf("Message: %s\n\n", snap.Message)
			}
		}
	}
}

func LogSingleSnapshot(snapshotNum int) {
	snapshotGlob := filepath.Join(".gover", "snapshots", "*.json")
	snapshotPaths, err := filepath.Glob(snapshotGlob)
	check(err)

	snapshotPath := snapshotPaths[snapshotNum - 1]

	snap := ReadSnapshotFile(snapshotPath)

	if JsonMode {
		type SnapshotFile struct {
			File string
			StoredFile string
		}

		snapFiles := []SnapshotFile{}

		for _, file := range snap.Files {
			snapFile := SnapshotFile{File: file, StoredFile:snap.StoredFiles[file]}
			snapFiles = append(snapFiles, snapFile)
		}

		PrintJson(snapFiles)
	} else {
		for _, file := range snap.Files {
			fmt.Println(file)
		}		
	}	
}
