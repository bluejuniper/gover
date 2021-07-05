package snapshots

import (
	"fmt"
	"os"
	"time"
	"strings"
	"path/filepath"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/akbarnes/gover/src/util"
	"github.com/akbarnes/gover/src/options"	
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
	// util.Check(err)
	workingDirectory := "."
	head := ReadHead()

	goverDir := filepath.Join(workingDirectory, ".gover", "**")

	var VersionFile = func(fileName string, info os.FileInfo, err error) error {
		fileName = strings.TrimSuffix(fileName, "\n")

		if info.IsDir() {
			return nil
		}

		matched, err := doublestar.PathMatch(goverDir, fileName)

		if matched {
			if options.VerboseMode {
				fmt.Printf("Skipping file %s in .gover\n", fileName)
			}

			return nil
		}

		for _, pattern := range filters {
			matched, err := doublestar.PathMatch(pattern, fileName)

			util.Check(err)
			if matched {
				if options.VerboseMode {
					fmt.Printf("Skipping file %s which matches with %s\n", fileName, pattern)
				}

				return nil
			}
		}

		ext := filepath.Ext(fileName)
		hash, hashErr := util.HashFile(fileName, util.NumChars)

		if hashErr != nil {
			return hashErr
		}

		verFolder := filepath.Join(".gover", "data", hash[0:2]) 
		verFile := filepath.Join(verFolder, hash + ext)

		props, err := os.Stat(fileName)

		if err != nil {
			if options.VerboseMode {
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
			if options.VerboseMode {
				fmt.Printf("Skipping %s\n", fileName)
			}
		} else {
			util.CopyFile(fileName, verFile)

			if !options.JsonMode {
				if options.VerboseMode {
						fmt.Printf("%s -> %s\n", fileName, verFile)
				} else {
					fmt.Println(fileName)
				}
			}
		}

		return nil
	}

	// fmt.Printf("No changes detected in %s for commit %s\n", workDir, snapshot.ID)
	filepath.Walk(workingDirectory, VersionFile)

	if options.JsonMode {
		util.PrintJson(snap)
	}

	snapFolder := filepath.Join(".gover", "snapshots")
	os.MkdirAll(snapFolder, 0777)
	snapFile := filepath.Join(snapFolder, ts + ".json")
	snap.Write(snapFile)
	WriteHead(ts)
}