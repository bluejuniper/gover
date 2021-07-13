package snapshots

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/akbarnes/gover/src/options"
	"github.com/akbarnes/gover/src/util"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/restic/chunker"
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
		verFile := filepath.Join(verFolder, hash+ext)

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

			if options.VerboseMode {
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

func CommitChunkedSnapshot(message string, filters []string, mypoly chunker.Pol, compressionLevel uint16, maxPackBytes int64) error {
	t := time.Now()
	ts := t.Format("2006-01-02T15-04-05")
	snap := Snapshot{Time: ts}
	snap.Files = []string{}
	snap.StoredFiles = make(map[string]string)
	snap.ModTimes = make(map[string]string)
	snap.Message = message
	snap.PackIds = make(map[string][]string)
	snap.ChunkIds = make(map[string][]string)

	if err := os.MkdirAll(archiveFolder, 0777); err != nil {
		if VerboseMode {
			fmt.Printf("Error creating archive folder %s\n", archiveFolder)
		}

		return err
	}

	packCount := 1
	packPath := filepath.Join(archiveFolder, fmt.Sprintf("pack%d.dat", packCount))
	packFile, err := os.Create(packPath)
	var packOffset int64 = 0
	var packBytesRemaining int64 = maxPackBytes

	if err != nil {
		if VerboseMode {
			fmt.Printf("Error creating pack file %s\n", packPath)
		}

		return err
	}

	workingDirectory := "."
	head := ReadHead()

	goverDir := filepath.Join(workingDirectory, ".gover", "**")

	var VersionFile = func(fileName string, info os.FileInfo, err error) error {
		fileName = strings.TrimSuffix(fileName, "\n")

		if info.IsDir() {
			return nil
		}

		props, err := os.Stat(fileName)

		if err != nil {
			if VerboseMode {
				fmt.Printf("Can't stat file %s, skipping\n", fileName)
			}

			return err
		}

		in, err := os.Open(fileName)

		if err != nil {
			if VerboseMode {
				fmt.Printf("Can't open file %s for reading, skipping\n", fileName)
			}

			return err
		}

		defer in.Close()

		if VerboseMode {
			fmt.Printf("Storing %s\n", fileName)
		} else {
			fmt.Println(fileName)
		}

		fileBytesRemaining := props.Size()

		snap.Files = append(snap.Files, fileName)
		snap.PackNumbers[fileName] = []int{}
		snap.Offsets[fileName] = []int64{}
		snap.Lengths[fileName] = []int64{}

		for fileBytesRemaining > 0 {
			copyBytes := min64(packBytesRemaining, fileBytesRemaining)
			bytesCopied, err := io.CopyN(packFile, in, copyBytes)

			if err == nil {
				fileBytesRemaining -= bytesCopied
				packBytesRemaining -= bytesCopied
				snap.PackNumbers[fileName] = append(snap.PackNumbers[fileName], packCount)
				snap.Offsets[fileName] = append(snap.Offsets[fileName], packOffset)
				snap.Lengths[fileName] = append(snap.Lengths[fileName], bytesCopied)
				packOffset += bytesCopied
			} else {
				if VerboseMode {
					fmt.Printf("Error writing file %s to pack %s, aborting\n", fileName, packPath)
				}

				return err
			}

			if packBytesRemaining <= 0 {
				packFile.Close()
				packCount++
				packOffset = 0
				packBytesRemaining = maxPackBytes
				packPath = filepath.Join(archiveFolder, fmt.Sprintf("pack%d.dat", packCount))
				var err error
				packFile, err = os.Create(packPath)

				if err != nil {
					if VerboseMode {
						fmt.Printf("Error creating pack file %s\n", packPath)
					}

					return err
				}

				if VerboseMode {
					fmt.Printf("Creating new pack file %s\n", packPath)
				}
			}
		}

		return nil
	}

	// fmt.Printf("No changes detected in %s for commit %s\n", workDir, snapshot.ID)
	filepath.Walk(workingDirectory, VersionFile)
	packFile.Close()
	snap.Write(archiveFolder)
	return nil
}
