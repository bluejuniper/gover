package snapshots

func WriteHead(snapshotPath string) {
	headPath := filepath.Join(".gover", "head.json")
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
	headPath := filepath.Join(".gover", "head.json")
	f, err := os.Open(headPath)

	if err != nil {
		// panic(fmt.Sprintf("Error: Could not read snapshot file %s", snapshotPath))
		return Snapshot{Files: []string{}, StoredFiles: make(map[string]string), ModTimes: make(map[string]string)}
	}

	snapshotId := ""
	myDecoder := json.NewDecoder(f)

	if err := myDecoder.Decode(&snapshotId); err != nil {
		fmt.Printf("Error:could not decode head file %s\n", headPath)
		check(err)
	}

	f.Close()

	snapshotPath := filepath.Join(".gover", "snapshots", snapshotId + ".json")
	return ReadSnapshotFile(snapshotPath)
}