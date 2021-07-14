package gover

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

	"github.com/bmatcuk/doublestar"
	"github.com/restic/chunker"
)

func CommitSnapshot(message string, filters []string, poly chunker.Pol, compressionLevel uint16, maxPackBytes int64) error {
	buf := make([]byte, 8*1024*1024) // reuse this buffer

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

	zipWriter := zip.NewWriter(packFile)

	if err != nil {
		if VerboseMode {
			fmt.Printf("Error creating zip writer for %s\n", packPath)
		}

		return err
	}

	var packBytesRemaining int64 = maxPackBytes

	var VersionFile = func(fileName string, info os.FileInfo, err error) error {
		fileName = strings.TrimSuffix(fileName, "\n")

		if info.IsDir() {
			return nil
		}

		matched, err := doublestar.PathMatch(goverDir, fileName)

		if matched {
			if VerboseMode {
				fmt.Printf("Skipping file %s in .gover\n", fileName)
			}

			return nil
		}

		for _, pattern := range filters {
			matched, err := doublestar.PathMatch(pattern, fileName)

			Check(err)
			if matched {
				if VerboseMode {
					fmt.Printf("Skipping file %s which matches with %s\n", fileName, pattern)
				}

				return nil
			}
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
			if VerboseMode {
				fmt.Printf("Skipping %s\n", fileName)
			}

			snap.FileChunkIds[fileName] = head.FileChunkIds[fileName]

			for _, chunkId := range snap.FileChunkIds[fileName] {
				snap.ChunkPackIds[chunkId] = head.ChunkPackIds[chunkId]
			}

			return nil
		}

		if VerboseMode {
			fmt.Printf("Chunking %s\n", fileName)
		} else {
			fmt.Println(fileName)
		}

		in, err := os.Open(fileName)

		if err != nil {
			if VerboseMode {
				fmt.Printf("Can't open file %s for reading, skipping\n", fileName)
			}

			return err
		}

		defer in.Close()
		mychunker := chunker.New(in, chunker.Pol(poly))

		if VerboseMode {
			fmt.Printf("Storing %s\n", fileName)
		} else {
			fmt.Println(fileName)
		}

		for fileBytesRemaining > 0 {
			chunk, err := mychunker.Next(buf)

			if err == nil {
				fileBytesRemaining -= int64(chunk.Length)
			} else if err == io.EOF {
				fileBytesRemaining = 0
			} else if err != nil {
				if VerboseMode {
					fmt.Printf("Error writing file %s to pack %s, aborting\n", fileName, packPath)
				}

				return err
			}

			chunkId := fmt.Sprintf("%064x", sha256.Sum256(chunk.Data))
			snap.FileChunkIds = append(snap.FileChunkIds, chunkId)

			if _, ok := snap.ChunkPackIds[chunkId]; ok {
				if VerboseMode {
					fmt.Printf("Skipping Chunk ID %s already in pack %s\n", chunkId[0:16], snap.ChunkPackIds[chunkId][0:16])
				}
			} else {
				if VerboseMode {
					fmt.Printf("Chunk %s: chunk size %d kB, total size %d kB\n", chunkId[0:16], chunk.Length/1024, (maxPackBytes-packBytesRemaining)/1024)
					// fmt.Printf("Compression level: %d\n", compressionLevel)
				}

				snap.ChunkPackIds[chunkId] = packId

				var header zip.FileHeader
				header.Name = chunkId
				header.Method = compressionLevel

				writer, err := zipWriter.CreateHeader(&header)

				if err != nil {
					if VerboseMode {
						fmt.Printf("Error creating zip file header for %s\n", packPath)
					}

					return err
				}

				writer.Write(chunk.Data)
				snap.ChunkPackIds[chunkId] = packId
				packBytesRemaining -= int64(chunk.Length)
			}

			if packBytesRemaining <= 0 {
				packFile.Close()
				zipWriter.Close()
				packId := RandHexString(PACK_ID_LEN)
				packFolderPath := path.Join(goverDir, "packs", packId[0:2])
				os.MkdirAll(packFolderPath, 0777)
				packPath := path.Join(packFolderPath, packId+".zip")

				if VerboseMode {
					fmt.Printf("Creating pack number: %3d, ID: %s\n", packCount, packId[0:16])
				}

				var err error
				packFile, err = os.Create(packPath)

				if err != nil {
					if VerboseMode {
						fmt.Printf("Error creating pack file for %s\n", packPath)
					}

					return err
				}

				zipWriter = zip.NewWriter(packFile)

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
	if err := filepath.Walk(workingDirectory, VersionFile); err != nil {
		fmt.Printf("Error committing:\n")
		fmt.Println(err)
	}

	packFile.Close()

	snapFolder := filepath.Join(".gover2", "snapshots")
	os.MkdirAll(snapFolder, 0777)
	snapFile := filepath.Join(snapFolder, ts+".json")
	snap.Write(snapFile)

	WriteHead(ts)
	return nil
}
