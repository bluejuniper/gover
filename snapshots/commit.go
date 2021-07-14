package snapshots

import (
	"archive/zip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/akbarnes/gover/src/options"
	"github.com/restic/chunker"
)

func CommitChunkedSnapshot(message string, filters []string, mypoly chunker.Pol, compressionLevel uint16, maxPackBytes int64) error {
	t := time.Now()
	ts := t.Format("2006-01-02T15-04-05")
	snap := Snapshot{Time: ts, Message: message}
	snap.ChunkPackIds = make(map[string]string)
	snap.FileChunkIds = make(map[string][]string)
	snap.FileModTimes = make(map[string]string)

	workingDirectory := "."
	head := ReadHead()

	goverDir := filepath.Join(workingDirectory, ".gover2")

	if err := os.MkdirAll(goverDir, 0777); err != nil {
		if VerboseMode {
			fmt.Printf("Error creating gover folder %s\n", goverDir)
		}

		return err
	}

	packCount := 0

	packId := RandHexString(PACK_ID_LEN)
	packFolderPath := path.Join(goverDir, "packs", packId[0:2])
	os.MkdirAll(packFolderPath, 0777)
	packPath := path.Join(packFolderPath, packId+".zip")
	packCount++

	if VerboseMode {
		fmt.Printf("Creating pack number: %3d, ID: %s\n", packCount, packId[0:16])
	}

	packFile, err := os.Create(packPath)

	if err != nil {
		panic(fmt.Sprintf("Error creating pack file %s", packPath))
	}

	defer packFile.Close()
	zipWriter := zip.NewWriter(packFile)

	if err != nil {
		if VerboseMode {
			fmt.Printf("Error creating zip writer for %s\n", packPath)
		}

		return err
	}

	defer zipWriter.Close()
	var packBytesRemaining int64 = maxPackBytes

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

		fileBytesRemaining := props.Size()
		modTime := props.ModTime().Format("2006-01-02T15-04-05")
		snap.FileModTimes[fileName] = modTime
		snap.FileChunkIds[fileName] = []string{}

		if headModTime, ok := head.FileModTimes[fileName]; ok && modTime == headModTime {
			if options.VerboseMode {
				fmt.Printf("Skipping %s\n", fileName)
			}

			snap.FileChunkIds[fileName] = head.FileChunkIds[fileName]

			for _, chunkId := range snap.FileChunkIds[fileName] {
				snap.ChunkPackIds[chunkId] = head.ChunkPackIds[chunkId]
			}

			return nil
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

		for fileBytesRemaining > 0 {
			chunk, err := mychunker.Next(buf)

			if err == nil {
				fileBytesRemaining -= chunk.Length
			} else if err == io.EOF {
				fileBytesRemaining = 0
			} else if err != nil {
				if VerboseMode {
					fmt.Printf("Error writing file %s to pack %s, aborting\n", fileName, packPath)
				}

				return err
			}

			chunkId := fmt.Sprintf("%064x", sha256.Sum256(chunk.Data))
			snap.ChunkPackIds[fileName] = append(snap.ChunkPackIds[fileName], chunkId)
			packBytesRemaining -= chunk.Length

			if packBytesRemaining <= 0 {
				packFile.Close()
				packId := RandHexString(PACK_ID_LEN)
				packFolderPath := path.Join(goverDir, "packs", packId[0:2])
				os.MkdirAll(packFolderPath, 0777)
				packPath := path.Join(packFolderPath, packId+".zip")

				if VerboseMode {
					fmt.Printf("Creating pack number: %3d, ID: %s\n", packCount, packId[0:16])
				}

				var err Error
				packFile, err = os.Create(packPath)

				if err != nil {
					panic(fmt.Sprintf("Error creating pack file %s", packPath))
				}

				defer packFile.Close()
				zipWriter := zip.NewWriter(packFile)

				if err != nil {
					if VerboseMode {
						fmt.Printf("Error creating zip writer for %s\n", packPath)
					}

					return err
				}

				defer zipWriter.Close()
				packCount++
				packBytesRemaining = maxPackBytes

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

	snapFolder := filepath.Join(".gover", "snapshots")
	os.MkdirAll(snapFolder, 0777)
	snapFile := filepath.Join(snapFolder, ts+".json")
	snap.Write(snapFile)

	WriteHead(ts)
	return nil
}
