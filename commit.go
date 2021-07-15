package gover

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func CommitSnapshot(message string, filters []string) {
	// Optional timestamp
	t := time.Now()
	ts := t.Format("2006-01-02T15-04-05")
	snap := Snapshot{Time: ts}
	snap.Files = []string{}
	snap.StoredFiles = make(map[string]string)
	snap.ModTimes = make(map[string]string)
	snap.Message = message

	// workingDirectory, err := os.Getwd()
	// Check(err)
	workingDirectory := "."
	head := ReadHead()

	var VersionFile = func(fileName string, info os.FileInfo, err error) error {
		fileName = strings.TrimSuffix(fileName, "\n")

		if ExcludedFile(fileName, info, filters) {
			return nil
		}

		ext := filepath.Ext(fileName)
		hash, hashErr := HashFile(fileName, NumChars)

		if hashErr != nil {
			return hashErr
		}

		verFolder := filepath.Join(".gover", "data", hash[0:2])
		verFile := filepath.Join(verFolder, hash+ext)

		props, err := os.Stat(fileName)

		if err != nil {
			if VerboseMode {
				fmt.Printf("Skipping unreadable file %s\n", fileName)
			}

			return nil
		}

		modTime := props.ModTime().Format("2006-01-02T15-04-05")

		snap.Files = append(snap.Files, fileName)
		snap.StoredFiles[fileName] = verFile
		snap.ModTimes[fileName] = modTime

		os.MkdirAll(verFolder, 0777)

		if headModTime, ok := head.ModTimes[fileName]; ok && modTime == headModTime {
			if VerboseMode {
				fmt.Printf("Skipping %s\n", fileName)
			}
		} else {
			CopyFile(fileName, verFile)

			if VerboseMode {
				fmt.Printf("%s -> %s\n", fileName, verFile)
			} else {
				fmt.Println(fileName)
			}
		}

		return nil
	}

	// fmt.Printf("No changes detected in %s for commit %s\n", workDir, snapshot.ID)
	filepath.Walk(workingDirectory, VersionFile)

	snapFolder := filepath.Join(".gover", "snapshots")
	os.MkdirAll(snapFolder, 0777)
	snapFile := filepath.Join(snapFolder, ts+".json")
	snap.Write(snapFile)
	WriteHead(ts)
}
