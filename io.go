package gover

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func (snap Snapshot) Write(snapshotPath string) {
	f, err := os.Create(snapshotPath)

	if err != nil {
		panic(fmt.Sprintf("Error: Could not create snapshot file %s", snapshotPath))
	}
	myEncoder := json.NewEncoder(f)
	myEncoder.SetIndent("", "  ")
	myEncoder.Encode(snap)
	f.Close()
}

func ReadSnapshot(snapId string) Snapshot {
	snapshotPath := filepath.Join(".gover", "snapshots", snapId+".json")

	if VerboseMode {
		fmt.Printf("Reading %s\n", snapshotPath)
	}

	return ReadSnapshotFile(snapId)
}

// Read a snapshot given a file path
func ReadSnapshotFile(snapshotPath string) Snapshot {
	var mySnapshot Snapshot
	f, err := os.Open(snapshotPath)

	if err != nil {
		// panic(fmt.Sprintf("Error: Could not read snapshot file %s", snapshotPath))
		return Snapshot{Files: []string{}, StoredFiles: make(map[string]string), ModTimes: make(map[string]string)}
	}

	myDecoder := json.NewDecoder(f)

	if err := myDecoder.Decode(&mySnapshot); err != nil {
		fmt.Printf("Error:could not decode head file %s\n", snapshotPath)
		Check(err)
	}

	f.Close()
	return mySnapshot
}
