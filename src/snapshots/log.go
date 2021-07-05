package snapshots

import (
	"fmt"
	"path/filepath"
	"github.com/akbarnes/gover/src/util"
	"github.com/akbarnes/gover/src/options"
)

func LogAllSnapshots() {
	if options.JsonMode {
		type Snap struct {
			Time string
			Message string
		}

		snaps := []Snap{}
		
		snapshotGlob := filepath.Join(".gover","snapshots","*.json")
		snapshotPaths, err := filepath.Glob(snapshotGlob)
		util.Check(err)

		for _, snapshotPath := range snapshotPaths {
			snapshot := ReadSnapshotFile(snapshotPath)
			snap := Snap{Time: snapshot.Time, Message: snapshot.Message}
			snaps = append(snaps, snap)
		}

		util.PrintJson(snaps)
	} else {
		snapshotGlob := filepath.Join(".gover", "snapshots", "*.json")
		snapshotPaths, err := filepath.Glob(snapshotGlob)
		util.Check(err)

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
	util.Check(err)

	snapshotPath := snapshotPaths[snapshotNum - 1]

	snap := ReadSnapshotFile(snapshotPath)

	if options.JsonMode {
		type SnapshotFile struct {
			File string
			ModTime string
			StoredFile string
		}

		snapFiles := []SnapshotFile{}

		for _, file := range snap.Files {
			snapFile := SnapshotFile{File: file, ModTime: snap.ModTimes[file], StoredFile:snap.StoredFiles[file]}
			snapFiles = append(snapFiles, snapFile)
		}

		util.PrintJson(snapFiles)
	} else {
		for _, file := range snap.Files {
			fmt.Println(file)
		}		
	}	
}
