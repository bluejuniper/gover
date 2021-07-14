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
	snapshotPath := filepath.Join(".gover2", "snapshots", snapId+".json")

	if VerboseMode {
		fmt.Printf("Reading %s\n", snapshotPath)
	}

	return ReadSnapshotFile(snapId)
}

// Read a snapshot given a file path
func ReadSnapshotFile(snapshotPath string) Snapshot {
	var snap Snapshot
	f, err := os.Open(snapshotPath)

	// ChunkPackIds map[string]string
	// FileChunkIds map[string][]string
	// FileModTimes map[string]string

	if err != nil {
		// panic(fmt.Sprintf("Error: Could not read snapshot file %s", snapshotPath))
		snap := Snapshot{}
		snap.ChunkPackIds = make(map[string]string)
		snap.FileChunkIds = make(map[string][]string)
		snap.FileModTimes = make(map[string]string)
		return snap
	}

	myDecoder := json.NewDecoder(f)

	if err := myDecoder.Decode(&snap); err != nil {
		fmt.Printf("Error:could not decode head file %s\n", snapshotPath)
		Check(err)
	}

	f.Close()
	return snap
}

func WriteHead(snapshotPath string) {
	headPath := filepath.Join(".gover2", "head.json")
	f, err := os.Create(headPath)

	if err != nil {
		panic(fmt.Sprintf("Error: Could not create head file %s", headPath))
	}

	myEncoder := json.NewEncoder(f)
	myEncoder.SetIndent("", "  ")
	myEncoder.Encode(snapshotPath)
	f.Close()
}

// Read a snapshot given a file path
func ReadHead() Snapshot {
	headPath := filepath.Join(".gover2", "head.json")
	f, err := os.Open(headPath)

	if err != nil {
		// panic(fmt.Sprintf("Error: Could not read snapshot file %s", snapshotPath))
		snap := Snapshot{}
		snap.ChunkPackIds = make(map[string]string)
		snap.FileChunkIds = make(map[string][]string)
		snap.FileModTimes = make(map[string]string)
		return snap
	}

	snapshotId := ""
	myDecoder := json.NewDecoder(f)

	if err := myDecoder.Decode(&snapshotId); err != nil {
		fmt.Printf("Error:could not decode head file %s\n", headPath)
		Check(err)
	}

	f.Close()

	snapshotPath := filepath.Join(".gover2", "snapshots", snapshotId+".json")
	return ReadSnapshotFile(snapshotPath)
}
