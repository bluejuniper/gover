package snapshots

import (
	"fmt"
	"os"
	"strings"
	"path/filepath"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/akbarnes/gover/src/util"
	"github.com/akbarnes/gover/src/options"
)

func DiffSnapshot(snapId string, filters []string) {
	var snap Snapshot

	if len(snapId) > 0 {
		snap = ReadSnapshot(snapId)
	} else {
		snap = ReadHead()
	}

	status := make(map[string]string)

	for _, fileName := range snap.Files {
		status[fileName] = "-"
	}

	// workingDirectory, err := os.Getwd()
	// util.Check(err)
	workingDirectory := "."
	head := ReadHead()

	goverDir := filepath.Join(workingDirectory, ".gover", "**")

	var DiffFile = func(fileName string, info os.FileInfo, err error) error {
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

		if headModTime, ok := head.ModTimes[fileName]; ok {
			if modTime == headModTime {
					status[fileName] = "="
			} else {
				status[fileName] = "M"
			}
		} else {
			status[fileName] = "+"
		}

		return nil
	}

	// fmt.Printf("No changes detected in %s for commit %s\n", workDir, snapshot.ID)
	filepath.Walk(workingDirectory, DiffFile)

	if options.JsonMode {

	} else {
		for fileName, fileStatus := range status {
			if fileStatus == "=" && !options.VerboseMode {
				continue
			}
	
			fmt.Printf("%s %s\n", fileStatus, fileName)
		}
	}
}